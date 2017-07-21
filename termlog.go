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
	iow		io.Writer
	format 	string
}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() *ConsoleLogWriter {
	c := &ConsoleLogWriter{
		iow:	stdout,
		format: "[%T %D %Z] [%L] (%S) %M",
	}
	return c
}

// Set the logging format (chainable).  Must be called before the first log
// message is written.
func (c *ConsoleLogWriter) SetFormat(format string) *ConsoleLogWriter {
	c.format = format
	return c
}

func (c *ConsoleLogWriter) Close() {
}

func (c *ConsoleLogWriter) LogWrite(rec *LogRecord) {
	fmt.Fprint(c.iow, FormatLogRecord(c.format, rec))
}
