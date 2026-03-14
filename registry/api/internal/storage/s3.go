package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3 provides upload and presigned download URLs for package artifacts.
type S3 struct {
	client            *s3.Client
	bucket            string
	domain            string // CloudFront domain for presigned URLs, or empty for S3
	presign           *s3.PresignClient
	useCustomEndpoint bool // true when using MinIO/LocalStack (disables SSE)
}

// NewS3 creates an S3 client. If endpoint is non-empty, it configures a custom
// S3-compatible endpoint (MinIO, LocalStack, etc.) with path-style addressing.
// publicEndpoint, if non-empty and different from endpoint, is used to create a
// separate presign client so that signatures are computed for the public hostname.
func NewS3(ctx context.Context, region, bucket, cloudFrontDomain, endpoint, publicEndpoint string) (*S3, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	var opts []func(*s3.Options)
	if endpoint != "" {
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true // required for MinIO / LocalStack
		})
	}

	client := s3.NewFromConfig(cfg, opts...)

	// If a public endpoint is configured and differs from the internal one,
	// create a separate client for presigning so the signature uses the
	// public hostname (otherwise MinIO rejects the request with 403).
	var presignClient *s3.PresignClient
	if publicEndpoint != "" && publicEndpoint != endpoint {
		var publicOpts []func(*s3.Options)
		publicOpts = append(publicOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(publicEndpoint)
			o.UsePathStyle = true
		})
		publicClient := s3.NewFromConfig(cfg, publicOpts...)
		presignClient = s3.NewPresignClient(publicClient)
	} else {
		presignClient = s3.NewPresignClient(client)
	}

	return &S3{
		client:            client,
		bucket:            bucket,
		domain:            cloudFrontDomain,
		presign:           presignClient,
		useCustomEndpoint: endpoint != "",
	}, nil
}

// PresignedDownloadURL returns a signed URL valid for 15 minutes.
// The presign client is already configured with the public endpoint (if any),
// so the returned URL uses the correct hostname and valid signature.
func (s *S3) PresignedDownloadURL(key string) (string, error) {
	req, err := s.presign.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// Upload writes body to the given key (implements packages.Storage).
// In production (no custom endpoint), enables AES-256 server-side encryption.
// Uses a 30-second timeout to prevent hanging on S3 failures.
func (s *S3) Upload(key string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(body),
	}

	// H6: Enable server-side encryption only for real AWS S3 (not MinIO/LocalStack).
	if !s.useCustomEndpoint {
		input.ServerSideEncryption = types.ServerSideEncryptionAes256
	}

	_, err := s.client.PutObject(ctx, input)
	return err
}

// Key returns the S3 key for a package file: namespace/name/version/filename.
func Key(namespace, name, version, filename string) string {
	if namespace == "" {
		return fmt.Sprintf("%s/%s/%s", name, version, filename)
	}
	return fmt.Sprintf("%s/%s/%s/%s", namespace, name, version, filename)
}
