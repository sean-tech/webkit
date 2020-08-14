package logging

import (
	"fmt"
	"github.com/robfig/cron"
	"github.com/sean-tech/gokit/foundation"
	"github.com/sean-tech/gokit/validate"
	"io"
	"log"
	"os"
	"sync"
)

type ILogger interface {
	Writer() io.Writer

	Gin(v ...interface{})
	Rpcx(v ...interface{})
	Db(v ...interface{})

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

const (
	_defaultPrefix      = ""
	_defaultCallerDepth = 2
)
var (
	_file 		*os.File
	_logger     *log.Logger
	_lock       sync.RWMutex
	_rotating 	sync.WaitGroup
	_logMsgCh	chan LogMessage = make(chan LogMessage)
)

type LogConfig struct {
	RunMode 		string	`validate:"required,oneof=debug test release"`
	LogSavePath 	string	`validate:"required,gt=1"`
	LogPrefix		string	`validate:"required,gt=1"`
}
var _config LogConfig

/**
 * setup
 */
func Setup(config LogConfig) {
	if err := validate.ValidateParameter(config); err != nil {
		log.Fatal(err)
	}
	_config = config

	var err error
	filePath := getLogFilePath()
	fileName := getLogFileName()
	_file, err = openLogFile(fileName, filePath)
	if err != nil {
		log.Fatalln(err)
	}
	_logger = log.New(_file, _defaultPrefix, log.LstdFlags)

	go logPrint(_logMsgCh)
	rotateTimingStart()
}

type LevelFlag string
const (

	LEVEL_FLAG_GIN 		= "GIN"
	LEVEL_FLAG_RPCX		= "RPCX"
	LEVEL_FLAG_DB		= "DB"
	LEVEL_FLAG_DEBUG	= "DEBUG"
	LEVEL_FLAG_INFO		= "INFO"
	LEVEL_FLAG_WARNING	= "WARN"
	LEVEL_FLAG_ERROR	= "ERROR"
	LEVEL_FLAG_FATAL	= "FATAL"
)

type LogMessage struct {
	Flag  LevelFlag
	Value []interface{}
}
func logPrint(msgCh <-chan LogMessage)  {
	for msg := range msgCh {
		setPrefix(msg.Flag)
		_rotating.Wait()
		_logger.Print(msg.Value...)
	}
}

func Debug(v ...interface{})  {
	if _config.RunMode == foundation.RUN_MODE_DEBUG {
		fmt.Println(v)
		_logMsgCh <- LogMessage{
			LEVEL_FLAG_DEBUG,
			v,
		}
	}
}

func Info(v ...interface{})  {
	_logMsgCh <- LogMessage{
		LEVEL_FLAG_INFO,
		v,
	}
}

func Warn(v ...interface{})  {
	_logMsgCh <- LogMessage{
		LEVEL_FLAG_WARNING,
		v,
	}
}

func Error(v ...interface{})  {
	_logMsgCh <- LogMessage{
		LEVEL_FLAG_ERROR,
		v,
	}
}

func Fatal(v ...interface{})  {
	_logMsgCh <- LogMessage{
		LEVEL_FLAG_FATAL,
		v,
	}
}

func setPrefix(flag LevelFlag)  {
	//_, file, line, ok := runtime.Caller(_defaultCallerDepth)
	var logPrefix string = fmt.Sprintf("[%s]", flag)
	//if ok {
	//	logPrefix = fmt.Sprintf("[%s]:[%s:%d]", flag, filepath.Base(file), line)
	//} else {
	//	logPrefix = fmt.Sprintf("[%s]", flag)
	//}
	_logger.SetPrefix(logPrefix)
}

func rotateTimingStart() {
	c := cron.New()
	spec := "0 0 0 * * *"
	if err := c.AddFunc(spec, func() {
		if logFileShouldRotate() == true {
			_rotating.Add(1)
			logFileRotate()
			_rotating.Done()
		}

	}); err != nil {
		log.Fatal(err)
	}
	c.Start()
}



type loggerImpl struct {}
var _formatlogger = &loggerImpl{}

func Logger() ILogger {
	return _formatlogger
}

func (this *loggerImpl) Writer() io.Writer {
	return _logger.Writer()
}

func (this *loggerImpl) Gin(v ...interface{}) {
	_logMsgCh <- LogMessage{
		LEVEL_FLAG_GIN,
		v,
	}
}

func (this *loggerImpl) Rpcx(v ...interface{}) {
	_logMsgCh <- LogMessage{
		LEVEL_FLAG_RPCX,
		v,
	}
}

func (this *loggerImpl) Db(v ...interface{}) {
	_logMsgCh <- LogMessage{
		LEVEL_FLAG_DB,
		v,
	}
}

func (this *loggerImpl) Debug(v ...interface{}) {
	Debug(v)
}

func (this *loggerImpl) Debugf(format string, v ...interface{}) {
	Debug(fmt.Sprintf(format, v))
}

func (this *loggerImpl) Info(v ...interface{}) {
	Info(v)
}

func (this *loggerImpl) Infof(format string, v ...interface{}) {
	Info(fmt.Sprintf(format, v))
}

func (this *loggerImpl) Warn(v ...interface{}) {
	Warn(v)
}

func (this *loggerImpl) Warnf(format string, v ...interface{}) {
	Warn(fmt.Sprintf(format, v))
}

func (this *loggerImpl) Error(v ...interface{}) {
	Error(v)
}

func (this *loggerImpl) Errorf(format string, v ...interface{}) {
	Error(fmt.Sprintf(format, v))
}

func (this *loggerImpl) Fatal(v ...interface{}) {
	Fatal(v)
}

func (this *loggerImpl) Fatalf(format string, v ...interface{}) {
	Fatal(fmt.Sprintf(format, v))
}

func (this *loggerImpl) Panic(v ...interface{}) {
	Fatal(v)
}

func (this *loggerImpl) Panicf(format string, v ...interface{}) {
	Fatal(fmt.Sprintf(format, v))
}



