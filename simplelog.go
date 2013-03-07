package simplelog

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"strings"
	"time"
	"unsafe"
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
var istty bool

func init() {
	defaultLogger = &Logger{level: INFO}
	istty = isatty(os.Stderr)
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

	if level < l.level {
		return
	}
	
	postfix := reset
	prefix, levelTxt := parseLevel(level)
	if !istty {
		prefix = ""
		postfix = ""
	}

	dt := time.Now()
	year, month, day := dt.Date()
	hour, minute, second := dt.Clock()
	dateTime := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d.%06d", year, month, day,
		hour, minute, second,
		dt.Nanosecond()/1e3)

	logMsg := fmt.Sprintf(s, args...)
	fmt.Fprintf(os.Stderr, "%s[%s %s] %s%s\n", prefix, levelTxt, dateTime, logMsg, postfix)
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

func parseLevel(level int) (string, string) {
	switch level {
	case DEBUG:
		return blue, "DEBUG"
	case INFO:
		return green, "INFO"
	case WARNING:
		return yellow, "WARNING"
	case ERROR:
		return red, "ERROR"
	}
	return green, "INFO"
}

func ioctl(fd, request, argp uintptr) syscall.Errno {
	_, _, errorp := syscall.Syscall(syscall.SYS_IOCTL, fd, request, argp)
	return errorp
}

func isatty(f *os.File) bool {
	switch runtime.GOOS {
	case "darwin":
	case "linux":
	default:
		return false
	}
	var t [2]byte
	errno := ioctl(f.Fd(), syscall.TIOCGPGRP, uintptr(unsafe.Pointer(&t)))
	return errno == 0
}
