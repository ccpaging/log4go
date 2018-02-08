// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"io"
	"os"
)

var stdout io.Writer = os.Stdout

var isColorful = (os.Getenv("TERM") != "" && os.Getenv("TERM") != "dumb") ||
	 os.Getenv("ConEmuANSI") == "ON"

// 0, Black; 1, Red; 2, Green; 3, Yellow; 4, Blue; 5, Purple; 6, Cyan; 7, White
var ColorBytes = [...][]byte{
	[]byte("\x1b[0;34m"),	   // FINEST, Blue
	[]byte("\x1b[0;36m"),	   // FINE, Cyan
	[]byte("\x1b[0;32m"),	   // DEBUG, Green
	[]byte("\x1b[0;35m"), 	   // TRACE, Purple
 	nil,					   // INFO, Default
 	[]byte("\x1b[1;33m"), 	   // WARNING, Yellow
 	[]byte("\x1b[0;31m"), 	   // ERROR, Red
 	[]byte("\x1b[0;31m;47m"),  // CRITICAL, Red - White
}
var ColorReset = []byte("\x1b[0m")

// This is the standard writer that prints to standard output.
type ConsoleLogWriter struct {
	out		io.Writer
	color 	bool	
	format 	string
}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() *ConsoleLogWriter {
	c := &ConsoleLogWriter{
		out:	stdout,
		color:	false,
		format: "[%T %D %Z] [%L] (%S) %M",
	}
	return c
}

// Must be called before the first log message is written.
func (c *ConsoleLogWriter) SetColor(color bool) *ConsoleLogWriter {
	c.color = color
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
	if c.color {
		c.out.Write(ColorBytes[rec.Level])
		defer c.out.Write(ColorReset)
	}
	fmt.Fprint(c.out, FormatLogRecord(c.format, rec))
}
