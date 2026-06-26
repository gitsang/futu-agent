package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gitsang/configer"
	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port int `yaml:"port" mapstructure:"port" default:"9080" json:"port"`
}

type FutuConfig struct {
	Host        string `yaml:"host" mapstructure:"host" default:"localhost" json:"host"`
	Port        int    `yaml:"port" mapstructure:"port" default:"11111" json:"port"`
	Account     string `yaml:"account" mapstructure:"account" json:"account"`
	PasswordMD5 string `yaml:"password_md5" mapstructure:"password_md5" json:"-"`
}

type LLMConfig struct {
	BaseURL string `yaml:"base_url" mapstructure:"base_url" default:"https://api.openai.com/v1" json:"base_url"`
	Model   string `yaml:"model" mapstructure:"model" default:"gpt-4" json:"model"`
	APIKey  string `yaml:"api_key" mapstructure:"api_key" json:"-"`
}

type ProxyConfig struct {
	HTTP  string `yaml:"http" mapstructure:"http" json:"http"`
	HTTPS string `yaml:"https" mapstructure:"https" json:"https"`
}

type TradingConfig struct {
	Enabled bool `yaml:"enabled" mapstructure:"enabled" default:"false" json:"enabled"`
}

type RuleTemplate struct {
	AggressionLevel       string  `yaml:"aggression_level" mapstructure:"aggression_level" json:"aggression_level"`
	BuyOnDipThreshold     float64 `yaml:"buy_on_dip_threshold" mapstructure:"buy_on_dip_threshold" json:"buy_on_dip_threshold"`
	TakeProfitThreshold   float64 `yaml:"take_profit_threshold" mapstructure:"take_profit_threshold" json:"take_profit_threshold"`
	StopLossThreshold     float64 `yaml:"stop_loss_threshold" mapstructure:"stop_loss_threshold" json:"stop_loss_threshold"`
	CashUsageMin          int     `yaml:"cash_usage_min" mapstructure:"cash_usage_min" json:"cash_usage_min"`
	CashUsageMax          int     `yaml:"cash_usage_max" mapstructure:"cash_usage_max" json:"cash_usage_max"`
	MaxCashRatio          int     `yaml:"max_cash_ratio" mapstructure:"max_cash_ratio" json:"max_cash_ratio"`
	PositionLossThreshold float64 `yaml:"position_loss_threshold" mapstructure:"position_loss_threshold" json:"position_loss_threshold"`
	LotSize               int     `yaml:"lot_size" mapstructure:"lot_size" json:"lot_size"`
	LotSizeRule           string  `yaml:"lot_size_rule" mapstructure:"lot_size_rule" json:"lot_size_rule"`
}

type AgentRules = RuleTemplate

type StockUniverseTemplate struct {
	ID           string                    `yaml:"id" mapstructure:"id" json:"id"`
	Market       string                    `yaml:"market" mapstructure:"market" json:"market"`
	Name         string                    `yaml:"name" mapstructure:"name" json:"name"`
	Schedule     string                    `yaml:"schedule" mapstructure:"schedule" json:"schedule"`
	Source       string                    `yaml:"source" mapstructure:"source" json:"source"`
	ScreenConfig StockUniverseScreenConfig `yaml:"screen_config" mapstructure:"screen_config" json:"screen_config"`
	Watchlist    []string                  `yaml:"watchlist" mapstructure:"watchlist" json:"watchlist"`
}

type StockUniverseConfig = StockUniverseTemplate

type StockUniverseScreenConfig struct {
	Market  string                      `yaml:"market" mapstructure:"market" json:"market"`
	Filters []StockUniverseFilterConfig `yaml:"filters" mapstructure:"filters" json:"filters"`
	Sort    []StockUniverseSortConfig   `yaml:"sort" mapstructure:"sort" json:"sort"`
	Limit   int                         `yaml:"limit" mapstructure:"limit" json:"limit"`
}

type StockUniverseFilterConfig struct {
	Field    string  `yaml:"field" mapstructure:"field" json:"field"`
	Operator string  `yaml:"operator" mapstructure:"operator" json:"operator"`
	Value    float64 `yaml:"value" mapstructure:"value" json:"value"`
	Unit     string  `yaml:"unit" mapstructure:"unit" json:"unit"`
}

type StockUniverseSortConfig struct {
	Field     string `yaml:"field" mapstructure:"field" json:"field"`
	Direction string `yaml:"direction" mapstructure:"direction" json:"direction"`
}

type AgentConfig struct {
	ID               string              `yaml:"id" mapstructure:"id" json:"id"`
	Market           string              `yaml:"market" mapstructure:"market" json:"market"`
	Name             string              `yaml:"name" mapstructure:"name" json:"name"`
	Description      string              `yaml:"description" mapstructure:"description" json:"description"`
	LLMModel         string              `yaml:"llm_model" mapstructure:"llm_model" json:"llm_model"`
	Enabled          bool                `yaml:"enabled" mapstructure:"enabled" json:"enabled"`
	StockUniverseRef string              `yaml:"stock_universe_ref" mapstructure:"stock_universe_ref" json:"stock_universe_ref"`
	RuleRef          string              `yaml:"rule_ref" mapstructure:"rule_ref" json:"rule_ref"`
	LotSize          int                 `yaml:"lot_size" mapstructure:"lot_size" json:"lot_size"`
	LotSizeRule      string              `yaml:"lot_size_rule" mapstructure:"lot_size_rule" json:"lot_size_rule"`
	TradingStrategy  string              `yaml:"trading_strategy" mapstructure:"trading_strategy" json:"trading_strategy"`
	Rules            RuleTemplate        `yaml:"-" mapstructure:"-" json:"rules"`
	StockUniverse    StockUniverseConfig `yaml:"-" mapstructure:"-" json:"stock_universe"`
}

