package logging

import (
	"context"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestContext(ctx context.Context, fields ...zap.Field) (context.Context, func() error) {
	one.Do(func() {
		je := zap.NewProductionEncoderConfig()
		je.EncodeTime = zapcore.ISO8601TimeEncoder

		core := zapcore.NewTee(
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(je),
				zapcore.AddSync(os.Stdout),
				zap.InfoLevel,
			),
		)
		logger = zap.New(core)
	})

	return context.WithValue(ctx, loggerKey, WithContext(ctx).With(fields...)), logger.Sync
}

func NewContext(ctx context.Context, fields ...zap.Field) (context.Context, func() error) {
	one.Do(func() {
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "app.log",
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
		})

		je := zap.NewProductionEncoderConfig()
		je.EncodeTime = zapcore.ISO8601TimeEncoder

		core := zapcore.NewTee(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(je),
				w,
				zap.InfoLevel,
			),
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(je),
				zapcore.AddSync(os.Stdout),
				zap.InfoLevel,
			),
		)
		logger = zap.New(core)
	})

	return context.WithValue(ctx, loggerKey, WithContext(ctx).With(fields...)), logger.Sync
}

func WithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return logger
	}

	if ctxLogger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return ctxLogger
	} else {
		return logger
	}
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Info(msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	WithContext(ctx).Error(msg, fields...)
}

type loggerKeyType int

const loggerKey loggerKeyType = iota

var logger *zap.Logger

var one sync.Once
