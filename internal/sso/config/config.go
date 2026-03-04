package config

import (
	"cinema/internal/lib/config"
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string           `yaml:"env" env-default:"local"`
	GRPCConfig  GRPCConfig       `yaml:"grpc"`
	DBConfig    DBConfig         `yaml:"db"`
	RedisConfig RedisConfig      `yaml:"redis"`
	JWTConfig   config.JWTConfig `yaml:"jwt"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type DBConfig struct {
	DSN             string        `env:"DATABASE_DSN" env-required:"true"`
	MaxConns        int32         `yaml:"max_conns" env-default:"10"`
	MinConns        int32         `yaml:"min_conns" env-default:"2"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime" env-default:"1h"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time" env-default:"30m"`
}

type RedisConfig struct {
	URL string `env:"REDIS_URL" env-default:"localhost:6379"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("error reading env file: " + err.Error())
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config path is empty: " + err.Error())
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