type ResolvedAgentConfig struct {
	AgentConfig
	StockUniverse StockUniverseConfig `json:"stock_universe"`
	Rules         RuleTemplate        `json:"rules"`
}

type Config struct {
	Server         ServerConfig            `yaml:"server" mapstructure:"server" json:"server"`
	Futu           FutuConfig              `yaml:"futu" mapstructure:"futu" json:"futu"`
	LLM            LLMConfig               `yaml:"llm" mapstructure:"llm" json:"llm"`
	Proxy          ProxyConfig             `yaml:"proxy" mapstructure:"proxy" json:"proxy"`
	Trading        TradingConfig           `yaml:"trading" mapstructure:"trading" json:"trading"`
	StockUniverses []StockUniverseTemplate `yaml:"stock_universes" mapstructure:"stock_universes" json:"stock_universes"`
	RuleTemplates  map[string]RuleTemplate `yaml:"rule_templates" mapstructure:"rule_templates" json:"rule_templates"`
	Agents         []AgentConfig           `yaml:"agents" mapstructure:"agents" json:"agents"`
	Log            map[string]string       `yaml:"log" mapstructure:"log" json:"log"`
	ResolvedAgents []ResolvedAgentConfig   `yaml:"-" mapstructure:"-" json:"-"`
	ServerPort     int                     `yaml:"-" mapstructure:"-" json:"-"`
	FutuOpendHost  string                  `yaml:"-" mapstructure:"-" json:"-"`
	FutuOpendPort  int                     `yaml:"-" mapstructure:"-" json:"-"`
	LLMBaseURL     string                  `yaml:"-" mapstructure:"-" json:"-"`
	LLMModel       string                  `yaml:"-" mapstructure:"-" json:"-"`
	LLMAPIKey      string                  `yaml:"-" mapstructure:"-" json:"-"`
	HTTPProxy      string                  `yaml:"-" mapstructure:"-" json:"-"`
	TradingEnabled bool                    `yaml:"-" mapstructure:"-" json:"-"`
	ConfigPath     string                  `yaml:"-" mapstructure:"-" json:"-"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	cfger := configer.New(
		configer.WithTemplate(new(Config)),
		configer.WithEnvBind(
			configer.WithEnvPrefix("FUTU_AGENT"),
			configer.WithEnvDelim("_"),
		),
	)

	configPath := resolveConfigPath()
	if err := cfger.Load(cfg, configPath); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	cfg.ConfigPath = configPath
	if err := cfg.ResolveAgents(); err != nil {
		return nil, err
	}
	cfg.syncLegacyFields()
	log.Printf("Loaded %d agents from config", len(cfg.Agents))
	return cfg, nil
}

func (c *Config) ResolveAgents() error {
	stockUniverses := make(map[string]StockUniverseTemplate, len(c.StockUniverses))
	for _, universe := range c.StockUniverses {
		stockUniverses[universe.ID] = universe
	}

	resolved := make([]ResolvedAgentConfig, 0, len(c.Agents))
	for i := range c.Agents {
		agent := c.Agents[i]
		universe, ok := stockUniverses[agent.StockUniverseRef]
		if !ok {
			return fmt.Errorf("agent %s references unknown stock universe %s", agent.ID, agent.StockUniverseRef)
		}
		rules, ok := c.RuleTemplates[agent.RuleRef]
		if !ok {
			return fmt.Errorf("agent %s references unknown rule template %s", agent.ID, agent.RuleRef)
		}

		rules.LotSize = agent.LotSize
		rules.LotSizeRule = agent.LotSizeRule
		if universe.ScreenConfig.Market == "" {
			universe.ScreenConfig.Market = universe.Market
		}

		agent.Rules = rules
		agent.StockUniverse = universe
		c.Agents[i].Rules = rules
		c.Agents[i].StockUniverse = universe
		resolved = append(resolved, ResolvedAgentConfig{
			AgentConfig:   agent,
			StockUniverse: universe,
			Rules:         rules,
		})
	}
	c.ResolvedAgents = resolved
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

func (c *Config) GetResolvedAgents() []ResolvedAgentConfig {
	result := make([]ResolvedAgentConfig, len(c.ResolvedAgents))
	copy(result, c.ResolvedAgents)
	return result
}

func (c *Config) UpdateAgent(id string, enabled bool) bool {
	for i := range c.Agents {
		if c.Agents[i].ID == id {
			c.Agents[i].Enabled = enabled
			if err := c.ResolveAgents(); err != nil {
				log.Printf("Failed to resolve agents config: %v", err)
				return false
			}
			if err := c.save(); err != nil {
				log.Printf("Failed to save config: %v", err)
				return false
			}
			return true
		}
	}
	return false
}

func (c *Config) save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(c.ConfigPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

func (c *Config) syncLegacyFields() {
	c.ServerPort = c.Server.Port
	c.FutuOpendHost = c.Futu.Host
	c.FutuOpendPort = c.Futu.Port
	c.LLMBaseURL = c.LLM.BaseURL
	c.LLMModel = c.LLM.Model
	c.LLMAPIKey = c.LLM.APIKey
	c.HTTPProxy = c.Proxy.HTTP
	c.TradingEnabled = c.Trading.Enabled
}

func resolveConfigPath() string {
	if path := os.Getenv("FUTU_AGENT_CONFIG_PATH"); path != "" {
		return path
	}
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	if dir := os.Getenv("FUTU_AGENT_CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "config.yaml")
	}
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "config.yaml")
	}

	paths := []string{"config/config.yaml", "../config/config.yaml", "/app/config/config.yaml"}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "config/config.yaml"
}
