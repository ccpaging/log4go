package log4go

import (
	"io"
	"bufio"
	"os"
	"fmt"
	"strings"
	"runtime"
	"path/filepath"
)

var (
	// Default filename. Set by init
	DefaultFileName = ""

	// Default flush size of cache writing file
	DefaultFileFlush = 4096

	// Default log file and directory perm
	DefaultFilePerm = os.FileMode(0660)
)

type FileWriter struct {
	filename string
	fileflush  int

	file   *os.File
	bufWriter *bufio.Writer
	writer io.Writer
}

func init() {
	base := filepath.Base(os.Args[0])
	ext := filepath.Ext(base)
	DefaultFileName = strings.TrimSuffix(base, ext) + ".log"
	if runtime.GOOS != "windows" {
		DefaultFileName = "~/" + DefaultFileName
	}
}

func (fw *FileWriter) openFile(flag int) (*os.File, error) {
	fd, err := os.OpenFile(fw.filename, flag, DefaultFilePerm)
	if err != nil {
		return nil, err
	}

	fw.file = fd
	fw.writer = fw.file

	if fw.fileflush > 0 {
		fw.bufWriter = bufio.NewWriterSize(fw.file, fw.fileflush)
		fw.writer = fw.bufWriter
	}
	return fd, nil
}

func (fw *FileWriter) CloseFile() {
	defer func() {
		fw.file = nil
		fw.writer = nil
		fw.bufWriter = nil
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

func (fw *FileWriter) FlushFile() {
	if fw.bufWriter != nil {
		fw.bufWriter.Flush()
		return
	}
	if fw.file != nil {
		fw.file.Sync()
	}
}

func (fw *FileWriter) SeekFile(offset int64, whence int) (int64, error) {
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
		_, err := fw.openFile(os.O_WRONLY|os.O_APPEND|os.O_CREATE)
		if err != nil {
			return 0, err
		}
	}
	return fmt.Fprint(fw.writer, s)
}