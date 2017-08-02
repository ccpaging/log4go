// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"sync"
)

var (
	// Default rotate cycle in seconds
	DefaultRotCycle int64 = 86400

	// Default rotate delay since midnight in seconds
	DefaultRotDelay0 int64 = 0

	// Default rotate max size
	DefaultRotSize int64 = 1024 * 1024 * 10
)

var DEBUG_ROTATE bool = false

// This log writer sends output to a file
type FileLogWriter struct {
	// The logging format
	format string

	// File header/footer
	header, footer string

	// 2nd cache, formatted message
	messages chan string

	// 3nd cache, bufio
	sync.RWMutex
	FileWriter

	rotate  int	   // Keep old logfiles (.001, .002, etc)
	maxsize int64  // Rotate at size
	cycle, delay0 int64  // Rotate cycle in seconds

	// write loop closed
	isRunLoop bool
	closedLoop chan struct{}
	resetLoop chan time.Time
}

func (f *FileLogWriter) Close() {
	close(f.messages)

	// wait for write Loop return
	if f.isRunLoop {  // Write loop may not running if no message write
		f.isRunLoop = false
		<- f.closedLoop
	}
}

// NewFileLogWriter creates a new LogWriter which writes to the given file and
// has rotation enabled if rotate > 0.
//
// If rotate > 0, rotate a new log file is opened, the old one is renamed
// with a .### extension to preserve it.  
// 
// The chainable Set* methods can be used to configure log rotation 
// based on cycle and size. Or by change Default* variables.
//
// The standard log-line format is:
//   [%D %T] [%L] (%S) %M
func NewFileLogWriter(fname string, rotate int) *FileLogWriter {
	f := &FileLogWriter{
		format:   FORMAT_DEFAULT,

		messages: make(chan string,  DefaultBufferLength),

		rotate:   rotate,
		cycle:	  DefaultRotCycle,
		delay0:	  DefaultRotDelay0,
		maxsize:  DefaultRotSize,

		isRunLoop: false,
		closedLoop: make(chan struct{}),
		resetLoop: make(chan time.Time, 5),
	}

	f.filename = fname; f.fileflush = DefaultFileFlush
	return f
}

// Get first rotate time
func (f *FileLogWriter) nextRotateTime() time.Time {
	nrt := time.Now()
	if f.delay0 < 0 {
		// Now + cycle
		nrt = nrt.Add(time.Duration(f.cycle) * time.Second)
	} else {
		// Tomorrow midnight (Clock 0) + delay0
		tomorrow := nrt.Add(24 * time.Hour)
		nrt = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 
						0, 0, 0, 0, tomorrow.Location())
		nrt = nrt.Add(time.Duration(f.delay0) * time.Second)
	}
	return nrt
}

func (f *FileLogWriter) writeLoop() {
	defer func() {
		f.isRunLoop = false
		close(f.closedLoop)
	}()

	if DEBUG_ROTATE { fmt.Println("Set cycle, delay0:", f.cycle, f.delay0) }

	nrt := f.nextRotateTime()
	var old_cycle int64 = f.cycle; var old_delay0 int64 = f.delay0

	timer := time.NewTimer(nrt.Sub(time.Now()))
	for {
		select {
		case msg, ok := <-f.messages:
			f.writeMessage(msg)
			if len(f.messages) <= 0 {
				f.FlushFile()
			}
			if !ok { // drain the log channel and write directly
				for msg := range f.messages {
					f.writeMessage(msg)
				}
				goto CLOSE
			}
		case <-timer.C:
			if DEBUG_ROTATE { fmt.Println("Get cycle, delay0:", f.cycle, f.delay0) }

			nrt = nrt.Add(time.Duration(f.cycle) * time.Second)
			timer.Reset(nrt.Sub(time.Now()))
			f.intRotate()
		case <-f.resetLoop:
			if old_cycle == f.cycle && old_delay0 == f.delay0 {
				continue
			}
			// Make sure cycle > 0
			if f.cycle < 2 {
				f.cycle = 86400
			}
			old_cycle = f.cycle; old_delay0 = f.delay0

			if DEBUG_ROTATE { fmt.Println("Reset cycle, delay0:", f.cycle, f.delay0) }

			nrt = f.nextRotateTime()
			timer.Reset(nrt.Sub(time.Now()))
		}
	}

CLOSE:
	f.Lock()
	f.CloseFile()
	f.Unlock()
}

