package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/mymmrac/syodo-telegram-bot/logger"
)

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

type Config struct {
	Log      Log
	Settings Settings
}

type Log struct {
	Level       string
	Destination string
	Filename    string
}

type Settings struct {
	BotToken      string
	ProviderToken string
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
	LogLevelDebug = "debug"
)

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
		return fmt.Errorf("unkown logger destination: %q", c.Log.Destination)
	}

	switch c.Log.Level {
	case logLevelError, logLevelWarn, logLevelInfo, LogLevelDebug:
		log.SetLevel(c.Log.Level)
	default:
		return fmt.Errorf("unkown logger level: %q", c.Log.Level)
	}

	return nil
}
