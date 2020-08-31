package logging

import (
	"fmt"
	"github.com/robfig/cron"
	"github.com/sean-tech/gokit/foundation"
	"log"
	"sync"
)

var (
	_lock      sync.RWMutex
	_rotating  sync.WaitGroup
	_loggers   map[FileFlag]*log.Logger     = make(map[FileFlag]*log.Logger, 3)
	_logmsgchs map[FileFlag]chan LogMessage = make(map[FileFlag]chan LogMessage, 3)
)

type LevelFlag string; type FileFlag string
const (
	LEVEL_FLAG_GIN     	LevelFlag = "GIN"
	LEVEL_FLAG_RPC     	LevelFlag = "RPC"
	LEVEL_FLAG_INFO    	LevelFlag = "INFO"
	LEVEL_FLAG_WARNING 	LevelFlag = "WARN"
	LEVEL_FLAG_ERROR   	LevelFlag = "ERROR"
	LEVEL_FLAG_FATAL   	LevelFlag = "FATAL"

	file_flag_api   FileFlag = "api"
	file_flag_info  FileFlag = "info"
	file_flag_error FileFlag = "error"
)

type LogMessage struct {
	Flag  LevelFlag
	Value []interface{}
}

func logPrint(logger *log.Logger, msgCh <-chan LogMessage)  {
	for msg := range msgCh {
		logger.SetPrefix(fmt.Sprintf("[%s]", msg.Flag))
		_rotating.Wait()
		if msg.Flag == LEVEL_FLAG_FATAL {
			logger.Fatal(msg.Value...)
		} else {
			logger.Print(msg.Value...)
		}
	}
}

func Debug(v ...interface{})  {
	if _config.RunMode == foundation.RUN_MODE_DEBUG {
		fmt.Println(v)
		_logmsgchs[file_flag_info] <- LogMessage{
			LEVEL_FLAG_INFO,
			v,
		}
	}
}

func Info(v ...interface{})  {
	_logmsgchs[file_flag_info] <- LogMessage{
		LEVEL_FLAG_INFO,
		v,
	}
}

func Warn(v ...interface{})  {
	_logmsgchs[file_flag_info] <- LogMessage{
		LEVEL_FLAG_WARNING,
		v,
	}
}

func Error(v ...interface{})  {
	_logmsgchs[file_flag_error] <- LogMessage{
		LEVEL_FLAG_ERROR,
		v,
	}
}

func Fatal(v ...interface{})  {
	_logmsgchs[file_flag_error] <- LogMessage{
		LEVEL_FLAG_FATAL,
		v,
	}
}

func rotateTimingStart() {
	c := cron.New()
	spec := "0 0 0 * * *"
	if err := c.AddFunc(spec, func() {
		_rotating.Add(1)
		if logFileShouldRotate(file_flag_api) {
			logFileRotate(file_flag_api)
		}
		if logFileShouldRotate(file_flag_info) {
			logFileRotate(file_flag_info)
		}
		if logFileShouldRotate(file_flag_error) {
			logFileRotate(file_flag_error)
		}
		_rotating.Done()
	}); err != nil {
		log.Fatal(err)
	}
	c.Start()
}




