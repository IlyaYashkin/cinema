package s3

import (
	"cinema/internal/lib/config"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
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

func (s *S3) Client() *s3.Client {
	return s.client
}

func (s *S3) MustConnect() {
	_, err := s.client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		panic("s3 connection failed: " + err.Error())
	}
	s.log.Info("s3 connection successful")
}

func (s *S3) MustCreateBucketIfNotExists(ctx context.Context) {
	const op = "s3.MustCreateBucketIfNotExists"

	log := s.log.With(
		slog.String("op", op),
		slog.String("bucket", s.bucket),
	)

	log.Info("creating s3 bucket")

	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		log.Info("s3 bucket already exists")

		return
	}

	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		panic("error creating s3 bucket: " + err.Error())
	}

	log.Info("s3 bucket created")
}

func (s *S3) MustSetupTemporaryFilesLifecycleConfigurationIfNotExists(ctx context.Context) {
	const op = "s3.MustSetupTemporaryFilesLifecycleConfigurationIfNotExists"
	const prefix = "tmp/"

	log := s.log.With(slog.String("op", op))

	conf, err := s.client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchLifecycleConfiguration" {
			log.Info("lifecycle configuration for temporary files not found")
		} else {
			panic("error getting bucket lifecycle configuration: " + err.Error())
		}
	}

	rules := make([]types.LifecycleRule, 0)

	if conf != nil {
		for _, rule := range conf.Rules {
			if rule.ID != nil && *rule.ID == prefix {
				log.Info("lifecycle configuration for temporary files already exists")

				return
			}
		}

		rules = conf.Rules
	}

	rules = append(rules, types.LifecycleRule{
		ID:     aws.String(prefix),
		Status: "Enabled",
		Filter: &types.LifecycleRuleFilter{
			Prefix: aws.String(prefix),
		},
		Expiration: &types.LifecycleExpiration{
			Days: aws.Int32(1),
		},
	})

	_, err = s.client.PutBucketLifecycleConfiguration(ctx, &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(s.bucket),
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{
			Rules: rules,
		},
	})
	if err != nil {
		panic("error setting bucket lifecycle configuration: " + err.Error())
	}

	log.Info("lifecycle configuration for temporary files created")
}
