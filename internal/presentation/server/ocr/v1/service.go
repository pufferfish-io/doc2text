package ocr

import (
	"context"
	"strings"

	"doc2text/internal/core/abstraction/cqrs"
	"doc2text/internal/core/usecase/extracttext"
	ocrv1 "doc2text/internal/presentation/proto/ocr/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	ocrv1.UnimplementedOcrServiceServer
	bus *cqrs.Bus
}

func New(bus *cqrs.Bus) *Service { return &Service{bus: bus} }

func (s *Service) Process(ctx context.Context, req *ocrv1.ParseRequest) (*ocrv1.ParseResponse, error) {
	objectKey := strings.TrimSpace(req.GetObjectkey())
	if objectKey == "" {
		return nil, status.Error(codes.InvalidArgument, "objectkey is required")
	}

	q := extracttext.Query{ObjectKey: objectKey}

	res, err := cqrs.Ask[extracttext.Query, extracttext.Result](s.bus, ctx, q)
	if err != nil {
		return nil, err
	}
	return &ocrv1.ParseResponse{Text: res.Text}, nil
}
