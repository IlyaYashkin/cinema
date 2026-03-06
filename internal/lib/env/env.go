package env

import (
	"os"

	"github.com/joho/godotenv"
)

func Load() {
	_ = godotenv.Load()
	_ = godotenv.Load(os.Getenv("ENV_PATH"))
}
