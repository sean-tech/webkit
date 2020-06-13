package logging

import (
	"fmt"
	"github.com/sean-tech/gokit/fileutils"
	"os"
	"time"
)

const (
	__time_format = "20060101"
	__logfile_ext = "log"
)


func getLogFilePath() string {
	return _config.LogSavePath
}

func getLastDayLogFileName() string {
	lastDayTime := time.Now().AddDate(0, 0, -1)
	return fmt.Sprintf("%s_%s.%s",
		_config.LogPrefix,
		lastDayTime.Format(__time_format),
		__logfile_ext,
	)
}

func getLogFileName() string {
	return fmt.Sprintf("%s.%s",
		_config.LogPrefix,
		__logfile_ext,
	)
}

func openLogFile(fileName, filePath string) (*os.File, error) {

	src := filePath
	perm := fileutils.CheckPermission(src)
	if perm == true {
		return nil, fmt.Errorf("file.CheckPermission Permission denied src: %s", src)
	}

	err := fileutils.MKDirIfNotExist(src)
	if err != nil {
		return nil, fmt.Errorf("file.IsNotExistMkDir src: %s, err: %v", src, err)
	}

	f, err := fileutils.Open(src + fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("Fail to OpenFile :%v", err)
	}

	return f, nil
}

/**
 * return result of log file should rotate, if should, return true
 */
func logFileShouldRotate() bool {
	// 默认日志文件不存在，无初始化文件，不继续处理
	currentLogFileExist := fileutils.CheckExist(getLogFilePath() + getLogFileName())
	if !currentLogFileExist {
		return false
	}
	// 昨日日志文件存在，说明已处理，不继续处理
	lastDayLogFileExist := fileutils.CheckExist(getLogFilePath() + getLastDayLogFileName())
	if lastDayLogFileExist {
		return false
	}
	return true
}

/**
 * log file rotate
 */
func logFileRotate() error {
	src := getLogFilePath() + getLogFileName()
	dst := getLogFilePath() + getLastDayLogFileName()
	if _, err := fileutils.CopyFile(dst, src); err != nil {
		return err
	}
	return fileutils.ClearFile(src)
}

/**
 * unused
 */
func fileTimePassDaySlice() bool {
	// 默认日志文件不存在，无初始化文件，不继续处理
	currentLogFileExist := fileutils.CheckExist(getLogFilePath() + getLogFileName())
	if !currentLogFileExist {
		return false
	}
	// 昨日日志文件存在，说明已处理，不继续处理
	lastDayLogFileExist := fileutils.CheckExist(getLogFilePath() + getLastDayLogFileName())
	if lastDayLogFileExist {
		return false
	}
	// 把当前日志文件重命名为昨日日志文件
	originalPath := getLogFilePath() + getLogFileName()
	newPath := getLogFilePath() + getLastDayLogFileName()
	err := os.Rename(originalPath, newPath)
	if err != nil {
		Error(err)
		return false
	}
	return true
}








