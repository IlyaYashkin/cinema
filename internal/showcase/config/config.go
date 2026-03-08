package config

import (
	"cinema/internal/lib/config"
	"cinema/internal/lib/env"
)

type Config struct {
	Env      env.Env         `yaml:"env" env-default:"local"`
	S3Config config.S3Config `yaml:"s3"`
}
