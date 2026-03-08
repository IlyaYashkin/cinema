package config

import (
	"cinema/internal/lib/config"
	"cinema/internal/lib/env"
	"time"
)

type Config struct {
	Env                  env.Env            `yaml:"env" env-default:"local"`
	GRPCConfig           config.GRPCConfig  `yaml:"grpc"`
	DBConfig             config.DBConfig    `yaml:"db"`
	SMTPConfig           config.SMTPConfig  `yaml:"smtp"`
	JWTConfig            config.JWTConfig   `yaml:"jwt"`
	RedisConfig          config.RedisConfig `yaml:"redis"`
	ResetTokenTTL        time.Duration      `yaml:"reset_token_ttl"`
	ResetPasswordBaseUrl string             `yaml:"reset_password_base_url"`
}
