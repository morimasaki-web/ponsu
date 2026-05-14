package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

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

	data, err := io.ReadAll(r)
	if err != nil {
		return attachmentsuc.PutResult{}, err
	}
	h := sha256.Sum256(data)

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return attachmentsuc.PutResult{}, err
	}

	return attachmentsuc.PutResult{SizeBytes: int64(len(data)), SHA256Hex: hex.EncodeToString(h[:])}, nil
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

func (s *S3Storage) GeneratePresignedURL(ctx context.Context, key string, filename string, expiration time.Duration) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("storage key is empty")
	}

	// Content-Dispositionヘッダーを設定してダウンロード時のファイル名を指定
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"", filename)

	presignClient := s3.NewPresignClient(s.client)
	out, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.bucket),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(contentDisposition),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("generate presigned url: %w", err)
	}
	return out.URL, nil
}
