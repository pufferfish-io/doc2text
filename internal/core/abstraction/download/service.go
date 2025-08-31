package download

import "context"

type Downloader interface {
	GetInfo(ctx context.Context, req GetInfoRequest) (GetInfoResponse, error)
	GetFile(ctx context.Context, req GetFileRequest) (GetFileResponse, error)
}

type GetInfoRequest struct {
	ObjectKey string
}

type GetInfoResponse struct {
	MimeType string
}

type GetFileRequest struct {
	ObjectKey string
}

type GetFileResponse struct {
	Content []byte
}
