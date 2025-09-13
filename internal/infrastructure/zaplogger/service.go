package zaplogger

import (
	"go.uber.org/zap"

	corelog "doc2text/internal/core/abstraction/logger"
)

var _ corelog.Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	log *zap.SugaredLogger
}

func NewZapLogger() (*ZapLogger, func()) {
	l, _ := zap.NewProduction()
	return &ZapLogger{log: l.Sugar()}, func() { _ = l.Sync() }
}

func (z *ZapLogger) Info(msg string, args ...any)  { z.log.Infof(msg, args...) }
func (z *ZapLogger) Error(msg string, args ...any) { z.log.Errorf(msg, args...) }
