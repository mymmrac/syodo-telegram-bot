package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"

	"github.com/mymmrac/syodo-telegram-bot/logger"
)

//nolint:gosec
const (
	botTokenEnv         = "BOT_TOKEN"
	providerTokenEnv    = "PROVIDER_TOKEN"
	googleMapsAPIKeyEnv = "GOOGLE_MAPS_API_KEY"
	syodoAPIKeyEnv      = "SYODO_API_KEY"
	liqPayPrivetKeyEnv  = "LIQ_PAY_PRIVET_KEY"
)

// LoadConfig loads config from config file and environment variables
func LoadConfig(filename string) (*Config, error) {
	cfg := &Config{}

	_, err := toml.DecodeFile(filename, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	var ok bool

	cfg.App.BotToken, ok = os.LookupEnv(botTokenEnv)
	if !ok {
		return nil, fmt.Errorf("no %q environment variable", botTokenEnv)
	}

	cfg.App.ProviderToken, ok = os.LookupEnv(providerTokenEnv)
	if !ok {
		return nil, fmt.Errorf("no %q environment variable", providerTokenEnv)
	}

	cfg.App.LiqPayPrivetKeyEnv, ok = os.LookupEnv(liqPayPrivetKeyEnv)
	if !ok {
		return nil, fmt.Errorf("no %q environment variable", liqPayPrivetKeyEnv)
	}

	cfg.App.GoogleMapsAPIKey, ok = os.LookupEnv(googleMapsAPIKeyEnv)
	if !ok {
		return nil, fmt.Errorf("no %q environment variable", googleMapsAPIKeyEnv)
	}

	cfg.App.SyodoAPIKey, ok = os.LookupEnv(syodoAPIKeyEnv)
	if !ok {
		return nil, fmt.Errorf("no %q environment variable", syodoAPIKeyEnv)
	}

	validate := validator.New()
	if err = validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

// Config represents general config structure
type Config struct {
	Log      Log
	Settings Settings
	App      App
}

// Log represents logger config
type Log struct {
	Level       string `validate:"required,oneof=error warn info debug"`
	Destination string `validate:"required,oneof=stdout stderr file"`
	Filename    string `validate:"required_if=Destination file"`
}

// Settings represents general settings
type Settings struct {
	StopTimeout        time.Duration `validate:"gte=0"`
	UseLongPulling     bool          `validate:"-"`
	ServerHost         string        `validate:"hostname_port"`
	WebhookURL         string        `validate:"url"`
	LongPullingTimeout int           `validate:"gte=0"`
	RequestTimeout     time.Duration `validate:"gt=0"`
	TestMode           bool          `validate:"-"`
}

// App represents business logic settings
type App struct {
	BotToken           string `validate:"required"`
	ProviderToken      string `validate:"required"`
	LiqPayPrivetKeyEnv string `validate:"required"`
	GoogleMapsAPIKey   string `validate:"required"`
	SyodoAPIKey        string `validate:"required"`
	WebAppURL          string `validate:"url"`
	SyodoAPIURL        string `validate:"url"`
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
