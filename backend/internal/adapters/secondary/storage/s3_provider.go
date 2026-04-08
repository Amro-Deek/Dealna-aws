package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Provider struct {
	client     *s3.Client
	presignCli *s3.PresignClient
	bucketName string
}

func NewS3Provider(ctx context.Context, region, bucketName string) (*S3Provider, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	presignCli := s3.NewPresignClient(client)

	return &S3Provider{
		client:     client,
		presignCli: presignCli,
		bucketName: bucketName,
	}, nil
}

func (p *S3Provider) GeneratePresignedUploadURL(ctx context.Context, objectKey string, contentType string, expiry time.Duration) (string, error) {
	req, err := p.presignCli.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.bucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiry))

	if err != nil {
		return "", fmt.Errorf("unable to generate presigned URL: %w", err)
	}

	return req.URL, nil
}
