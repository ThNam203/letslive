package logger

import (
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
const LOG_FILE = "livestream_log.txt"

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

	Logger = zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(config), zapcore.AddSync(logFile), l)).Sugar()
}

func Warnw(message string, keysAndValues ...interface{}) {
	Logger.Warnw(message, keysAndValues...)
}

func Warnf(message string, keysAndValues ...interface{}) {
	Logger.Warnf(message, keysAndValues...)
}

func Infow(message string, keysAndValues ...interface{}) {
	Logger.Infow(message, keysAndValues...)
}

func Panicw(message string, keysAndValues ...interface{}) {
	Logger.Panicw(message, keysAndValues...)
}

func Debugw(message string, keysAndValues ...interface{}) {
	Logger.Debugw(message, keysAndValues...)
}

func Errorw(message string, keysAndValues ...interface{}) {
	Logger.Errorw(message, keysAndValues...)
}

func Infof(template string, keysAndValues ...interface{}) {
	Logger.Infof(template, keysAndValues...)
}

func Panicf(template string, keysAndValues ...interface{}) {
	Logger.Panicf(template, keysAndValues...)
}

func Debugf(template string, keysAndValues ...interface{}) {
	Logger.Debugf(template, keysAndValues...)
}

func Errorf(template string, keysAndValues ...interface{}) {
	Logger.Errorf(template, keysAndValues...)
}
