package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3 provides upload and presigned download URLs for package artifacts.
type S3 struct {
	client   *s3.Client
	bucket   string
	domain   string // CloudFront domain for presigned URLs, or empty for S3
	presign  *s3.PresignClient
}

// NewS3 creates an S3 client from environment (AWS_REGION, etc.).
func NewS3(ctx context.Context, region, bucket, cloudFrontDomain string) (*S3, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg)
	return &S3{
		client:  client,
		bucket:  bucket,
		domain:  cloudFrontDomain,
		presign: s3.NewPresignClient(client),
	}, nil
}

// PresignedDownloadURL returns a signed URL valid for 1 hour.
func (s *S3) PresignedDownloadURL(key string) (string, error) {
	req, err := s.presign.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(1 * time.Hour)
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// Upload writes body to the given key (implements packages.Storage).
func (s *S3) Upload(key string, body []byte) error {
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(body),
	})
	return err
}

// Key returns the S3 key for a package file: namespace/name/version/filename.
func Key(namespace, name, version, filename string) string {
	if namespace == "" {
		return fmt.Sprintf("%s/%s/%s", name, version, filename)
	}
	return fmt.Sprintf("%s/%s/%s/%s", namespace, name, version, filename)
}
