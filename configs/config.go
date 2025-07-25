package configs

import (
	"KinopoiskTwoActors/configs/loader"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"
)

type KinopoiskConfig struct {
	Token string `validate:"required"`
	Path  string `validate:"required"`
}

type RedisConfig struct {
	Host         string        `validate:"required"`
	DB           int           `validate:"required"`
	User         string        `validate:"required"`
	Password     string        `validate:"required"`
	MaxRetries   int           `validate:"required"`
	DialTimeout  time.Duration `validate:"required"`
	ReadTimeout  time.Duration `validate:"required"`
	WriteTimeout time.Duration `validate:"required"`
}

type TelegramConfig struct {
	Token             string        `validate:"required"`
	ConnectionTimeout time.Duration `validate:"required"`
}

type Config struct {
	KP  KinopoiskConfig
	TG  TelegramConfig
	RD  RedisConfig
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
			ConnectionTimeout: getEnvAsDuration(envs["TELEGRAM_CONNECTION_TIMEOUT"], 5*time.Second),
		},
		RD: RedisConfig{
			Host:         envs["REDIS_HOST"],
			DB:           getEnvAsInt(envs["REDIS_DB"], 0),
			User:         envs["REDIS_USER"],
			Password:     envs["REDIS_PASSWORD"],
			MaxRetries:   getEnvAsInt(envs["REDIS_MAX_RETRIES"], 3),
			DialTimeout:  getEnvAsDuration(envs["REDIS_DIAL_TIMEOUT"], 5*time.Second),
			ReadTimeout:  getEnvAsDuration(envs["REDIS_READ_TIMEOUT"], 5*time.Second),
			WriteTimeout: getEnvAsDuration(envs["REDIS_WRITE_TIMEOUT"], 5*time.Second),
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
		log.Printf("%s:Invalid value for %s, using default: %v", op, strValue, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsInt(strValue string, defaultValue int) int {
	const op = "configs.getEnvAsInt"
	if strValue == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(strValue)
	if err != nil {
		log.Printf("%s:Invalid value for %s, using default: %v", op, strValue, defaultValue)
		return defaultValue
	}
	return value
}
