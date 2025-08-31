package extracttext

import "context"

type Query struct {
	ObjectKey string
}

type Result struct {
	Text string
}

func (Query) IsQuery() {}

type Handler interface {
	Handle(ctx context.Context, q Query) (Result, error)
}
