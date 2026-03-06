package config

import (
	"cinema/internal/lib/config"
)

type Config struct {
	config.Config
	GRPCConfig  config.GRPCConfig `yaml:"grpc"`
	DBConfig    config.DBConfig   `yaml:"db"`
	RedisConfig RedisConfig       `yaml:"redis"`
	JWTConfig   config.JWTConfig  `yaml:"jwt"`
}

type RedisConfig struct {
	URL string `env:"REDIS_URL" env-default:"localhost:6379"`
}
