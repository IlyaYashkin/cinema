package storage

import "errors"

var (
	ErrOriginalKeyExists = errors.New("original already key exists")
	ErrOriginalNotFound  = errors.New("original not found")
)
