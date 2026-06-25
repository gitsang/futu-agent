package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("Connected to database")
	return db, nil
}

func Migrate(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS trade_decisions (
			id SERIAL PRIMARY KEY,
			agent_id VARCHAR(50) NOT NULL,
			stock_code VARCHAR(20) NOT NULL,
			market VARCHAR(10) NOT NULL,
			action VARCHAR(10) NOT NULL,
			quantity INTEGER,
			price DECIMAL(15, 4),
			reason TEXT,
			llm_response JSONB,
			executed BOOLEAN DEFAULT FALSE,
			executed_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS positions (
			id SERIAL PRIMARY KEY,
			stock_code VARCHAR(20) NOT NULL,
			market VARCHAR(10) NOT NULL,
			quantity INTEGER NOT NULL,
			avg_cost DECIMAL(15, 4) NOT NULL,
			current_price DECIMAL(15, 4),
			unrealized_pnl DECIMAL(15, 4),
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(stock_code, market)
		)`,
		`CREATE TABLE IF NOT EXISTS account_funds (
			id SERIAL PRIMARY KEY,
			market VARCHAR(10) NOT NULL DEFAULT 'ALL',
			currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
			total_assets DECIMAL(15, 4),
			cash DECIMAL(15, 4),
			market_value DECIMAL(15, 4),
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(market)
		)`,
		`CREATE TABLE IF NOT EXISTS agent_configs (
			id SERIAL PRIMARY KEY,
			agent_id VARCHAR(50) NOT NULL,
			market VARCHAR(10) NOT NULL DEFAULT 'CN',
			name VARCHAR(100),
			description TEXT,
			llm_model VARCHAR(100),
			llm_endpoint VARCHAR(255),
			trading_strategy TEXT,
			risk_parameters JSONB,
			enabled BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(agent_id, market)
		)`,
		`CREATE TABLE IF NOT EXISTS system_configs (
			key VARCHAR(100) PRIMARY KEY,
			value TEXT NOT NULL,
			description TEXT,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	alterMigrations := []string{
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='account_funds' AND column_name='market') THEN
				ALTER TABLE account_funds ADD COLUMN market VARCHAR(10) NOT NULL DEFAULT 'ALL';
			END IF;
		END $$`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='account_funds' AND column_name='currency') THEN
				ALTER TABLE account_funds ADD COLUMN currency VARCHAR(10) NOT NULL DEFAULT 'CNY';
			END IF;
		END $$`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='agent_configs' AND column_name='market') THEN
				ALTER TABLE agent_configs ADD COLUMN market VARCHAR(10) NOT NULL DEFAULT 'CN';
			END IF;
		END $$`,
	}

	for _, migration := range alterMigrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute alter migration: %w", err)
		}
	}

	indexMigrations := []string{
		`CREATE INDEX IF NOT EXISTS idx_trade_decisions_agent_id ON trade_decisions(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_trade_decisions_market ON trade_decisions(market)`,
		`CREATE INDEX IF NOT EXISTS idx_trade_decisions_created_at ON trade_decisions(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_trade_decisions_stock_code ON trade_decisions(stock_code)`,
		`CREATE INDEX IF NOT EXISTS idx_trade_decisions_executed ON trade_decisions(executed)`,
	}

	for _, migration := range indexMigrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute index migration: %w", err)
		}
	}

	log.Println("Database migrations completed")
	return nil
}
