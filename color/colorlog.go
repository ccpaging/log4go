// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package colorlog

import (
	"fmt"
	"io"
	"os"

	l4g "github.com/ccpaging/log4go"
	"github.com/ccpaging/go-colortext"
)

var stdout io.Writer = os.Stdout

// This is the standard writer that prints to standard output.
type ColorLogWriter struct {
	iow		io.Writer
	format 	string
}

// This creates a new ConsoleLogWriter
func NewLogWriter() *ColorLogWriter {
	c := &ColorLogWriter{
		iow:	stdout,
		format: "[%T %D %Z] [%L] (%S) %M",
	}
	return c
}

// Set the logging format (chainable).  Must be called before the first log
// message is written.
func (c *ColorLogWriter) SetFormat(format string) *ColorLogWriter {
	c.format = format
	return c
}

func (c *ColorLogWriter) Close() {
}

func (c *ColorLogWriter) LogWrite(rec *l4g.LogRecord) {
	switch rec.Level {
		case l4g.CRITICAL:
			ct.ChangeColor(ct.Red, true, ct.White, false)
		case l4g.ERROR:
			ct.ChangeColor(ct.Red, false, 0, false)
		case l4g.WARNING:
			ct.ChangeColor(ct.Yellow, false, 0, false)
		case l4g.INFO:
			ct.ChangeColor(ct.Green, false, 0, false)
		case l4g.DEBUG:
			ct.ChangeColor(ct.Magenta, false, 0, false)
		case l4g.TRACE:
			ct.ChangeColor(ct.Cyan, false, 0, false)
		default:
	}
	defer ct.ResetColor()
	fmt.Fprint(c.iow, l4g.FormatLogRecord(c.format, rec))
}

// Set option. chainable
func (c *ColorLogWriter) Set(name string, v interface{}) *ColorLogWriter {
	c.SetOption(name, v)
	return c
}

// Set option. checkable
func (c *ColorLogWriter) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case "format":
		if c.format, ok = v.(string); !ok {
			return l4g.ErrBadValue
		}
		return nil
	default:
		return l4g.ErrBadOption
	}
}

// Get option. checkable
func (c *ColorLogWriter) GetOption(name string) (interface{}, error) {
	switch name {
	case "format":
		return c.format, nil
	default:
		return nil, l4g.ErrBadOption
	}
}
