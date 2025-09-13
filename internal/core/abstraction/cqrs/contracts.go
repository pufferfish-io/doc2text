package cqrs

import (
	"context"
	"fmt"
)

type Command interface{ IsCommand() }
type Query interface{ IsQuery() }

type CommandHandler[C Command, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}
type QueryHandler[Q Query, R any] interface {
	Handle(ctx context.Context, q Q) (R, error)
}

type Bus struct {
	cmds map[string]any
	qrys map[string]any
}

func NewBus() *Bus {
	return &Bus{
		cmds: map[string]any{},
		qrys: map[string]any{},
	}
}

func typeKey[T any]() string {
	var t T
	return fmt.Sprintf("%T", t)
}

func RegisterCommand[C Command, R any](b *Bus, h CommandHandler[C, R]) {
	b.cmds[typeKey[C]()] = h
}

func RegisterQuery[Q Query, R any](b *Bus, h QueryHandler[Q, R]) {
	b.qrys[typeKey[Q]()] = h
}

func Exec[C Command, R any](b *Bus, ctx context.Context, cmd C) (R, error) {
	h, ok := b.cmds[typeKey[C]()].(CommandHandler[C, R])
	if !ok {
		var zero R
		return zero, fmt.Errorf("no handler for %T", cmd)
	}
	return h.Handle(ctx, cmd)
}

func Ask[Q Query, R any](b *Bus, ctx context.Context, q Q) (R, error) {
	h, ok := b.qrys[typeKey[Q]()].(QueryHandler[Q, R])
	if !ok {
		var zero R
		return zero, fmt.Errorf("no handler for %T", q)
	}
	return h.Handle(ctx, q)
}
