package storage

import (
	"context"
	"fmt"
	"io"
	"kaffein/product-service/config"
	"kaffein/product-service/utils/constants"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client     *minio.Client
	bucketName string
}

func New(cfg *config.Config) (*Storage, error) {
	client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKeyID, cfg.Minio.SecretAccessKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf(constants.ErrMinioCreateClient, err)
	}

	exists, err := client.BucketExists(context.Background(), cfg.Minio.BucketName)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrMinioBucketExists, err)
	}

	if !exists {
		err = client.MakeBucket(context.Background(), cfg.Minio.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf(constants.ErrMinioBucketExists, err)
		}
	}

	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, cfg.Minio.BucketName)
	if err := client.SetBucketPolicy(context.Background(), cfg.Minio.BucketName, policy); err != nil {
		return nil, fmt.Errorf("set public product image policy: %w", err)
	}

	return &Storage{
		client:     client,
		bucketName: cfg.Minio.BucketName,
	}, nil
}

func (s *Storage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucketName, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf(constants.ErrMinioUpload, err)
	}

	return fmt.Sprintf("http://%s/%s/%s", s.client.EndpointURL().Host, s.bucketName, key), nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
}

func (s *Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucketName, key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf(constants.ErrMinioExists, err)
	}
	return true, nil
}
