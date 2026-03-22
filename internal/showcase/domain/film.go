package domain

import "github.com/google/uuid"

type FilmImage struct {
	Id  int64
	Url string
}

type Film struct {
	Id          uuid.UUID
	Name        string
	Description string
	PosterUrl   *string
	Images      []FilmImage
}
