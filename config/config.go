package config

import (
	"github.com/jinzhu/configor"
)

// Config - Application configuration
type Config struct {
	Log    string `yaml:"log" default:"" env:"LOG_PATH"`
	Debug  bool   `yaml:"debug" default:"false" env:"DEBUG"`
	MySQL struct {
		Host          string `yaml:"host" default:"localhost" env:"MYSQL_HOST"`
		User          string `yaml:"user" default:"root" env:"MYSQL_USER"`
		Password      string `yaml:"password" default:"" env:"MYSQL_PASSWORD"`
		Port          int    `yaml:"port" default:"3306" env:"MYSQL_PORT"`
		Database      string `yaml:"database" default:"" env:"MYSQL_DATABASE"`
		DSN           string `yaml:"dsn" default:"" env:"MYSQL_DSN"`
		ReadOnly      bool   `yaml:"read_only" default:"false" env:"MYSQL_READ_ONLY"`
		ExplainCheck  bool   `yaml:"explain_check" default:"false" env:"MYSQL_EXPLAIN_CHECK"`
	} `yaml:"mysql"`
}

// LoadConfig - Load configuration file
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.New(&configor.Config{
		Debug:      false,
		Verbose:    false,
		Silent:     true,
		AutoReload: false,
	}).Load(cfg, path)
	return cfg, err
}
