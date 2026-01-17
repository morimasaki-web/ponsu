package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	attachmentsuc "github.com/morimasaki-web/ponsu/internal/usecase/attachments"
)

type S3StorageConfig struct {
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
}

type S3Storage struct {
	bucket string
	client *s3.Client
}

func NewS3Storage(ctx context.Context, cfg S3StorageConfig) (*S3Storage, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, errors.New("s3 bucket is empty")
	}
	if strings.TrimSpace(cfg.Region) == "" {
		cfg.Region = "us-east-1"
	}
	if strings.TrimSpace(cfg.AccessKey) == "" {
		return nil, errors.New("s3 access key is empty")
	}
	if strings.TrimSpace(cfg.SecretKey) == "" {
		return nil, errors.New("s3 secret key is empty")
	}

	awsCfg, err := awscfg.LoadDefaultConfig(
		ctx,
		awscfg.WithRegion(cfg.Region),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if strings.TrimSpace(cfg.Endpoint) != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.ForcePathStyle
	})

	return &S3Storage{bucket: cfg.Bucket, client: client}, nil
}

func (s *S3Storage) Put(ctx context.Context, key string, r io.Reader) (attachmentsuc.PutResult, error) {
	if strings.TrimSpace(key) == "" {
		return attachmentsuc.PutResult{}, errors.New("storage key is empty")
	}

	h := sha256.New()
	cr := &countingReader{r: io.TeeReader(r, h)}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   cr,
	})
	if err != nil {
		return attachmentsuc.PutResult{}, err
	}

	return attachmentsuc.PutResult{SizeBytes: cr.n, SHA256Hex: hex.EncodeToString(h.Sum(nil))}, nil
}

func (s *S3Storage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("storage key is empty")
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" {
		return errors.New("storage key is empty")
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}
