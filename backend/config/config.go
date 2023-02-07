// Package config handles loading settings
package config

import (
	"github.com/alexsasharegan/dotenv"
	"github.com/kelseyhightower/envconfig"

	log "github.com/sirupsen/logrus"
)

// A Config holds all configurable settings from a yml config file
type Config struct {
	Debug             bool   `default:"false"`
	ListenAddr        string `default:"0.0.0.0:8080"`
	MetricsListenAddr string `default:"127.0.0.1:9991"`

	CORS struct {
		Origin  string
		Enabled bool
	}

	Cookie struct {
		Key    string
		Domain string
	}

	S3 struct {
		Address string
		Access  string
		Secret  string
		Bucket  string
	}

	DB struct {
		Name     string
		Host     string
		Port     string
		User     string
		Password string
	}
}

// New uses a file path stored in the CONFIG environment variable to populate the Config struct
func New() (*Config, error) {
	cfg := &Config{}

	if err := dotenv.Load(); err != nil {
		log.Warningln("Failed to load .env, using existing environment variables")
	}

	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