func (f *FileLogWriter) writeMessage(msg string) {
	if msg == "" {
		return
	}
	
	f.Lock()
	defer f.Unlock()

	if len(f.header) > 0 {
		if n, _ := f.SeekFile(0, os.SEEK_CUR); n <= 0 {
			_, err := f.WriteString(FormatLogRecord(f.header, &LogRecord{Created: time.Now()}))
			if err != nil {
				fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", f.filename, err)
			}
		}
	}

	_, err := f.WriteString(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", f.filename, err)
		return
	}
}

func (f *FileLogWriter) LogWrite(rec *LogRecord) {
	if !f.isRunLoop {
		f.isRunLoop = true
		go f.writeLoop()
	}
	f.messages <- FormatLogRecord(f.format, rec)
}

func (f *FileLogWriter) intRotate() {
	f.Lock()
	defer f.Unlock()

	if n, _ := f.SeekFile(0, os.SEEK_CUR); n <= f.maxsize {
		return
	}
	
	// File existed and File size > maxsize
	
	if len(f.footer) > 0 { // Append footer
		f.WriteString(FormatLogRecord(f.footer, &LogRecord{Created: time.Now()}))
	}

	f.CloseFile() 

	if f.rotate <= 0 {
		os.Remove(f.filename)
		return
	}

	// File existed. File size > maxsize. Rotate
	newLog := f.filename + time.Now().Format(".20060102-150405")
	err := os.Rename(f.filename, newLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%q): Rename to %s. %v\n", f.filename, newLog, err)
		return
	}
	
	go FileRotate(f.filename, f.rotate, newLog)
}

// Set option. chainable
func (f *FileLogWriter) Set(name string, v interface{}) *FileLogWriter {
	f.SetOption(name, v)
	return f
}

// Set option. checkable. Must be set before the first log message is written.
func (f *FileLogWriter) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case "filename":
		if f.filename, ok = v.(string); !ok {
			return ErrBadValue
		}
		if len(f.filename) <= 0 {
			return ErrBadValue
		}
		err := os.MkdirAll(filepath.Dir(f.filename), DefaultFilePerm)
		if err != nil {
			return err
		}
		f.Lock()
		f.CloseFile()
		f.Unlock()
	case "flush":
		switch value := v.(type) {
		case int:
			f.fileflush = value
		case string:
			f.fileflush = StrToNumSuffix(strings.Trim(value, " \r\n"), 1024)
		default:
			return ErrBadValue
		}
		f.Lock()
		f.CloseFile()
		f.Unlock()
	case "rotate":
		switch value := v.(type) {
		case int:
			f.rotate = value
		case string:
			f.rotate = StrToNumSuffix(strings.Trim(value, " \r\n"), 1)
		default:
			return ErrBadValue
		}
	case "cycle":
		switch value := v.(type) {
		case int:
			f.cycle = int64(value)
		case int64:
			f.cycle = value
		case string:
			dur, _ := time.ParseDuration(value)
			f.cycle = int64(dur/time.Millisecond)
		default:
			return ErrBadValue
		}
		// Make sure cycle > 0
		if f.cycle < 2 {
			f.cycle = 86400
		}
		if f.isRunLoop {
			f.resetLoop <- time.Now()
		}
	case "delay0":
		switch value := v.(type) {
		case int:
			f.delay0 = int64(value)
		case int64:
			f.delay0 = value
		case string:
			dur, _ := time.ParseDuration(value)
			f.delay0 = int64(dur/time.Millisecond)
		default:
			return ErrBadValue
		}
		if f.isRunLoop {
			f.resetLoop <- time.Now()
		}
	case "maxsize":
		switch value := v.(type) {
		case int:
			f.maxsize = int64(value)
		case int64:
			f.maxsize = value
		case string:
			f.maxsize = int64(StrToNumSuffix(strings.Trim(value, " \r\n"), 1024))
		default:
			return ErrBadValue
		}
	case "format":
		if f.format, ok = v.(string); !ok {
			return ErrBadValue
		}
	case "head":
		if f.header, ok = v.(string); !ok {
			return ErrBadValue
		}
	case "foot":
		if f.footer, ok = v.(string); !ok {
			return ErrBadValue
		}
	default:
		return ErrBadOption
	}
	return nil
}

/* Not using now
func (w *FileLogWriter) GetOption(name string) (interface{}, error) {
	switch name {
	case "filename":
		return f.filename, nil
	case "flush":
		return f.flush, nil
	default:
		return nil, ErrBadOption
	}
}
*/