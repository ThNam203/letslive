package logger

import (
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func Init() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// defer logger.Sync() // flushes buffer, if any
	logger = zapLogger.Sugar()
}

func Infow(message string, keysAndValues ...interface{}) {
	logger.Infow(message, keysAndValues...)
}

func Panicw(message string, keysAndValues ...interface{}) {
	logger.Panicw(message, keysAndValues...)
}

func Debugw(message string, keysAndValues ...interface{}) {
	logger.Debugw(message, keysAndValues...)
}

func Errorw(message string, keysAndValues ...interface{}) {
	logger.Errorw(message, keysAndValues...)
}

func Infof(template string, keysAndValues ...interface{}) {
	logger.Infof(template, keysAndValues...)
}

func Panicf(template string, keysAndValues ...interface{}) {
	logger.Panicf(template, keysAndValues...)
}

func Debugf(template string, keysAndValues ...interface{}) {
	logger.Debugf(template, keysAndValues...)
}

func Errorf(template string, keysAndValues ...interface{}) {
	logger.Errorf(template, keysAndValues...)
}
