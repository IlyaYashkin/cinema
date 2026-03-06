package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	Member Role = "member"
	Admin  Role = "admin"
)

func (r Role) IsValid() bool {
	switch r {
	case Member, Admin:
		return true
	default:
		return false
	}
}

type User struct {
	Id        uuid.UUID
	Email     string
	Role      Role
	PassHash  []byte
	CreatedAt time.Time
}
