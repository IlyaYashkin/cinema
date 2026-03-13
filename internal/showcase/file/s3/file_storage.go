package s3

import (
	"cinema/internal/lib/file/s3"
)

type FileStorage struct {
	*s3.S3
}
