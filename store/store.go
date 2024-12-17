package store

import (
	"context"
	"io"
)

/*
 * oss 适配器
 * 适配minio
 */

type OssAdapter interface {
	PutObject(ctx context.Context, bucket, object string, reader io.Reader, size int64) error
	GetUploadId(ctx context.Context, bucket, object string) (string, error)
	GetPresignUrl(ctx context.Context, bucketName, objectName string, params Params) (string, error)
	ObjectExists(ctx context.Context, bucketName, objectName string) (bool, error)
	CompleteMultipart(ctx context.Context, bucketName, objectName string, uploadID string) error
}

type Store struct {
	Oss OssAdapter
}

func NewStore(oss OssAdapter) *Store {
	return &Store{oss}
}

func (s *Store) PutObject(ctx context.Context, bucket, object string, reader io.Reader, size int64) error {
	return s.Oss.PutObject(ctx, bucket, object, reader, size)
}
func (s *Store) GetPresignUrl(ctx context.Context, bucketName, objectName string, params Params) (string, error) {
	return s.GetPresignUrl(ctx, bucketName, objectName, params)
}
func (s *Store) GetUploadId(ctx context.Context, bucket, object string) (string, error) {
	return s.Oss.GetUploadId(ctx, bucket, object)
}
func (s *Store) ObjectExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	return s.Oss.ObjectExists(ctx, bucketName, objectName)
}
func (s *Store) CompleteMultipart(ctx context.Context, bucketName, objectName string, uploadID string) error {
	return s.CompleteMultipart(ctx, bucketName, objectName, uploadID)
}
func (s *Store) AbortMultipart(ctx context.Context, bucketName, objectName string, uploadID string) error {
	return s.AbortMultipart(ctx, bucketName, objectName, uploadID)
}
