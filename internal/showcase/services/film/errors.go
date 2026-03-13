package film

import "errors"

var (
	ErrIncorrectMIMEType = errors.New("incorrect mime type")
	ErrFilmNotFound      = errors.New("film not found")
	ErrFileNotFound      = errors.New("file not found")
)
