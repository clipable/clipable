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

	FFmpeg struct {
		Concurrency int    `default:"1"`
		Preset      string `default:"medium"` // https://trac.ffmpeg.org/wiki/Encode/H.264#:~:text=preset%20and%20tune-,Preset,-A%20preset%20is
		Tune        string `default:"film"`   // https://trac.ffmpeg.org/wiki/Encode/H.264#:~:text=x264%20%2D%2Dfullhelp.-,Tune,-You%20can%20optionally
	}

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

		IDHashKey string
	}
}

// New uses a file path stored in the CONFIG environment variable to populate the Config struct
func New() (*Config, error) {
	cfg := &Config{}

	if err := dotenv.Load(); err != nil {
		log.Warningln("Couldn't find .env, using existing environment variables")
	}

	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
