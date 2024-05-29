package minio

import (
	"context"
	"fmt"
	"github.com/Vladislav747/minio-project/pkg/logging"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"time"
)

type Object struct {
	ID   string
	Size int64
	Tags map[string]string
}

type Client struct {
	logger      logging.Logger
	minioClient *minio.Client
}

// Создание самого клиента
func NewClient(endpoint, accessKeyID, secretAccessKey string, logger logging.Logger) (*Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &Client{
		logger:      logger,
		minioClient: minioClient,
	}, nil
}

// Получение файла из бакета
func (c *Client) GetFile(ctx context.Context, bucketName, fileId string) (*minio.Object, error) {
	reqCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	obj, err := c.minioClient.GetObject(reqCtx, bucketName, fileId, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file %s:from minio bucket %s err: %w", fileId, bucketName, err)
	}
	return obj, nil
}

func (c *Client) GetBucketFiles(ctx context.Context, bucketName string) ([]*minio.Object, error) {
	reqCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	var files []*minio.Object
	for lobj := range c.minioClient.ListObjects(reqCtx, bucketName, minio.ListObjectsOptions{WithMetadata: true}) {
		if lobj.Err != nil {
			c.logger.Errorf("failed to list object in bucket %s: err: %w", bucketName, lobj.Err)
			continue
		}
		object, err := c.minioClient.GetObject(reqCtx, bucketName, lobj.Key, minio.GetObjectOptions{})
		if err != nil {
			c.logger.Errorf("failed to get object key=%s in bucket %s: err: %w", lobj.Key, bucketName, err)
			continue
		}
		files = append(files, object)
	}
	return files, nil
}

func (c *Client) UploadFile(ctx context.Context, fileId, fileName, bucketName string, fileSize int64, reader io.Reader) error {
	reqCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	exists, errBucketExists := c.minioClient.BucketExists(reqCtx, bucketName)
	if errBucketExists != nil || !exists {
		c.logger.Warnf("no such bucket: %s. creating new one...", bucketName)
		err := c.minioClient.MakeBucket(reqCtx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
		}
	}

	c.logger.Debugf("put new object %s in bucket %s", fileName, bucketName)

	_, err := c.minioClient.PutObject(reqCtx, bucketName, fileId, reader, fileSize,
		minio.PutObjectOptions{
			UserMetadata: map[string]string{
				"Name": fileName,
			},
			ContentType: "application/octet-stream",
		})
	if err != nil {
		return fmt.Errorf("failed to upload file. err: %w", err)
	}

	return nil
}

func (c *Client) DeleteFile(ctx context.Context, noteUUID, fileName string) error {
	err := c.minioClient.RemoveObject(ctx, noteUUID, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", fileName, err)
	}
	return nil
}
