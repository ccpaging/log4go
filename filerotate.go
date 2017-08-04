package log4go

import (
	"os"
	"fmt"
	"strings"
	"path/filepath"
	"time"
)

type FileRotate struct {
	rotCount int
	rotFiles chan string
}

var (
	DefaultRotateLen = 5
)

func (r *FileRotate) initRot() {
	r.rotCount = 0
	r.rotFiles = make(chan string, DefaultRotateLen)
}

// Rename history log files to "<name>.00?.<ext>"
func (r *FileRotate) rotFile(filename string, rotate int, newLog string) {
	r.rotFiles <- newLog 
	if r.rotCount > 0 {
		if DEBUG_ROTATE { fmt.Println("queued", newLog) }
		return
	}

	r.rotCount++

	for len(r.rotFiles) > 0 {
		newFile, _ := <- r.rotFiles
	
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
	r.rotCount--
}

func (r *FileRotate) closeRot() {
	for i := 10; i > 0; i-- {
		// Must call Sleep here, otherwise, may panic send on closed channel
		time.Sleep(100 * time.Millisecond)
		if r.rotCount <= 0 {
			break
		}
	}

	close(r.rotFiles)

	// drain the files not rotated
	for file := range r.rotFiles {
		fmt.Fprintf(os.Stderr, "FileLogWriter: Not rotate %s\n", file)
	}
}
