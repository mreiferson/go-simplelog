// simplelog is Golang package to replace the standard library's log adding logging
// level, colors, and an easy to read format modeled after Tornado (http://tornadoweb.org/)
// 
// It is and designed to be usable out of the box with no dependencies.
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

// Logger is the basic type if you want to maintain multiple instances
// with a different loggin level.  In most cases just use the public (global)
// functions.
type Logger struct {
	sync.Mutex
	level int
}

// NewLogger creates a new Logger instance with the specified initial log level
func NewLogger(level int) *Logger {
	return &Logger{level: level}
}

// SetLevel takes either a string of int specifying the the new logging level
//
// The string form is useful for easily passing command line parameters, ie:
//
//     var logLevel = flag.String("logging", "info", "log level")
//     ...
//     simplelog.SetLevel(logLevel)
//
// Valid levels (string = int):
//
//     DEBUG   = 0
//     INFO    = 1
//     WARNING = 2
//     ERROR   = 3
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

// Log formats the message with the supplied arguments to fmt.Sprintf, applies
// color based on log level, and prints to os.Stderr
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

// SetLevel sets the logging level for the default (global) logger
func SetLevel(lvl interface{}) {
	defaultLogger.SetLevel(lvl)
}

// Debug is a convenience method to log a DEBUG message on the default (global) logger
func Debug(s string, args ...interface{}) {
	defaultLogger.Log(DEBUG, s, args...)
}

// Info is a convenience method to log an INFO message on the default (global) logger
func Info(s string, args ...interface{}) {
	defaultLogger.Log(INFO, s, args...)
}

// Warning is a convenience method to log a WARNING message on the default (global) logger
func Warning(s string, args ...interface{}) {
	defaultLogger.Log(WARNING, s, args...)
}

// Error is a convenience method to log an ERROR message on the default (global) logger
func Error(s string, args ...interface{}) {
	defaultLogger.Log(ERROR, s, args...)
}

// Log is a convenience method to log a message on the default (global) logger for any level
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
