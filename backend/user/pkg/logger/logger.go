package logger

import (
	"context"
	"log"
	"os"
	"path/filepath"
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

// const dateTimeFormat = "[30/11/2024 17:30:24]"
const LOG_FOLDER = "logs"
const LOG_FILE = "user_log.txt"

func Init(level LogLevel) {
	// configure log option
	var l zapcore.Level
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	//config.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	//	enc.AppendString(t.Format(dateTimeFormat))
	//}

	e, err := os.Executable()
	if err != nil {
		panic("failed to get the execute path")
	}
	logFile, err := os.OpenFile(filepath.Join(filepath.Dir(e), LOG_FOLDER, LOG_FILE), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		defer logFile.Close()
		log.Panicf("failed to open log file: %s", err)
	}

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

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		zapcore.AddSync(logFile),
		l,
	)

	baseLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	Logger = baseLogger.Sugar()
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
