// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
	"bufio"
	"io"
	"sync"
)

// This log writer sends output to a file
type CacheFileLogWriter struct {
	// The opened file
	filename string
	file   *os.File

	// The logging format
	format string

	// File header/trailer
	header, trailer string

	// 2nd cache, formatted message
	messages chan string
	closedWriteLoop chan struct{} // write loop closed

	// 3nd cache, bufio
	sync.RWMutex
	flush  int
	bufWriter *bufio.Writer
	writer io.Writer

	// Keep old logfiles (.001, .002, etc)
	rotate int
	cycle  int64  // criterium in seconds
	delay0  int64  // start rotating work at clock 3am = 10800
	// Rotate at size
	maxsize int64
}

func (w *CacheFileLogWriter) Close() {
	close(w.messages)
	// wait for writeLoop return
	<- w.closedWriteLoop
}

func (w *CacheFileLogWriter) fileOpen(flag int) *os.File {
	fd, err := os.OpenFile(w.filename, flag, DefaultFilePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CacheFileLogWriter(%q): %s\n", w.filename, err)
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

func (w *CacheFileLogWriter) fileClose() {
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

// NewCacheFileLogWriter creates a new LogWriter which writes to the given file and
// has rotation enabled if rotate > 0.
//
// If rotate > 0, any time a new log file is opened, the old one is renamed
// with a .### extension to preserve it.  The various Set* methods can be used
// to configure log rotation based on size, and cycle.
//
// The standard log-line format is:
//   [%D %T] [%L] (%S) %M
func NewCacheFileLogWriter(fname string, rotate int) *CacheFileLogWriter {
    err := os.MkdirAll(path.Dir(fname), DefaultFilePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CacheFileLogWriter(%s): %s\n", fname, err)
		return nil
	}
	w := &CacheFileLogWriter{
		filename: fname,
		format:   "[%D %z %T] [%L] (%S) %M",

		messages: make(chan string,  DefaultBufferLength),
		closedWriteLoop: make(chan struct{}),

		flush:	  DefaultFileFlush,
		bufWriter: nil,

		rotate:   rotate,
	}

	go w.writeLoop()
	return w
}

func (w *CacheFileLogWriter) writeLoop() {
	defer close(w.closedWriteLoop)

	nrt := time.Now()
	if w.delay0 < 0 {
		// Now + cycle
		nrt = nrt.Add(time.Duration(w.cycle) * time.Second)
	} else {
		// tomorrow midnight (Clock 0) + delay0
		tomorrow := nrt.Add(24 * time.Hour)
        nrt = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 
						0, 0, 0, 0, tomorrow.Location())
		nrt = nrt.Add(time.Duration(w.delay0) * time.Second)
	}
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
			nrt = nrt.Add(time.Duration(w.cycle) * time.Second)
			timer.Reset(nrt.Sub(time.Now()))
			w.intRotate()
		}
	}

CLOSE:
	w.Lock()
	w.fileClose()
	w.Unlock()
}

func (w *CacheFileLogWriter) writeMessage(msg string) {
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
		fmt.Fprintf(os.Stderr, "CacheFileLogWriter(%q): %s\n", w.filename, err)
		return
	}
}

func (w *CacheFileLogWriter) LogWrite(rec *LogRecord) {
	w.messages <- FormatLogRecord(w.format, rec)
}

func (w *CacheFileLogWriter) intRotate() {
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

	// Append trailer
	fd, _ := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND, DefaultFilePerm)
	if fd != nil {
		fmt.Fprint(fd, FormatLogRecord(w.trailer, &LogRecord{Created: time.Now()}))
		fd.Sync()
		fd.Close()
	}

	// File existed. File size > maxsize. Rotate
	newLog := w.filename + time.Now().Format(".20060102-150405")
	err = os.Rename(w.filename, newLog)
	fmt.Println(w.filename, "Rename", newLog, err)
	// May replace with compress 

	go func() {
		ext := path.Ext(w.filename) // like ".log"
		base := strings.TrimSuffix(w.filename, ext) // include dir
		fmt.Println(w.rotate, base, ext)
	
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
	
		fmt.Println(slot)
		if err == nil { // Full
			fmt.Println("Remove:", slot)
			os.Remove(slot)
			n--
		}
	
		for ; n > 1; n-- {
			prev := base + fmt.Sprintf(".%03d", n - 1) + ext
			fmt.Println(prev, "Rename", slot)
			os.Rename(prev, slot)
			slot = prev
		}
		
		fmt.Println(newLog, "Rename", base + ".001" + ext)
		os.Rename(newLog, base + ".001" + ext)
	}()
}

// Set the logging format (chainable).  Must be called before the first log
// message is written.
func (w *CacheFileLogWriter) SetFormat(format string) *CacheFileLogWriter {
	w.format = format
	return w
}

// Set the logfile header and footer (chainable).  Must be called before the first log
// message is written.  These are formatted similar to the FormatLogRecord (e.g.
// you can use %D and %T in your header/footer for date and time).
func (w *CacheFileLogWriter) SetHeadFoot(head, foot string) *CacheFileLogWriter {
	w.header, w.trailer = head, foot
	return w
}

func (w *CacheFileLogWriter) SetFlush(flush int) *CacheFileLogWriter {
	w.Lock()
	defer w.Unlock()

	// close file
	w.fileClose()
	w.flush = flush
	return w
}

// SetRotate changes whether or not the old logs are kept. (chainable) Must be
// called before the first log message is written. If rotate < 0, the
// files are overwritten; otherwise, they are rotated to another file before the
// new log is opened.
func (w *CacheFileLogWriter) SetRotate(rotate int) *CacheFileLogWriter {
	//fmt.Fprintf(os.Stderr, "CacheFileLogWriter.SetRotate: %v\n", rotate)
	if rotate > 999 {
		rotate = 999
	}
	w.rotate = rotate
	return w
}

// Set rotate cycle (chainable). Must be called before the first log message is
// written.
func (w *CacheFileLogWriter) SetRotateCycle(cycle int64, delay0 int64) *CacheFileLogWriter {
	w.cycle = cycle
	w.delay0 = delay0
	return w
}

// Set rotate at size (chainable). Must be called before the first log message
// is written.
func (w *CacheFileLogWriter) SetRotateSize(maxsize int64) *CacheFileLogWriter {
	//fmt.Fprintf(os.Stderr, "CacheFileLogWriter.SetRotateSize: %v\n", maxsize)
	w.maxsize = maxsize
	return w
}
