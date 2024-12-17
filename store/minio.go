package store

import (
	"context"
	"github.com/cylScripter/chest/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/url"
	"strconv"
	"time"
)

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	BaseHost  string
}

type Params struct {
	UploadId   string
	PartNumber int
}

type MinioStore struct {
	MinioConfig
	client *minio.Client
	core   *minio.Core
}

func NewMinioStore(cfg MinioConfig) (*MinioStore, error) {
	minioStore := &MinioStore{
		MinioConfig: cfg,
	}
	// Initialize minio client object.
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Errorf("NewMinioStore failed, err:%v", err)
		return nil, err
	}
	coreClient, err := minio.NewCore(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		log.Errorf("NewCore failed, err:%v", err)
		return nil, err
	}
	minioStore.client = minioClient
	minioStore.core = coreClient
	return minioStore, nil
}

func (m *MinioStore) PutObject(ctx context.Context, bucket, object string, reader io.Reader, size int64) error {
	putInfo, err := m.client.PutObject(ctx, bucket, object, reader, size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		log.CtxErrorf(ctx, "PutObject failed, err:%v", err)
		return err
	}
	log.CtxInfof(ctx, "putInfo:%v", putInfo)
	return nil
}

func (m *MinioStore) GetUploadId(ctx context.Context, bucket, object string) (string, error) {
	if m.core == nil {
		log.CtxErrorf(ctx, "core is nil")
		return "", nil
	}
	upload, err := m.core.NewMultipartUpload(ctx, bucket, object, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		log.CtxErrorf(ctx, "NewMultipartUpload failed, err:%v", err)
		return "", err
	}
	return upload, nil
}

func (m *MinioStore) GetPresignUrl(ctx context.Context, bucketName, objectName string, params Params) (string, error) {
	presign, err := m.client.Presign(ctx, "PUT", bucketName, objectName, time.Second*60*60, url.Values{
		"uploadId":   []string{params.UploadId},
		"partNumber": []string{strconv.Itoa(params.PartNumber)},
	})
	if err != nil {
		log.Errorf("presign err %v", err)
		return "", err
	}
	presign.Host = m.BaseHost
	return presign.String(), nil
}

func (m *MinioStore) ObjectExists(ctx context.Context, bucketName, objectName string) (bool, error) {
	_, err := m.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		log.CtxErrorf(ctx, "ObjectExists failed, err:%v", err)
		// 如果是 NoSuchKey 错误，则对象不存在
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *MinioStore) CompleteMultipart(ctx context.Context, bucketName, objectName string, uploadID string) error {
	parts, err := m.core.ListObjectParts(ctx, bucketName, objectName, uploadID, 0, 10000)
	if err != nil {
		log.CtxErrorf(ctx, "ListObjectParts failed, err:%v", err)
		return err
	}
	completePart := make([]minio.CompletePart, 0)
	for _, part := range parts.ObjectParts {
		completePart = append(completePart, minio.CompletePart{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		})
	}
	if len(completePart) == 0 {
		return nil
	}
	_, err = m.core.CompleteMultipartUpload(ctx, bucketName, objectName, uploadID, completePart, minio.PutObjectOptions{})
	if err != nil {
		log.CtxErrorf(ctx, "CompleteMultipartUpload failed, err:%v", err)
		return err
	}
	return nil
}

func (m *MinioStore) AbortMultipart(ctx context.Context, bucketName, objectName string, uploadID string) error {
	return m.core.AbortMultipartUpload(ctx, bucketName, objectName, uploadID)
}
