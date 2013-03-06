package simplelog

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"strings"
	"time"
)

const (
	DEBUG = iota
	INFO
	WARNING
	ERROR
)

const (
	red    = "\033[0;31;49m"
	green  = "\033[0;32;49m"
	yellow = "\033[0;33;49m"
	blue   = "\033[0;34;49m"
	reset  = "\033[0m"
)

var defaultLogger *Logger

func init() {
	defaultLogger = &Logger{level: INFO}
}

type Logger struct {
	sync.Mutex
	level int
}

func NewLogger(level int) *Logger {
	return &Logger{level: level}
}

func (l *Logger) SetLevel(lvl interface{}) error {
	switch lvl.(type) {
	case int:
		l.level = lvl.(int)
	case string:
		switch strings.ToLower(lvl.(string)) {
		case "debug":
			l.level = DEBUG
		case "info":
			l.level = INFO
		case "warning":
			l.level = WARNING
		case "error":
			l.level = ERROR
		default:
			return errors.New("invalid level")
		}
	default:
		return errors.New("invalid level")
	}
	return nil
}

func (l *Logger) Log(level int, s string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()

	var levelTxt string
	var color string

	if level >= l.level {
		switch level {
		case DEBUG:
			color = blue
			levelTxt = "DEBUG"
		case INFO:
			color = green
			levelTxt = "INFO"
		case WARNING:
			color = yellow
			levelTxt = "WARNING"
		case ERROR:
			color = red
			levelTxt = "ERROR"
		}

		dt := time.Now()
		year, month, day := dt.Date()
		hour, minute, second := dt.Clock()
		dateTime := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d.%06d", year, month, day,
			hour, minute, second,
			dt.Nanosecond()/1e3)

		logMsg := fmt.Sprintf(s, args...)
		fmt.Fprintf(os.Stderr, "%s[%s %s] %s%s\n", color, levelTxt, dateTime, logMsg, reset)
	}
}

func SetLevel(lvl interface{}) {
	defaultLogger.SetLevel(lvl)
}

func Debug(s string, args ...interface{}) {
	defaultLogger.Log(DEBUG, s, args...)
}

func Info(s string, args ...interface{}) {
	defaultLogger.Log(INFO, s, args...)
}

func Warning(s string, args ...interface{}) {
	defaultLogger.Log(WARNING, s, args...)
}

func Error(s string, args ...interface{}) {
	defaultLogger.Log(ERROR, s, args...)
}

func Log(level int, s string, args ...interface{}) {
	defaultLogger.Log(level, s, args...)
}
