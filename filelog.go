// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"bufio"
	"io"
	"sync"
	"runtime"
)

var (
	// Default filename. Set by init
	DefaultFileName = ""

	// Default log file and directory perm
	DefaultFilePerm = os.FileMode(0660)

	// Default flush size of cache writing file
	DefaultFileFlush = 4096

	// Default rotate cycle in seconds
	DefaultRotCycle int64 = 86400

	// Default rotate delay since midnight in seconds
	DefaultRotDelay0 int64 = 0

	// Default rotate max size
	DefaultRotSize int64 = 1024 * 1024 * 10
)

var DEBUG_ROTATE bool = false


func init() {
	base := filepath.Base(os.Args[0])
	ext := filepath.Ext(base)
	DefaultFileName = strings.TrimSuffix(base, ext) + ".log"
	if runtime.GOOS != "windows" {
		DefaultFileName = "~/log/" + DefaultFileName
	}
}

// This log writer sends output to a file
type FileLogWriter struct {
	// The opened file
	filename string
	file   *os.File

	// The logging format
	format string

	// File header/footer
	header, footer string

	// 2nd cache, formatted message
	messages chan string

	// 3nd cache, bufio
	sync.RWMutex
	flush  int
	bufWriter *bufio.Writer
	writer io.Writer

	rotate  int	   // Keep old logfiles (.001, .002, etc)
	maxsize int64  // Rotate at size
	cycle, delay0 int64  // Rotate cycle in seconds

	// write loop closed
	isRunLoop bool
	closedLoop chan struct{}
	resetLoop chan time.Time
}

func (w *FileLogWriter) Close() {
	close(w.messages)

	// wait for write Loop return
	if w.isRunLoop {  // Write loop may not running if no message write
		w.isRunLoop = false
		<- w.closedLoop
	}
}

func (w *FileLogWriter) fileOpen(flag int) *os.File {
	fd, err := os.OpenFile(w.filename, flag, DefaultFilePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
		return nil
	}

	w.file = fd
	w.writer = w.file

	if w.flush > 0 {
		w.bufWriter = bufio.NewWriterSize(w.file, w.flush)
		w.writer = w.bufWriter
	}
	return fd
}

func (w *FileLogWriter) fileClose() {
	if w.file == nil {
		return
	}

	if w.bufWriter != nil {
		w.bufWriter.Flush()
	} else {
		w.file.Sync()
	}
	w.file.Close()

	w.file = nil
	w.writer = nil
	w.bufWriter = nil
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
	w := &FileLogWriter{
		filename: fname,
		format:   FORMAT_DEFAULT,

		messages: make(chan string,  DefaultBufferLength),

		flush:	  DefaultFileFlush,
		bufWriter: nil,

		rotate:   rotate,
		cycle:	  DefaultRotCycle,
		delay0:	  DefaultRotDelay0,
		maxsize:  DefaultRotSize,

		isRunLoop: false,
		closedLoop: make(chan struct{}),
		resetLoop: make(chan time.Time, 5),
	}
	if w.filename == "" {
		w.filename = DefaultFileName
	}
	w.resetLoop <- time.Now()
	return w
}

// Get first rotate time
func (w *FileLogWriter) nextRotateTime() time.Time {
	nrt := time.Now()
	if w.delay0 < 0 {
		// Now + cycle
		nrt = nrt.Add(time.Duration(w.cycle) * time.Second)
	} else {
		// Tomorrow midnight (Clock 0) + delay0
		tomorrow := nrt.Add(24 * time.Hour)
		nrt = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 
						0, 0, 0, 0, tomorrow.Location())
		nrt = nrt.Add(time.Duration(w.delay0) * time.Second)
	}
	return nrt
}

func (w *FileLogWriter) writeLoop() {
	defer func() {
		w.isRunLoop = false
		close(w.closedLoop)
	}()

	if DEBUG_ROTATE { fmt.Println("Set cycle, delay0:", w.cycle, w.delay0) }

	var old_cycle int64 = -1; var old_delay0 int64 = -1

	nrt := w.nextRotateTime()
	timer := time.NewTimer(nrt.Sub(time.Now()))
	for {
		select {
		case msg, ok := <-w.messages:
			if msg != "" {
				w.writeMessage(msg)
			}
			if w.bufWriter != nil && len(w.messages) <= 0 {
				w.bufWriter.Flush()
			}
			if !ok { // drain the log channel and write directly
				for msg := range w.messages {
					w.writeMessage(msg)
				}
				goto CLOSE
			}
		case <-timer.C:
			if DEBUG_ROTATE { fmt.Println("Get cycle, delay0:", w.cycle, w.delay0) }

			nrt = nrt.Add(time.Duration(w.cycle) * time.Second)
			timer.Reset(nrt.Sub(time.Now()))
			w.intRotate()
		case <-w.resetLoop:
			if old_cycle == w.cycle && old_delay0 == w.delay0 {
				continue
			}
			// Make sure cycle > 0
			if w.cycle < 2 {
				w.cycle = 86400
			}
			old_cycle = w.cycle; old_delay0 = w.delay0

			if DEBUG_ROTATE { fmt.Println("Reset cycle, delay0:", w.cycle, w.delay0) }

			nrt = w.nextRotateTime()
			timer.Reset(nrt.Sub(time.Now()))
		}
	}

CLOSE:
	w.Lock()
	w.fileClose()
	w.Unlock()
}

