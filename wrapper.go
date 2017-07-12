// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"runtime"
	"path/filepath"
)

var (
	Global Logger
)

func init() {
	Global = Logger{
		"stdout": NewFilter(DEBUG, NewConsoleLogWriter().SetColor(true).SetFormat("%T %L %s %M")),
	}
}

// Wrapper for (*Logger).LoadConfiguration
func LoadConfiguration(filename string) {
	Global.LoadConfig(filename)
}

func LoadConfigBuf(filename string, buf []byte) {
	Global.LoadConfigBuf(filename, buf)
}

// Wrapper for (*Logger).AddFilter
func AddFilter(name string, lvl Level, writer LogWriter) {
	Global.AddFilter(name, lvl, writer)
}

// Wrapper for (*Logger).Close (closes and removes all logwriters)
func Close() {
	Global.Close()
}

// Compatibility with `log`
func compat(lvl Level, calldepth int, args ...interface{}) {
	// Determine caller func
	pc, _, lineno, ok := runtime.Caller(calldepth)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%d", filepath.Base(runtime.FuncForPC(pc).Name()), lineno)
	}

	msg := ""
	if len(args) > 0 {
		msg = fmt.Sprintf(strings.Repeat(" %v", len(args))[1:], args...)
	}
	msg = strings.TrimRight(msg, "\r\n")

	Global.Log(lvl, src, msg)
	if lvl == ERROR {
		Global.Close()
		os.Exit(0)
	} else if lvl == CRITICAL {
		Global.Close()
		panic(msg)
	}
}

func compatf(lvl Level, calldepth int, format string, args ...interface{}) {
	// Determine caller func
	pc, _, lineno, ok := runtime.Caller(calldepth)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%d", filepath.Base(runtime.FuncForPC(pc).Name()), lineno)
	}

	msg := fmt.Sprintf(format, args...)
	msg = strings.TrimRight(msg, "\r\n")

	Global.Log(lvl, src, msg)
	if lvl == ERROR {
		Global.Close()
		os.Exit(0)
	} else if lvl == CRITICAL {
		Global.Close()
		panic(msg)
	}
}

func Crash(args ...interface{}) {
	compat(CRITICAL, DefaultCallerSkip, args ...)
}

// Logs the given message and crashes the program
func Crashf(format string, args ...interface{}) {
	compatf(CRITICAL, DefaultCallerSkip, format, args ...)
}

// Compatibility with `log`
func Exit(args ...interface{}) {
	compat(ERROR, DefaultCallerSkip, args ...)
}

// Compatibility with `log`
func Exitf(format string, args ...interface{}) {
	compatf(ERROR, DefaultCallerSkip, format, args ...)
}

// Compatibility with `log`
func Stderr(args ...interface{}) {
	compat(WARNING, DefaultCallerSkip, args ...)
}

// Compatibility with `log`
func Stderrf(format string, args ...interface{}) {
	compatf(WARNING, DefaultCallerSkip, format, args ...)
}

// Compatibility with `log`
func Stdout(args ...interface{}) {
	compat(INFO, DefaultCallerSkip, args ...)
}

// Compatibility with `log`
func Stdoutf(format string, args ...interface{}) {
	compatf(INFO, DefaultCallerSkip, format, args ...)
}

// Compatibility with `log`
func Fatal(v ...interface{}) {
	compat(ERROR, DefaultCallerSkip, v ...)
}

func Fatalf(format string, v ...interface{}) {
	compatf(ERROR, DefaultCallerSkip, format, v ...)
}

func Fatalln(v ...interface{}) {
	compat(ERROR, DefaultCallerSkip, v ...)
}

func Output(calldepth int, s string) error {
	compat(INFO, calldepth, s)
	return nil
}

func Panic(v ...interface{}) {
	compat(CRITICAL, DefaultCallerSkip, v ...)
}

func Panicf(format string, v ...interface{}) {
	compatf(CRITICAL, DefaultCallerSkip, format, v ...)
}

func Panicln(v ...interface{}) {
	compat(CRITICAL, DefaultCallerSkip, v ...)
}

func Print(v ...interface{}) {
	compat(INFO, DefaultCallerSkip, v ...)
}

func Printf(format string, v ...interface{}) {
	compatf(INFO, DefaultCallerSkip, format, v ...)
}

func Println(v ...interface{}) {
	compat(INFO, DefaultCallerSkip, v ...)
}

// Send a log message manually
// Wrapper for (*Logger).Log
func Log(lvl Level, source, message string) {
	Global.Log(lvl, source, message)
}

// Send a formatted log message easily
// Wrapper for (*Logger).Logf
func Logf(lvl Level, format string, args ...interface{}) {
	Global.intLogf(lvl, format, args...)
}

// Send a closure log message
// Wrapper for (*Logger).Logc
func Logc(lvl Level, closure func() string) {
	Global.intLogc(lvl, closure)
}

// Utility for finest log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Finest
func Finest(arg0 interface{}, args ...interface{}) {
	const (
		lvl = FINEST
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for fine log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Fine
func Fine(arg0 interface{}, args ...interface{}) {
	const (
		lvl = FINE
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for debug log messages
// When given a string as the first argument, this behaves like Logf but with the DEBUG log level (e.g. the first argument is interpreted as a format for the latter arguments)
// When given a closure of type func()string, this logs the string returned by the closure iff it will be logged.  The closure runs at most one time.
// When given anything else, the log message will be each of the arguments formatted with %v and separated by spaces (ala Sprint).
// Wrapper for (*Logger).Debug
func Debug(arg0 interface{}, args ...interface{}) {
	const (
		lvl = DEBUG
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for trace log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Trace
func Trace(arg0 interface{}, args ...interface{}) {
	const (
		lvl = TRACE
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for info log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Info
func Info(arg0 interface{}, args ...interface{}) {
	const (
		lvl = INFO
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for warn log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Warn
func Warn(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = WARNING
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
		return errors.New(fmt.Sprintf(first, args...))
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogf(lvl, "%s", str)
		return errors.New(str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
		return errors.New(fmt.Sprint(first) + fmt.Sprintf(strings.Repeat(" %v", len(args)), args...))
	}
	return nil
}

// Utility for error log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Error
func Error(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = ERROR
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
		return errors.New(fmt.Sprintf(first, args...))
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogf(lvl, "%s", str)
		return errors.New(str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
		return errors.New(fmt.Sprint(first) + fmt.Sprintf(strings.Repeat(" %v", len(args)), args...))
	}
	return nil
}

// Utility for critical log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Critical
func Critical(arg0 interface{}, args ...interface{}) error {
	const (
		lvl = CRITICAL
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
		return errors.New(fmt.Sprintf(first, args...))
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogf(lvl, "%s", str)
		return errors.New(str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
		return errors.New(fmt.Sprint(first) + fmt.Sprintf(strings.Repeat(" %v", len(args)), args...))
	}
	return nil
}
