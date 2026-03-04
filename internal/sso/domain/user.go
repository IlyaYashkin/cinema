package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id        uuid.UUID
	Email     string
	PassHash  []byte
	CreatedAt time.Time
}
