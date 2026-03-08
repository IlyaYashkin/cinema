package s3

import (
	"cinema/internal/lib/config"
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	s3lib "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	log    *slog.Logger
	client *s3.Client
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

	return &S3{log: log, client: client}, nil
}

func (s *S3) Client() *s3.Client {
	return s.client
}

func (s *S3) MustConnect() {
	_, err := s.client.ListBuckets(context.Background(), &s3lib.ListBucketsInput{})
	if err != nil {
		panic("s3 connection failed: " + err.Error())
	}
	s.log.Info("s3 connection successful")
}
