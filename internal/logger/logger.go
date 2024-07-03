package logger

import (
	"os"
	"sync"

	"github.com/charmbracelet/log"
)

type Logger struct {
	*log.Logger
}

var (
	instance *Logger
	once     sync.Once
)

func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			Logger: log.NewWithOptions(os.Stderr, log.Options{
				ReportTimestamp: true,
				Level:           log.InfoLevel,
			}),
		}
	})
	return instance
}

func (l *Logger) SetLevel(level log.Level) {
	l.Logger.SetLevel(level)
}

func Debug(msg interface{}, keyvals ...interface{}) {
	GetLogger().Debug(msg, keyvals...)
}

func Info(msg interface{}, keyvals ...interface{}) {
	GetLogger().Info(msg, keyvals...)
}

func Warn(msg interface{}, keyvals ...interface{}) {
	GetLogger().Warn(msg, keyvals...)
}

func Error(msg interface{}, keyvals ...interface{}) {
	GetLogger().Error(msg, keyvals...)
}

func Fatal(msg interface{}, keyvals ...interface{}) {
	GetLogger().Fatal(msg, keyvals...)
}
