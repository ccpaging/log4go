// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"io"
	"os"
	"sync"
	"github.com/daviddengcn/go-colortext"
)

var stdout io.Writer = os.Stdout
const DefaultTimeFormat string = "15:04:05 MST 2006/01/02"

// This is the standard writer that prints to standard output.
type ConsoleLogWriter struct {
	rec chan *LogRecord
	closing bool
    wg *sync.WaitGroup
	color bool
	longformat bool
	timeformat string
}

// This is the ConsoleLogWriter's output method.  This will block if the output
// buffer is full.
func (w *ConsoleLogWriter) LogWrite(rec *LogRecord) {
	if w.closing {
		fmt.Fprintf(os.Stderr, "ConsoleLogWriter: channel has been closed. Message is [%s]\n", rec.Message)
		return
	}
	w.rec <- rec
}

// Close stops the logger from sending messages to standard output.  Attempts to
// send log messages to this logger after a Close have undefined behavior.
func (w *ConsoleLogWriter) Close() {
	w.closing = true
	close(w.rec)
    w.wg.Wait()
}

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter() *ConsoleLogWriter {
	w := &ConsoleLogWriter{
		rec:  	make(chan *LogRecord, LogBufferLength),
		closing: 	false,
        wg: 	&sync.WaitGroup{},	
		color:	true,
		longformat: true,
		timeformat:	DefaultTimeFormat,
	}

    w.wg.Add(1)
	go w.run(stdout)
	return w
}

// for test only
func NewOutConsoleLogWriter(out io.Writer) *ConsoleLogWriter {
	w := &ConsoleLogWriter{
		rec:  	make(chan *LogRecord, LogBufferLength),
		color:	true,
		longformat: true,
		timeformat:	DefaultTimeFormat,
	}

	go w.run(out)
	return w
}

// Must be called before the first log message is written.
func (w *ConsoleLogWriter) SetColor(color bool) *ConsoleLogWriter {
	w.color = color
	return w
}

// Must be called before the first log message is written.
func (w *ConsoleLogWriter) SetLongFormat(longformat bool) *ConsoleLogWriter {
	w.longformat = longformat
	return w
}

// Must be called before the first log message is written.
func (w *ConsoleLogWriter) SetTimeFormat(timeformat string) *ConsoleLogWriter {
	w.timeformat = timeformat
	return w
}

func (w *ConsoleLogWriter) run(out io.Writer) {
    defer w.wg.Done()

	var timestr string
	var timestrAt int64

	for {
		rec, ok := <-w.rec
		if !ok {
			return
		}
		if w.color {
			switch rec.Level {
				case CRITICAL:
					ct.ChangeColor(ct.Red, true, ct.White, false)
				case ERROR:
					ct.ChangeColor(ct.Red, false, 0, false)
				case WARNING:
					ct.ChangeColor(ct.Yellow, false, 0, false)
				case INFO:
					ct.ChangeColor(ct.Green, false, 0, false)
				case DEBUG:
					ct.ChangeColor(ct.Magenta, false, 0, false)
				case TRACE:
					ct.ChangeColor(ct.Cyan, false, 0, false)
				default:
			}
		}
		if !w.longformat {
			fmt.Fprint(out, rec.Message)
		} else {
			if at := rec.Created.UnixNano() / 1e9; at != timestrAt {
				timestr, timestrAt = rec.Created.Format(w.timeformat), at
			}
			fmt.Fprint(out, "[", timestr, "] [", levelStrings[rec.Level], "] [", rec.Source, "] ", rec.Message)
		}
		if w.color {
			ct.ResetColor()
		}
		fmt.Fprint(out, "\n")
	}
}

