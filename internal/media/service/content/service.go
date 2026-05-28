package content

import (
	"cinema/internal/lib/sl"
	"cinema/internal/media/domain"
	"context"
	"fmt"
	"log/slog"
	"math"
	"path/filepath"

	"github.com/google/uuid"
)

const chunkSize int64 = 20 * 1024 * 1024 // 20 MB

type InitVideoUploadResult struct {
	UploadId      string
	Key           string
	PresignedURLs []string
	ChunkSize     int64
}

type Storage interface {
	CreateMultipartUpload(ctx context.Context, key string, contentType string) (string, error)
	PresignUploadParts(ctx context.Context, uploadId, key string, partsCount int) ([]string, error)
	CompleteMultipartUpload(ctx context.Context, uploadId string, key string, eTags []string) error
	AbortMultipartUpload(ctx context.Context, uploadId string, key string) error
}

type OriginalProvider interface {
	Create(ctx context.Context, filmId string, key string, status string) (string, error)
	GetByKey(ctx context.Context, key string) (domain.Original, error)
	UpdateStatus(ctx context.Context, id, status string) error
	DeleteByKey(ctx context.Context, key string) error
}

type Content struct {
	log *slog.Logger
	s   Storage
	op  OriginalProvider
}

func New(log *slog.Logger, s Storage, op OriginalProvider) *Content {
	return &Content{
		log: log,
		s:   s,
		op:  op,
	}
}

func (m *Content) InitUpload(
	ctx context.Context,
	filmId string,
	fileName string,
	contentType string,
	fileSize int64,
) (InitVideoUploadResult, error) {
	const op = "media.content.init_video_upload"

	log := m.log.With(
		slog.String("op", op),
		slog.String("fileName", fileName),
		slog.String("contentType", contentType),
		slog.String("fileSize", fmt.Sprintf("%v", fileSize)),
	)

	log.Info("initiating video upload")

	key := videoKey(fileName)

	cs := chunkSize

	if fileSize < cs {
		cs = fileSize
	}

	partsCount := int(math.Ceil(float64(fileSize) / float64(cs)))

	uploadId, err := m.s.CreateMultipartUpload(ctx, key, contentType)
	if err != nil {
		log.Error("failed creating multipart upload", sl.Err(err))

		return InitVideoUploadResult{}, sl.WrapErr(op, err)
	}

	presignedURLs, err := m.s.PresignUploadParts(ctx, uploadId, key, partsCount)
	if err != nil {
		log.Error("failed getting presigned urls", sl.Err(err))

		if abortErr := m.abortMultipartUpload(ctx, log, uploadId, key); abortErr != nil {
			return InitVideoUploadResult{}, sl.WrapErr(op, abortErr)
		}

		return InitVideoUploadResult{}, sl.WrapErr(op, err)
	}

	_, err = m.op.Create(ctx, filmId, key, string(domain.Uploading))
	if err != nil {
		log.Error("failed creating upload record", sl.Err(err))

		return InitVideoUploadResult{}, sl.WrapErr(op, err)
	}

	return InitVideoUploadResult{
		UploadId:      uploadId,
		Key:           key,
		PresignedURLs: presignedURLs,
		ChunkSize:     cs,
	}, nil
}

func (m *Content) CompleteUpload(
	ctx context.Context,
	uploadId string,
	key string,
	eTags []string,
) error {
	const op = "media.content.complete_upload"

	log := m.log.With(
		slog.String("op", op),
		slog.String("uploadId", uploadId),
		slog.String("key", key),
		slog.String("eTags", fmt.Sprintf("%v", eTags)),
	)

	log.Info("completing upload")

	err := m.s.CompleteMultipartUpload(ctx, uploadId, key, eTags)
	if err != nil {
		log.Error("failed completing multipart upload", sl.Err(err))

		if abortErr := m.abortMultipartUpload(ctx, log, uploadId, key); abortErr != nil {
			return sl.WrapErr(op, abortErr)
		}

		return sl.WrapErr(op, err)
	}

	original, err := m.op.GetByKey(ctx, key)
	if err != nil {
		log.Error("failed getting original", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	err = m.op.UpdateStatus(ctx, original.Id.String(), string(domain.Uploaded))
	if err != nil {
		log.Error("failed updating original record status", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	return nil
}

func (m *Content) AbortUpload(
	ctx context.Context,
	uploadId string,
	key string,
) error {
	const op = "media.content.abort_upload"

	log := m.log.With(
		slog.String("op", op),
		slog.String("uploadId", uploadId),
		slog.String("key", key),
	)

	log.Info("aborting upload")

	if abortErr := m.abortMultipartUpload(ctx, log, uploadId, key); abortErr != nil {
		return sl.WrapErr(op, abortErr)
	}

	err := m.op.DeleteByKey(ctx, key)
	if err != nil {
		log.Error("failed deleting original record", sl.Err(err))

		return sl.WrapErr(op, err)
	}

	return nil
}

func (m *Content) abortMultipartUpload(ctx context.Context, log *slog.Logger, uploadId, key string) error {
	err := m.s.AbortMultipartUpload(ctx, uploadId, key)
	if err != nil {
		log.Error("failed aborting multipart upload", sl.Err(err))

		return err
	}

	return nil
}

func videoKey(fileName string) string {
	ext := filepath.Ext(fileName)
	return fmt.Sprintf("originals/%s%s", uuid.New().String(), ext)
}
