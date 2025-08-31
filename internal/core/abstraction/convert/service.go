package convert

import "context"

type FileConverter interface {
	ToBase64(ctx context.Context, req ToBase64Request) (ToBase64Response, error)
}

type ToBase64Request struct {
	Data []byte
}

type ToBase64Response struct {
	Base64 string
}
