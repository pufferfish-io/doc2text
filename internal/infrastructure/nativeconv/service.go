package nativeconv

import (
	"bytes"
	"context"
	"doc2text/internal/core/abstraction/convert"
	"encoding/base64"
)

type Base64FileConverter struct{}

func NewFileConverter() *Base64FileConverter {
	return &Base64FileConverter{}
}

func (c *Base64FileConverter) ToBase64(ctx context.Context, req convert.ToBase64Request) (convert.ToBase64Response, error) {
	select {
	case <-ctx.Done():
		return convert.ToBase64Response{}, ctx.Err()
	default:
	}

	const chunk = 64 << 10
	var b bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &b)

	for off := 0; off < len(req.Data); off += chunk {
		end := off + chunk
		if end > len(req.Data) {
			end = len(req.Data)
		}

		if _, err := enc.Write(req.Data[off:end]); err != nil {
			return convert.ToBase64Response{}, err
		}

		select {
		case <-ctx.Done():
			enc.Close()
			return convert.ToBase64Response{}, ctx.Err()
		default:
		}
	}
	if err := enc.Close(); err != nil {
		return convert.ToBase64Response{}, err
	}
	return convert.ToBase64Response{Base64: b.String()}, nil
}
