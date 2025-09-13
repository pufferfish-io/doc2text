package yocr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"doc2text/internal/core/abstraction/recognize"
)

type Options struct {
	OcrEndpoint string
	ApiKey      string
	FolderID    string
	Model       string
	Languages   []string
}

type ycRecognizer struct {
	ocrEndpoint string
	apiKey      string
	folderID    string
	model       string
	languages   []string
	http        *http.Client
}

func New(o Options) recognize.Recognizer {
	return &ycRecognizer{
		ocrEndpoint: o.OcrEndpoint,
		apiKey:      o.ApiKey,
		folderID:    o.FolderID,
		model:       o.Model,
		languages:   o.Languages,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *ycRecognizer) Recognize(ctx context.Context, req recognize.Request) (recognize.Response, error) {
	if req.ContentBase64 == "" {
		return recognize.Response{}, errors.New("empty ContentBase64")
	}
	if req.MimeType == "" {
		return recognize.Response{}, errors.New("empty MimeType")
	}

	raw, status, err := r.send(ctx, req.MimeType, req.ContentBase64)
	if err != nil {
		if status != 0 {
			return recognize.Response{}, fmt.Errorf("yandex ocr http %d: %w", status, err)
		}
		return recognize.Response{}, err
	}

	text, err := r.parse(raw)
	if err != nil {
		return recognize.Response{}, err
	}

	return recognize.Response{ExtractedText: text}, nil
}

type ycRequest struct {
	MimeType      string   `json:"mimeType"`
	LanguageCodes []string `json:"languageCodes,omitempty"`
	Model         string   `json:"model,omitempty"`
	ContentB64    string   `json:"content"`
}

type ycError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (r *ycRecognizer) send(ctx context.Context, mimeType, contentB64 string) (raw []byte, status int, err error) {
	payload := ycRequest{
		MimeType:      mimeType,
		LanguageCodes: r.languages,
		Model:         r.model,
		ContentB64:    contentB64,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.ocrEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Api-Key "+r.apiKey)
	req.Header.Set("x-folder-id", r.folderID)

	res, err := r.http.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, 0, err
		}
		return nil, 0, fmt.Errorf("http do: %w", err)
	}
	defer res.Body.Close()

	raw, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, res.StatusCode, fmt.Errorf("read body: %w", readErr)
	}

	if res.StatusCode/100 != 2 {
		var apiErr ycError
		_ = json.Unmarshal(raw, &apiErr)
		if apiErr.Message != "" {
			return raw, res.StatusCode, fmt.Errorf("%s (code=%d)", apiErr.Message, apiErr.Code)
		}
		return raw, res.StatusCode, fmt.Errorf("unexpected status %d: %s", res.StatusCode, string(raw))
	}

	return raw, res.StatusCode, nil
}

type ycResponse struct {
	Result struct {
		TextAnnotation struct {
			FullText string `json:"fullText"`
			Blocks   []struct {
				Lines []struct {
					Text string `json:"text"`
				} `json:"lines"`
			} `json:"blocks"`
		} `json:"textAnnotation"`
	} `json:"result"`
}

func (r *ycRecognizer) parse(raw []byte) (string, error) {
	var ycr ycResponse
	if err := json.Unmarshal(raw, &ycr); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if txt := ycr.Result.TextAnnotation.FullText; txt != "" {
		return txt, nil
	}

	var buf bytes.Buffer
	for _, b := range ycr.Result.TextAnnotation.Blocks {
		for _, ln := range b.Lines {
			if ln.Text != "" {
				buf.WriteString(ln.Text)
				buf.WriteByte('\n')
			}
		}
	}
	return buf.String(), nil
}
