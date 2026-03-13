package domain

import "github.com/google/uuid"

type Film struct {
	Id          uuid.UUID
	Name        string
	Description string
	PosterUrl   *string
}
