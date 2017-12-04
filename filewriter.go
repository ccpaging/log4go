package log4go

import (
	"io"
	"bufio"
	"os"
	"fmt"
	"sync"
)

type FileWriter struct {
	filename string
	flush  int

	sync.RWMutex

	file   *os.File
	bufWriter *bufio.Writer
	writer io.Writer
}

func NewFileWriter(filename string, flush int) *FileWriter {
	return &FileWriter {
		filename: filename,
		flush:	  flush,
	}
}

func (fw *FileWriter) open(flag int) (*os.File, error) {
	fw.Lock()
	defer fw.Unlock()

	fd, err := os.OpenFile(fw.filename, flag, DefaultFilePerm)
	if err != nil {
		return nil, err
	}

	fw.file = fd
	fw.writer = fw.file

	if fw.flush > 0 {
		fw.bufWriter = bufio.NewWriterSize(fw.file, fw.flush)
		fw.writer = fw.bufWriter
	}
	return fd, nil
}

func (fw *FileWriter) Close() {
	fw.Lock()

	defer func() {
		fw.file = nil
		fw.writer = nil
		fw.bufWriter = nil
		fw.Unlock()
	}()

	if fw.file == nil {
		return
	}

	if fw.bufWriter != nil {
		fw.bufWriter.Flush()
	} else {
		fw.file.Sync()
	}
	fw.file.Close()
}

func (fw *FileWriter) Flush() {
	fw.Lock()
	defer fw.Unlock()

	if fw.bufWriter != nil {
		fw.bufWriter.Flush()
		return
	}
	if fw.file != nil {
		fw.file.Sync()
	}
}

func (fw *FileWriter) Seek(offset int64, whence int) (int64, error) {
	fw.Lock()
	defer fw.Unlock()

	if fw.file != nil {
		return fw.file.Seek(offset, whence)
	}
	
	fi, err := os.Lstat(fw.filename)
	if err != nil {
		return 0, err
	}

	return fi.Size(), nil 
}

func (fw *FileWriter) WriteString(s string) (int, error) {
	if fw.file == nil {
		_, err := fw.open(os.O_WRONLY|os.O_APPEND|os.O_CREATE)
		if err != nil {
			return 0, err
		}
	}

	fw.Lock()
	defer fw.Unlock()
	return fmt.Fprint(fw.writer, s)
}

func (fw *FileWriter) SetFileName(filename string) {
	fw.Close()
	fw.filename = filename
}

func (fw *FileWriter) SetFlush(flush int) {
	fw.Close()
	fw.flush = flush
}