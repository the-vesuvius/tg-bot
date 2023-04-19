package logger

import (
	"go.uber.org/zap"
)

var _logger *zap.Logger

func init() {
	logger, err := NewLogger()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	_logger = logger
}

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func Get() *zap.Logger {
	return _logger
}
