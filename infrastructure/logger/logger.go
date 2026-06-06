package logger

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ctxKey string

const requestIDKey ctxKey = "request-id"

func Init(mode string) {
	switch mode {
	case "dev":
		logx.SetUp(logx.LogConf{Mode: "console", Encoding: "plain", Level: "debug"})
	default:
		logx.SetUp(logx.LogConf{Mode: "console", Encoding: "json", Level: "info"})
	}
}

func WithContext(ctx context.Context) logx.Logger {
	logger := logx.WithContext(ctx)
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		logger = logger.WithFields(logx.Field("request_id", requestID))
	}
	return logger
}

func Info(ctx context.Context, format string, v ...interface{}) {
	WithContext(ctx).Infof(format, v...)
}

func Error(ctx context.Context, format string, v ...interface{}) {
	WithContext(ctx).Errorf(format, v...)
}

func Warn(ctx context.Context, format string, v ...interface{}) {
	WithContext(ctx).Errorf("[WARN] "+format, v...)
}

func Debug(ctx context.Context, format string, v ...interface{}) {
	WithContext(ctx).Debugf(format, v...)
}
