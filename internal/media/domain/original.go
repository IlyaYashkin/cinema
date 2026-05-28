package domain

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	Uploading Status = "uploading"
	Uploaded  Status = "uploaded"
)

type Original struct {
	Id        uuid.UUID
	FilmId    uuid.UUID
	Key       string
	Status    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
