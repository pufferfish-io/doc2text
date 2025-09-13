package s3

import (
	"bytes"
	"context"
	"doc2text/internal/core/abstraction/download"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type s3Downloader struct {
	client *minio.Client
	bucket string
}

const (
	defaultMimeType = "application/octet-stream"
)

func NewDownloader(cfg Config) (download.Downloader, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &s3Downloader{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (d *s3Downloader) GetInfo(ctx context.Context, req download.GetInfoRequest) (download.GetInfoResponse, error) {
	info, err := d.client.StatObject(ctx, d.bucket, req.ObjectKey, minio.StatObjectOptions{})
	if err != nil {
		return download.GetInfoResponse{}, fmt.Errorf("stat object: %w", err)
	}

	mimeType := info.ContentType
	if mimeType == "" {
		ext := filepath.Ext(req.ObjectKey)
		if ext != "" {
			if byExt := mime.TypeByExtension(ext); byExt != "" {
				mimeType = byExt
			}
		}
		if mimeType == "" {
			mimeType = defaultMimeType
		}
	}

	return download.GetInfoResponse{
		MimeType: mimeType,
	}, nil
}

func (d *s3Downloader) GetFile(ctx context.Context, req download.GetFileRequest) (download.GetFileResponse, error) {
	obj, err := d.client.GetObject(ctx, d.bucket, req.ObjectKey, minio.GetObjectOptions{})
	if err != nil {
		return download.GetFileResponse{}, fmt.Errorf("get object: %w", err)
	}
	defer obj.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, obj); err != nil {
		return download.GetFileResponse{}, fmt.Errorf("read object: %w", err)
	}
	content := buf.Bytes()

	_ = http.DetectContentType(content)

	return download.GetFileResponse{
		Content: content,
	}, nil
}
