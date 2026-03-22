package s3

import (
	"cinema/internal/lib/config"
	"cinema/internal/lib/file"
	"cinema/internal/lib/sl"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3 struct {
	log        *slog.Logger
	client     *s3.Client
	endpoint   string
	bucket     string
	presignTTL time.Duration
}

func New(log *slog.Logger, cfg config.S3Config) (*S3, error) {
	const op = "lib.s3.new"

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, sl.WrapErr(op, err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &S3{log: log, client: client, endpoint: cfg.Endpoint, bucket: cfg.Bucket, presignTTL: cfg.PresignTTL}, nil
}

func (s *S3) GetPresignedUploadURL(ctx context.Context, key string) (string, error) {
	const op = "lib.s3.get_presigned_upload_url"

	presignClient := s3.NewPresignClient(s.client)
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(s.presignTTL))
	if err != nil {
		return "", sl.WrapErr(op, err)
	}
	return req.URL, nil
}

func (s *S3) GetRestrictedPresignedUploadURL(ctx context.Context, key string, contentType string) (string, error) {
	const op = "lib.s3.get_restricted_presigned_upload_url"

	presignClient := s3.NewPresignClient(s.client)
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(s.presignTTL))
	if err != nil {
		return "", sl.WrapErr(op, err)
	}
	return req.URL, nil
}

func (s *S3) GetFileContentType(ctx context.Context, key string) (string, error) {
	const op = "lib.s3.get_file_content_type"

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Range:  aws.String("bytes=0-511"),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return "", sl.WrapErr(op, file.ErrFileNotFound)
		}

		return "", sl.WrapErr(op, err)
	}

	buf := make([]byte, 512)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		return "", sl.WrapErr(op, err)
	}

	err = resp.Body.Close()
	if err != nil {
		return "", sl.WrapErr(op, err)
	}

	return http.DetectContentType(buf[:n]), nil
}

func (s *S3) DeleteFile(ctx context.Context, key string) error {
	const op = "lib.s3.delete_file"

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}

func (s *S3) CopyFile(ctx context.Context, keyFrom string, keyTo string) error {
	const op = "lib.s3.copy_file"

	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		CopySource: aws.String(s.bucket + "/" + keyFrom),
		Key:        aws.String(keyTo),
	})
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}

func (s *S3) MoveFile(ctx context.Context, keyFrom string, keyTo string) error {
	const op = "lib.s3.move_file"

	err := s.CopyFile(ctx, keyFrom, keyTo)
	if err != nil {
		return sl.WrapErr(op, err)
	}
	err = s.DeleteFile(ctx, keyFrom)
	if err != nil {
		return sl.WrapErr(op, err)
	}

	return nil
}

func (s *S3) GetFilePath(key string) string {
	return s.endpoint + "/" + s.bucket + "/" + key
}

func (s *S3) Client() *s3.Client {
	return s.client
}

func (s *S3) MustConnect() {
	const op = "lib.s3.must_connect"

	_, err := s.client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		panic(sl.WrapErr(op, err))
	}

	_, err = s.client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		panic(sl.WrapErr(op, err))
	}

	s.log.With(slog.String("op", op)).Info("s3 connection successful")
}

func (s *S3) Close() {}
