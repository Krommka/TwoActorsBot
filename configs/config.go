package configs

import (
	"KinopoiskTwoActors/configs/loader"
	"flag"
	"fmt"
	"log"
	"time"
)

type KinopoiskConfig struct {
	Token string `validate:"required"`
	Path  string `validate:"required"`
}

type TelegramConfig struct {
	Token             string        `validate:"required"`
	ConnectionTimeout time.Duration `validate:"required"`
}

type Config struct {
	KP  KinopoiskConfig
	TG  TelegramConfig
	Env string
}

func MustLoad(loader loader.ConfigLoader) *Config {
	env := flag.String("env", "dev", "Environment type")
	flag.Parse()

	const op = "configs.MustLoad"
	envs, err := loader.Load()
	if err != nil {
		log.Fatalf("%s: config load failed: %+v", op, err)
	}
	cfg := &Config{
		KP: KinopoiskConfig{
			Token: envs["KINOPOISK_TOKEN"],
			Path:  envs["KINOPOISK_PATH"],
		},
		TG: TelegramConfig{
			Token:             envs["TELEGRAM_TOKEN"],
			ConnectionTimeout: getEnvAsDuration(envs["TELEGRAM_CONNECTION_TIMEOUT"], 10*time.Second),
		},
		Env: *env,
	}

	if err := validateConfig(cfg); err != nil {
		log.Fatalf("%s: config validation failed: %+v", op, err)
	}

	return cfg
}

func validateConfig(cfg *Config) error {
	if cfg.KP.Token == "" || cfg.TG.Token == "" {
		return fmt.Errorf("missing required configuration")
	}
	return nil
}

func getEnvAsDuration(strValue string, defaultValue time.Duration) time.Duration {
	const op = "configs.getEnvAsDuration"
	if strValue == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(strValue)
	if err != nil {
		log.Printf("Invalid value for %s, using default: %v", strValue, defaultValue)
		return defaultValue
	}
	return value
}
