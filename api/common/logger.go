package common

import (
	"fmt"
	"github.com/nettyrnp/ads-crawler/config"
	"gopkg.in/natefinch/lumberjack.v2"
	"path"
	"time"
)

const (
	timeFormat = `02/Jan/2006:15:04:05 -0700`
)

var Logger *lumberjack.Logger

func InitLogger(c config.Config) {
	var fileName string
	if c.LogDir != "" {
		fileName = path.Join(c.LogDir, "app.log")
	}

	Logger = &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    c.LogMaxSize, // megabytes
		MaxBackups: c.LogBackups,
		MaxAge:     c.LogMaxAge, //days
		Compress:   c.LogCompress,
	}
}

func GetLog(c config.Config) (string, error) {
	return ReadFile(Logger.Filename)
}

func LogInfof(format string, a ...interface{}) {
	LogInfo(fmt.Sprintf(format, a...))
}

func LogErrorf(format string, a ...interface{}) {
	LogError(fmt.Sprintf(format, a...))
}

func LogFatalf(format string, a ...interface{}) {
	LogFatal(fmt.Sprintf(format, a...))
}

func LogInfo(msg string) {
	if Logger == nil {
		InitLogger(config.Config{})
	}
	Logger.Write([]byte(fmt.Sprintf("[%v] INFO: %v\n", time.Now().Format(timeFormat), msg)))
}

func LogError(msg string) {
	if Logger == nil {
		InitLogger(config.Config{})
	}
	Logger.Write([]byte(fmt.Sprintf("[%v] ERROR: %v\n", time.Now().Format(timeFormat), msg)))
}

func LogFatal(msg string) {
	if Logger == nil {
		InitLogger(config.Config{})
	}
	Logger.Write([]byte(fmt.Sprintf("[%v] FATAL: %v\n", time.Now().Format(timeFormat), msg)))
}
