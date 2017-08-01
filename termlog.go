// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"io"
	"os"
)

var stdout io.Writer = os.Stdout

// This is the standard writer that prints to standard output.
type ConsoleLogWriter struct {
	out		io.Writer
	format 	string
	prefix  string
}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() *ConsoleLogWriter {
	c := &ConsoleLogWriter{
		out:	stdout,
		format: FORMAT_DEFAULT,
	}
	return c
}

func (c *ConsoleLogWriter) Close() {
}

func (c *ConsoleLogWriter) LogWrite(rec *LogRecord) {
	fmt.Fprint(c.out, c.prefix + FormatLogRecord(c.format, rec))
}

// Set option. chainable
func (c *ConsoleLogWriter) Set(name string, v interface{}) *ConsoleLogWriter {
	c.SetOption(name, v)
	return c
}

// Set option. checkable
func (c *ConsoleLogWriter) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case "format":
		if c.format, ok = v.(string); !ok {
			return ErrBadValue
		}
		return nil
	default:
		return ErrBadOption
	}
}

// Get option. checkable
func (c *ConsoleLogWriter) GetOption(name string) (interface{}, error) {
	switch name {
	case "format":
		return c.format, nil
	default:
		return nil, ErrBadOption
	}
}
