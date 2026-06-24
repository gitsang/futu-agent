package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	ID              string `yaml:"id" json:"id"`
	Market          string `yaml:"market" json:"market"`
	Name            string `yaml:"name" json:"name"`
	Description     string `yaml:"description" json:"description"`
	LLMModel        string `yaml:"llm_model" json:"llm_model"`
	TradingStrategy string `yaml:"trading_strategy" json:"trading_strategy"`
	Enabled         bool   `yaml:"enabled" json:"enabled"`
}

type AgentsConfig struct {
	Agents []AgentConfig `yaml:"agents"`
}

type Config struct {
	ServerPort     int
	FutuOpendHost  string
	FutuOpendPort  int
	LLMBaseURL     string
	LLMModel       string
	LLMAPIKey      string
	HTTPProxy      string
	TradingEnabled bool
	ConfigDir      string
	Agents         []AgentConfig
}

func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnvAsInt("SERVER_PORT", 9080),
		FutuOpendHost:  getEnv("FUTU_OPEND_HOST", "localhost"),
		FutuOpendPort:  getEnvAsInt("FUTU_OPEND_PORT", 11111),
		LLMBaseURL:     getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
		LLMModel:       getEnv("LLM_MODEL", "gpt-4"),
		LLMAPIKey:      getEnv("LLM_API_KEY", ""),
		HTTPProxy:      getEnv("HTTP_PROXY", ""),
		TradingEnabled: getEnvAsBool("TRADING_ENABLED", false),
		ConfigDir:      getEnv("CONFIG_DIR", "./config"),
	}

	if err := cfg.loadAgents(); err != nil {
		log.Printf("Warning: Failed to load agents config: %v", err)
	}

	return cfg, nil
}

func (c *Config) loadAgents() error {
	configPath := fmt.Sprintf("%s/agents.yaml", c.ConfigDir)
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Agents config file not found at %s, using defaults", configPath)
			c.Agents = []AgentConfig{
				{
					ID:              "cn-agent-01",
					Market:          "CN",
					Name:            "A股交易代理",
					Description:     "专注于A股市场的自动交易代理",
					LLMModel:        c.LLMModel,
					TradingStrategy: "基于技术分析的A股交易策略",
					Enabled:         true,
				},
			}
			return nil
		}
		return fmt.Errorf("failed to read agents config: %w", err)
	}

	var agentsConfig AgentsConfig
	if err := yaml.Unmarshal(data, &agentsConfig); err != nil {
		return fmt.Errorf("failed to parse agents config: %w", err)
	}

	c.Agents = agentsConfig.Agents
	log.Printf("Loaded %d agents from config", len(c.Agents))
	return nil
}

func (c *Config) GetAgent(id string) *AgentConfig {
	for i := range c.Agents {
		if c.Agents[i].ID == id {
			return &c.Agents[i]
		}
	}
	return nil
}

func (c *Config) GetAgentsByMarket(market string) []AgentConfig {
	var result []AgentConfig
	for _, agent := range c.Agents {
		if market == "" || market == "ALL" || agent.Market == market {
			result = append(result, agent)
		}
	}
	return result
}

func (c *Config) UpdateAgent(id string, enabled bool) bool {
	for i := range c.Agents {
		if c.Agents[i].ID == id {
			c.Agents[i].Enabled = enabled
			if err := c.saveAgents(); err != nil {
				log.Printf("Failed to save agents config: %v", err)
				return false
			}
			return true
		}
	}
	return false
}

func (c *Config) saveAgents() error {
	configPath := fmt.Sprintf("%s/agents.yaml", c.ConfigDir)
	
	agentsConfig := AgentsConfig{Agents: c.Agents}
	data, err := yaml.Marshal(&agentsConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal agents config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write agents config: %w", err)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(strings.TrimSpace(value)); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
