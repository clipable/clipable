// Package config handles loading settings
package config

import (
	"github.com/alexsasharegan/dotenv"
	"github.com/dustin/go-humanize"
	"github.com/kelseyhightower/envconfig"

	log "github.com/sirupsen/logrus"
)

// A Config holds all configurable settings from a yml config file
type Config struct {
	Debug              bool   `default:"false"`
	ListenAddr         string `default:"0.0.0.0:8080"`
	MetricsListenAddr  string `default:"127.0.0.1:9991"`
	MaxUploadSize      string `default:"5 GB" split_words:"true"`
	MaxUploadSizeBytes int64  `ignored:"true"` // This is set by the parser to the byte value of MaxUploadSize

	FFmpeg struct {
		Concurrency    int      `default:"1"`
		Threads        int      `default:"0"`
		Preset         string   `default:"medium"` // https://trac.ffmpeg.org/wiki/Encode/H.264#:~:text=preset%20and%20tune-,Preset,-A%20preset%20is
		Tune           string   `default:"film"`   // https://trac.ffmpeg.org/wiki/Encode/H.264#:~:text=x264%20%2D%2Dfullhelp.-,Tune,-You%20can%20optionally
		QualityPresets []string `split_words:"true" default:"640x360-30@1,854x480-30@2.5,1280x720-30@5,1920x1080-30@8,1920x1080-60@12,2560x1440-30@16,2560x1440-60@24,3840x2160-30@45,3840x2160-60@68,7680x4320-30@160,7680x4320-60@240"`
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

	// Parse the human readable max upload size into bytes
	maxUploadSizeBytes, err := humanize.ParseBytes(cfg.MaxUploadSize)
	if err != nil {
		return nil, err
	}

	cfg.MaxUploadSizeBytes = int64(maxUploadSizeBytes)

	return cfg, nil
}
