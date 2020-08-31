package logging

import (
	"fmt"
	"io"
)

type ILogger interface {
	APIWriter() io.Writer
	Gin(v ...interface{})
	Rpc(v ...interface{})

	Debug(v ...interface{})
	Debugf(format string, v ...interface{})

	Info(v ...interface{})
	Infof(format string, v ...interface{})

	Warn(v ...interface{})
	Warnf(format string, v ...interface{})

	Error(v ...interface{})
	Errorf(format string, v ...interface{})

	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})

	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
}

func Logger() ILogger {
	return _formatlogger
}

type loggerImpl struct {
}
var _formatlogger = &loggerImpl{}

func (this *loggerImpl) APIWriter() io.Writer {
	return _loggers[file_flag_api].Writer()
}

func (this *loggerImpl) Gin(v ...interface{}) {
	_logmsgchs[file_flag_api] <- LogMessage{
		LEVEL_FLAG_GIN,
		v,
	}
}

func (this *loggerImpl) Rpc(v ...interface{}) {
	_logmsgchs[file_flag_api] <- LogMessage{
		LEVEL_FLAG_RPC,
		v,
	}
}

func (this *loggerImpl) Debug(v ...interface{}) {
	Debug(v)
}

func (this *loggerImpl) Debugf(format string, v ...interface{}) {
	Debug(fmt.Sprintf(format, v...))
}

func (this *loggerImpl) Info(v ...interface{}) {
	Info(v)
}

func (this *loggerImpl) Infof(format string, v ...interface{}) {
	Info(fmt.Sprintf(format, v...))
}

func (this *loggerImpl) Warn(v ...interface{}) {
	Warn(v)
}

func (this *loggerImpl) Warnf(format string, v ...interface{}) {
	Warn(fmt.Sprintf(format, v...))
}

func (this *loggerImpl) Error(v ...interface{}) {
	Error(v)
}

func (this *loggerImpl) Errorf(format string, v ...interface{}) {
	Error(fmt.Sprintf(format, v...))
}

func (this *loggerImpl) Fatal(v ...interface{}) {
	Fatal(v)
}

func (this *loggerImpl) Fatalf(format string, v ...interface{}) {
	Fatal(fmt.Sprintf(format, v...))
}

func (this *loggerImpl) Panic(v ...interface{}) {
	Fatal(v)
}

func (this *loggerImpl) Panicf(format string, v ...interface{}) {
	Fatal(fmt.Sprintf(format, v...))
}