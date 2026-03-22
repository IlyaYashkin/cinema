package config

import (
	"cinema/internal/lib/config"
	"cinema/internal/lib/env"
)

type Config struct {
	Env        env.Env           `yaml:"env" env-default:"local"`
	GRPCConfig config.GRPCConfig `yaml:"grpc"`
	DBConfig   config.DBConfig   `yaml:"db"`
	S3Config   config.S3Config   `yaml:"s3"`
}
