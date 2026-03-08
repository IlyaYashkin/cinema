package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `yaml:"env" env-default:"local"`
}

type JWTConfig struct {
	PrivateKeyPath string        `env:"JWT_PRIVATE_KEY_PATH" env-required:"true"`
	PublicKeyPath  string        `env:"JWT_PUBLIC_KEY_PATH" env-required:"true"`
	AccessTTL      time.Duration `yaml:"access_ttl" env-default:"15m"`
	RefreshTTL     time.Duration `yaml:"refresh_ttl" env-default:"720h"`
}

type DBConfig struct {
	DSN             string        `env:"DATABASE_DSN" env-required:"true"`
	MaxConns        int32         `yaml:"max_conns" env-default:"10"`
	MinConns        int32         `yaml:"min_conns" env-default:"2"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime" env-default:"1h"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time" env-default:"30m"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type SMTPConfig struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-default:"587"`
	Username string `env:"SMTP_USERNAME" env-required:"true"`
	Password string `env:"SMTP_PASSWORD" env-required:"true"`
	From     string `yaml:"from" env-required:"true"`
}

type RedisConfig struct {
	URL string `env:"REDIS_URL" env-default:"localhost:6379"`
}

type StorageConfig struct {
	Endpoint  string `yaml:"endpoint" env-required:"true"`
	Region    string `yaml:"region" env-required:"true"`
	Bucket    string `yaml:"bucket" env-required:"true"`
	AccessKey string `env:"STORAGE_ACCESS_KEY" env-required:"true"`
	SecretKey string `env:"STORAGE_SECRET_KEY" env-required:"true"`
}

func MustLoad[T any]() *T {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg T

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		//panic("error reading env file: " + err.Error())
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("error reading config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
