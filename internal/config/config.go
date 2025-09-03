package config

import (
	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"`
	PostgresDB       string `env:"POSTGRES_DB"`
	PostgresPort     string `env:"POSTGRES_PORT"`
	PostgresHost     string `env:"POSTGRES_HOST"`
}

func LoadConfig() (*Config, error) {
	var config Config
	if err := env.Parse(&config); err != nil {
		return nil, err
	}
	return &config, nil
}
