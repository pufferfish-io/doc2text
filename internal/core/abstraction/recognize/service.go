package recognize

import "context"

type Recognizer interface {
	Recognize(ctx context.Context, req Request) (Response, error)
}

type Request struct {
	ContentBase64 string
	MimeType      string
}

type Response struct {
	ExtractedText string
}
