package logging

import (
	"github.com/sean-tech/gokit/validate"
	"log"
)

const (
	_defaultPrefix      = ""
	_defaultCallerDepth = 2
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

	_loggers = make(map[FileFlag]*log.Logger, 3)
	_logmsgchs = make(map[FileFlag]chan LogMessage, 3)
	filePath := getLogFilePath()

	for _, fileflag := range []FileFlag{file_flag_api, file_flag_info, file_flag_error} {
		fileName := getLogFileName(fileflag)
		file, err := openLogFile(fileName, filePath);
		if err != nil {
			log.Fatalln(err)
			continue
		}
		_lock.Lock()
		_loggers[fileflag] = log.New(file, _defaultPrefix, log.LstdFlags)
		_logmsgchs[fileflag] = make(chan LogMessage)
		_lock.Unlock()
		go logPrint(_loggers[fileflag], _logmsgchs[fileflag])
	}
	rotateTimingStart()
}

