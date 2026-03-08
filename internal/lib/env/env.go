package env

import (
	"os"

	"github.com/joho/godotenv"
)

type Env string

const (
	Local Env = "local"
	Dev   Env = "dev"
	Prod  Env = "prod"
)

func (e Env) Is(env Env) bool {
	return e == env
}

func Load() {
	_ = godotenv.Load()
	_ = godotenv.Load(os.Getenv("ENV_PATH"))
}
