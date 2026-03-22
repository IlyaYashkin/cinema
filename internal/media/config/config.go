package config

import "cinema/internal/lib/env"

type Config struct {
	Env env.Env `yaml:"env" env-default:"local"`
}
