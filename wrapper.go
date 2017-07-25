// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"errors"
	"fmt"
	"os"
	"io"
	"strings"
)

var (
	Global Logger
)

func init() {
	Global = Logger {
		"stdout": NewFilter(DEBUG, NewConsoleLogWriter().SetFormat("%T %L %s %M")),
	}
}

func GetGlobalLogger() Logger {
	return Global
}

// Wrapper for (*Logger).LoadConfiguration
func LoadConfiguration(filename string) {
	Global.LoadConfiguration(filename)
}

// Wrapper for (*Logger).LoadConfigBuf
func LoadConfigBuf(buf []byte) {
	Global.LoadConfigBuf(buf)
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
// New creates a new Logger. The out variable sets the
// destination to which log data will be written.
// The prefix appears at the beginning of each generated log line.
// The flag argument defines the logging properties.
func New(out io.Writer, prefix string, flag int) *Logger {
	return &Logger {
			"stdout": NewFilter(DEBUG, NewOutLogWriter(out, prefix, flag)),
		}
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

// These functions Compatibility with `log`
// Write to the standard logger.

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is the count of the number of
// frames to skip when computing the file name and line number
// if Llongfile or Lshortfile is set; a value of 1 will print the details
// for the caller of Output.
func Output(calldepth int, s string) error {
	Global.CompatOutput(INFO, calldepth, s) // +1 for this frame.
	return nil
}


// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Print(v ...interface{}) {
	Global.CompatOutput(INFO, DefaultCallerSkip, fmt.Sprint(v...))
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	Global.CompatOutput(INFO, DefaultCallerSkip, fmt.Sprintf(format, v...))
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...interface{}) {
	Global.CompatOutput(INFO, DefaultCallerSkip, fmt.Sprintln(v...))
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	Global.CompatOutput(ERROR, DefaultCallerSkip, fmt.Sprint(v...))
	Global.Close()
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	Global.CompatOutput(ERROR, DefaultCallerSkip, fmt.Sprintf(format, v...))
	Global.Close()
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Fatalln(v ...interface{}) {
	Global.CompatOutput(ERROR, DefaultCallerSkip, fmt.Sprintln(v...))
	Global.Close()
	os.Exit(1)
}

// Panic is equivalent to Print() followed by a call to panic().
func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	Global.CompatOutput(CRITICAL, DefaultCallerSkip, s)
	Global.Close()
	panic(s)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	Global.CompatOutput(CRITICAL, DefaultCallerSkip, s)
	Global.Close()
	panic(s)
}

// Panicln is equivalent to Println() followed by a call to panic().
func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	Global.CompatOutput(CRITICAL, DefaultCallerSkip, s)
	Global.Close()
	panic(s)
}
