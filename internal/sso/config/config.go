package config

import (
	"cinema/internal/lib/config"
	"time"
)

type Config struct {
	config.Config
	GRPCConfig           config.GRPCConfig `yaml:"grpc"`
	DBConfig             config.DBConfig   `yaml:"db"`
	SMTPConfig           config.SMTPConfig `yaml:"smtp"`
	JWTConfig            config.JWTConfig  `yaml:"jwt"`
	RedisConfig          RedisConfig       `yaml:"redis"`
	ResetTokenTTL        time.Duration     `yaml:"reset_token_ttl"`
	ResetPasswordBaseUrl string            `yaml:"reset_password_base_url"`
}

type RedisConfig struct {
	URL string `env:"REDIS_URL" env-default:"localhost:6379"`
}
