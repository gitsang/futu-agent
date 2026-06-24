package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/gitsang/futu-agent/backend/internal/config"
	"github.com/gitsang/futu-agent/backend/internal/database"
	"github.com/gitsang/futu-agent/backend/internal/handlers"
	"github.com/gitsang/futu-agent/backend/internal/services/agent"
	"github.com/gitsang/futu-agent/backend/internal/services/futu"
	"github.com/gitsang/futu-agent/backend/internal/services/llm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	futuClient, err := futu.NewClient(cfg.FutuOpendHost, cfg.FutuOpendPort)
	if err != nil {
		log.Fatalf("Failed to create Futu client: %v", err)
	}
	defer futuClient.Close()

	llmClient := llm.NewClient(cfg.LLMBaseURL, cfg.LLMModel, cfg.LLMAPIKey, cfg.HTTPProxy)
	agentEngine := agent.NewEngine(db, futuClient, llmClient, cfg)

	handler := handlers.NewHandler(db, agentEngine, futuClient)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Get("/account/funds", handler.GetAccountFunds)
		r.Get("/account/funds/all", handler.GetAllAccountFunds)
		r.Get("/account/positions", handler.GetPositions)

		r.Get("/decisions", handler.GetDecisions)
		r.Get("/decisions/{id}", handler.GetDecision)

		r.Get("/agents", handler.GetAgents)
		r.Post("/agents", handler.CreateAgent)
		r.Put("/agents/{id}", handler.UpdateAgent)
		r.Delete("/agents/{id}", handler.DeleteAgent)
		r.Post("/agents/{id}/start", handler.StartAgent)
		r.Post("/agents/{id}/stop", handler.StopAgent)

		r.Get("/config", handler.GetConfig)
		r.Put("/config", handler.UpdateConfig)

		r.Get("/status", handler.GetStatus)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := agentEngine.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start agent engine: %v", err)
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ServerPort),
		Handler: r,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Starting server on port %d", cfg.ServerPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}
