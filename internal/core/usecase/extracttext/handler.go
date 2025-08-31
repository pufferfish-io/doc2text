package extracttext

import (
	"context"
	"doc2text/internal/core/abstraction/convert"
	"doc2text/internal/core/abstraction/download"
	"doc2text/internal/core/abstraction/logger"
	"doc2text/internal/core/abstraction/recognize"
	"fmt"
)

type QueryHandler struct {
	fileConverter convert.FileConverter
	downloader    download.Downloader
	recognizer    recognize.Recognizer
}

func NewHandler(
	fc convert.FileConverter,
	d download.Downloader,
	l logger.Logger,
	r recognize.Recognizer) *QueryHandler {
	return &QueryHandler{
		fileConverter: fc,
		downloader:    d,
		recognizer:    r,
	}
}

func (h *QueryHandler) Handle(ctx context.Context, q Query) (Result, error) {
	fi, err := h.downloader.GetInfo(ctx, download.GetInfoRequest{ObjectKey: q.ObjectKey})
	f, err := h.downloader.GetFile(ctx, download.GetFileRequest{ObjectKey: q.ObjectKey})
	if err != nil {
		return Result{}, fmt.Errorf("extracttext: download by URL %q: %w", q.ObjectKey, err)
	}
	b64, err := h.fileConverter.ToBase64(ctx, convert.ToBase64Request{Data: f.Content})
	if err != nil {
		return Result{}, fmt.Errorf("extracttext: encode base64 for URL %q: %w", q.ObjectKey, err)
	}
	rcn, err := h.recognizer.Recognize(ctx, recognize.Request{ContentBase64: b64.Base64, MimeType: fi.MimeType})
	if err != nil {
		return Result{}, fmt.Errorf("extracttext: recognize text (url=%q, mime=%s): %w", q.ObjectKey, fi.MimeType, err)
	}
	return Result{Text: rcn.ExtractedText}, nil
}