func (w *FileLogWriter) writeMessage(msg string) {
	w.Lock()
	defer w.Unlock()

	// Open file when write first message
	if w.file == nil {
		isNewFile := true
		if fi, err := os.Lstat(w.filename); err == nil && fi.Size() > 0 {
			isNewFile = false 
		}
		fd := w.fileOpen(os.O_WRONLY|os.O_APPEND|os.O_CREATE)
		if fd == nil {
			return
		}
		if isNewFile { // write header
			fmt.Fprint(w.writer, FormatLogRecord(w.header, &LogRecord{Created: time.Now()}))
		}
	}

	// Perform the write
	_, err := fmt.Fprint(w.writer, msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.filename, err)
		return
	}
}

func (w *FileLogWriter) LogWrite(rec *LogRecord) {
	if !w.isRunLoop {
		w.isRunLoop = true
		go w.writeLoop()
	}
	w.messages <- FormatLogRecord(w.format, rec)
}

func (w *FileLogWriter) intRotate() {
	w.Lock()
	defer w.Unlock()

	w.fileClose() 

	fi, err := os.Lstat(w.filename)
	if err != nil { // File not exist. Create new.
		return
	}

	if fi.Size() < w.maxsize { // File exist and size normal
		return
	}

	// File existed. File size > maxsize
	if w.rotate <= 0 {
		os.Remove(w.filename)
		return
	}

	// Append footer
	fd, _ := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND, DefaultFilePerm)
	if fd != nil {
		fmt.Fprint(fd, FormatLogRecord(w.footer, &LogRecord{Created: time.Now()}))
		fd.Sync()
		fd.Close()
	}

	// File existed. File size > maxsize. Rotate
	newLog := w.filename + time.Now().Format(".20060102-150405")
	err = os.Rename(w.filename, newLog)
	
	if DEBUG_ROTATE { fmt.Println(w.filename, "Rename", newLog, err) }
	
	// May compress new log file here

	go func() {
		ext := filepath.Ext(w.filename) // like ".log"
		base := strings.TrimSuffix(w.filename, ext) // include dir
		
		if DEBUG_ROTATE { fmt.Println(w.rotate, base, ext) }
	
		// May create old directory here
	
		var n int
		var err error = nil 
		slot := ""
		for n = 1; n <= w.rotate; n++ {
			slot = base + fmt.Sprintf(".%03d", n) + ext
			_, err = os.Lstat(slot)
			if err != nil {
				break
			}
		}
	
		if DEBUG_ROTATE { fmt.Println(slot) }

		if err == nil { // Full
			fmt.Println("Remove:", slot)
			os.Remove(slot)
			n--
		}
	
		for ; n > 1; n-- {
			prev := base + fmt.Sprintf(".%03d", n - 1) + ext

			if DEBUG_ROTATE { fmt.Println(prev, "Rename", slot) }

			os.Rename(prev, slot)
			slot = prev
		}
		
		if DEBUG_ROTATE { fmt.Println(newLog, "Rename", base + ".001" + ext) }

		os.Rename(newLog, base + ".001" + ext)
	}()
}

// Set option. chainable
func (w *FileLogWriter) Set(name string, v interface{}) *FileLogWriter {
	w.SetOption(name, v)
	return w
}

// Set option. checkable. Must be set before the first log message is written.
func (w *FileLogWriter) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case "filename":
		if w.filename, ok = v.(string); !ok {
			return ErrBadValue
		}
		if len(w.filename) <= 0 {
			return ErrBadValue
		}
		err := os.MkdirAll(filepath.Dir(w.filename), DefaultFilePerm)
		if err != nil {
			return err
		}
	case "flush":
		switch value := v.(type) {
		case int:
			w.flush = value
		case string:
			w.flush = StrToNumSuffix(strings.Trim(value, " \r\n"), 1024)
		default:
			return ErrBadValue
		}
		w.Lock()
		w.fileClose()
		w.Unlock()
	case "rotate":
		switch value := v.(type) {
		case int:
			w.rotate = value
		case string:
			w.rotate = StrToNumSuffix(strings.Trim(value, " \r\n"), 1)
		default:
			return ErrBadValue
		}
	case "cycle":
		switch value := v.(type) {
		case int:
			w.cycle = int64(value)
		case int64:
			w.cycle = value
		case string:
			dur, _ := time.ParseDuration(value)
			w.cycle = int64(dur/time.Millisecond)
		default:
			return ErrBadValue
		}
		// Make sure cycle > 0
		if w.cycle < 2 {
			w.cycle = 86400
		}
		if w.isRunLoop {
			w.resetLoop <- time.Now()
		}
	case "delay0":
		switch value := v.(type) {
		case int:
			w.delay0 = int64(value)
		case int64:
			w.delay0 = value
		case string:
			dur, _ := time.ParseDuration(value)
			w.delay0 = int64(dur/time.Millisecond)
		default:
			return ErrBadValue
		}
		if w.isRunLoop {
			w.resetLoop <- time.Now()
		}
	case "maxsize":
		switch value := v.(type) {
		case int:
			w.maxsize = int64(value)
		case int64:
			w.maxsize = value
		case string:
			w.maxsize = int64(StrToNumSuffix(strings.Trim(value, " \r\n"), 1024))
		default:
			return ErrBadValue
		}
	case "format":
		if w.format, ok = v.(string); !ok {
			return ErrBadValue
		}
	case "head":
		if w.header, ok = v.(string); !ok {
			return ErrBadValue
		}
	case "foot":
		if w.footer, ok = v.(string); !ok {
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
		return w.filename, nil
	case "flush":
		return w.flush, nil
	case "cycle":
		return w.cycle, nil
	case "delay0":
		return w.delay0, nil
	case "rotate":
		return w.rotate, nil
	case "maxsize":
		return w.maxsize, nil
	case "format":
		return w.format, nil
	case "head":
		return w.header, nil
	case "foot":
		return w.footer, nil
	default:
		return nil, ErrBadOption
	}
}
*/