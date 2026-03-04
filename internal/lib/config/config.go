package config

import "time"

type JWTConfig struct {
	PrivateKeyPath string        `env:"JWT_PRIVATE_KEY_PATH" env-required:"true"`
	PublicKeyPath  string        `env:"JWT_PUBLIC_KEY_PATH" env-required:"true"`
	AccessTTL      time.Duration `yaml:"access_ttl" env-default:"15m"`
	RefreshTTL     time.Duration `yaml:"refresh_ttl" env-default:"720h"`
}
