package s3

import (
	"cinema/internal/lib/config"
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	log        *slog.Logger
	client     *s3.Client
	bucket     string
	presignTTL time.Duration
}

func New(log *slog.Logger, cfg config.S3Config) (*S3, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &S3{log: log, client: client, bucket: cfg.Bucket, presignTTL: cfg.PresignTTL}, nil
}

func (s *S3) GetPresignedUploadURL(ctx context.Context, key string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(s.presignTTL))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (s *S3) GetRestrictedPresignedUploadURL(ctx context.Context, key string, contentType string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(s.presignTTL))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (s *S3) Client() *s3.Client {
	return s.client
}

func (s *S3) MustConnect() {
	_, err := s.client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		panic("s3 connection failed: " + err.Error())
	}

	_, err = s.client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		panic("s3 bucket not found: " + s.bucket)
	}

	s.log.Info("s3 connection successful")
}
