package logger

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogLevel uint8

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)

var Logger *zap.SugaredLogger

func Init(level LogLevel) {
	// configure log option
	var l zapcore.Level
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	switch level {
	case Debug:
		l = zap.DebugLevel
	case Info:
		l = zap.InfoLevel
	case Warn:
		l = zap.WarnLevel
	case Error:
		l = zap.ErrorLevel
	default:
		if Logger == nil {
			l = zap.InfoLevel
		} else {
			return
		}
	}

	consoleWriter := zapcore.Lock(os.Stdout)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		consoleWriter,
		l,
	)

	baseLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	Logger = baseLogger.Sugar()
	defer Logger.Sync()
}

func appendAdditionalFieldsFromCtx(ctx context.Context, keysAndValues []any) {
	if ctx == nil {
		return
	}

	var oddCtx = true

	requestId, ok := ctx.Value("requestId").(string)
	if ok && len(requestId) > 0 {
		keysAndValues = append([]any{"requestId", requestId}, keysAndValues...)
	}

	// separate from request-specific error
	isSystemContext, ok := ctx.Value("systemContext").(string)
	if ok && isSystemContext == "true" {
		oddCtx = false
		keysAndValues = append([]any{"systemContext", "true"}, keysAndValues...)
	}

	// for debugging
	if oddCtx {
		keysAndValues = append([]any{"isOddContext", "true"}, keysAndValues...)
	}
}

func Warnw(ctx context.Context, message string, keysAndValues ...any) {
	appendAdditionalFieldsFromCtx(ctx, keysAndValues)
	Logger.Warnw(message, keysAndValues...)
}

func Warnf(ctx context.Context, message string, keysAndValues ...any) {
	appendAdditionalFieldsFromCtx(ctx, keysAndValues)
	Logger.Warnf(message, keysAndValues...)
}

func Infow(ctx context.Context, message string, keysAndValues ...any) {
	appendAdditionalFieldsFromCtx(ctx, keysAndValues)
	Logger.Infow(message, keysAndValues...)
}

func Panicw(ctx context.Context, message string, keysAndValues ...any) {
	appendAdditionalFieldsFromCtx(ctx, keysAndValues)
	Logger.Panicw(message, keysAndValues...)
}

func Debugw(ctx context.Context, message string, keysAndValues ...any) {
	appendAdditionalFieldsFromCtx(ctx, keysAndValues)
	Logger.Debugw(message, keysAndValues...)
}

func Errorw(ctx context.Context, message string, keysAndValues ...any) {
	appendAdditionalFieldsFromCtx(ctx, keysAndValues)
	Logger.Errorw(message, keysAndValues...)
}

func Infof(ctx context.Context, template string, keysAndValues ...any) {
	appendAdditionalFieldsFromCtx(ctx, keysAndValues)
	Logger.Infof(template, keysAndValues...)
}

func Panicf(ctx context.Context, template string, keysAndValues ...any) {
	appendAdditionalFieldsFromCtx(ctx, keysAndValues)
	Logger.Panicf(template, keysAndValues...)
}

func Debugf(ctx context.Context, template string, keysAndValues ...any) {
	appendAdditionalFieldsFromCtx(ctx, keysAndValues)
	Logger.Debugf(template, keysAndValues...)
}

func Errorf(ctx context.Context, template string, keysAndValues ...any) {
	appendAdditionalFieldsFromCtx(ctx, keysAndValues)
	Logger.Errorf(template, keysAndValues...)
}
