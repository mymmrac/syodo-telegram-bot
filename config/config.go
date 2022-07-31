package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/mymmrac/syodo-telegram-bot/logger"
)

// LoadConfig loads config from config file and environment variables
func LoadConfig(filename string) (*Config, error) {
	cfg := &Config{}

	_, err := toml.DecodeFile(filename, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	var ok bool

	const botTokenEnv = "BOT_TOKEN"
	cfg.Settings.BotToken, ok = os.LookupEnv(botTokenEnv)
	if !ok {
		return nil, fmt.Errorf("no %q environment variable", botTokenEnv)
	}

	const providerTokenEnv = "PROVIDER_TOKEN"
	cfg.Settings.ProviderToken, ok = os.LookupEnv(providerTokenEnv)
	if !ok {
		return nil, fmt.Errorf("no %q environment variable", providerTokenEnv)
	}

	return cfg, nil
}

// Config represents general config structure
type Config struct {
	Log      Log
	Settings Settings
}

// Log represents logger config
type Log struct {
	Level       string
	Destination string
	Filename    string
}

// Settings represents general settings
type Settings struct {
	BotToken      string
	ProviderToken string
	StopTimeout   time.Duration
}

const (
	logDestinationStdout = "stdout"
	logDestinationStderr = "stderr"
	logDestinationFile   = "file"
)

const (
	logLevelError = "error"
	logLevelWarn  = "warn"
	logLevelInfo  = "info"
	logLevelDebug = "debug"
)

// ConfigureLogger apples config to logger
func (c *Config) ConfigureLogger(log *logger.Log) error {
	switch c.Log.Destination {
	case logDestinationStdout:
		log.SetOutput(os.Stdout)
	case logDestinationStderr:
		log.SetOutput(os.Stderr)
	case logDestinationFile:
		if err := log.SetOutputFile(c.Log.Filename); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown logger destination: %q", c.Log.Destination)
	}

	switch c.Log.Level {
	case logLevelError, logLevelWarn, logLevelInfo, logLevelDebug:
		log.SetLevel(c.Log.Level)
	default:
		return fmt.Errorf("unknown logger level: %q", c.Log.Level)
	}

	return nil
}
