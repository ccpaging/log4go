package log4go

import (
	"os"
	"fmt"
	"strings"
	"path/filepath"
	"time"
)

type FileRotate struct {
	count int
	files chan string
}

var (
	DefaultRotateLen = 5
)

func NewFileRotate() *FileRotate {
	return &FileRotate{
		count: 0,
		files: make(chan string, DefaultRotateLen),
	}
}

// Rename history log files to "<name>.00?.<ext>"
func (r *FileRotate) Rotate(filename string, rotate int, newLog string) {
	r.files <- newLog 
	if r.count > 0 {
		if DEBUG_ROTATE { fmt.Println("queued", newLog) }
		return
	}

	go func() {
		r.count++
		for len(r.files) > 0 {
			newFile, _ := <- r.files
	
			// May compress new log file here

			if DEBUG_ROTATE { fmt.Println(filename, "Rename", newFile, "already") }
	
			ext := filepath.Ext(filename) // like ".log"
			path := strings.TrimSuffix(filename, ext) // include dir
		
			if DEBUG_ROTATE { fmt.Println(rotate, path, ext) }
	
			// May create old directory here
	
			var n int
			var err error = nil 
			slot := ""
			for n = 1; n <= rotate; n++ {
				slot = path + fmt.Sprintf(".%03d", n) + ext
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
	
			// May compress previous log file here
	
			for ; n > 1; n-- {
				prev := path + fmt.Sprintf(".%03d", n - 1) + ext

				if DEBUG_ROTATE { fmt.Println(prev, "Rename", slot) }

				os.Rename(prev, slot)
				slot = prev
			}
	
			if DEBUG_ROTATE { fmt.Println(newFile, "Rename", path + ".001" + ext) }

			os.Rename(newFile, path + ".001" + ext)
		}
		r.count--
	}()
}

func (r *FileRotate) Close() {
	for i := 10; i > 0; i-- {
		// Must call Sleep here, otherwise, may panic send on closed channel
		time.Sleep(100 * time.Millisecond)
		if r.count <= 0 {
			break
		}
	}

	close(r.files)

	// drain the files not rotated and print
	for file := range r.files {
		fmt.Fprintf(os.Stderr, "FileLogWriter: Not rotate %s\n", file)
	}
}
