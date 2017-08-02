package log4go

import (
	"os"
	"fmt"
	"strings"
	"path/filepath"
)

func FileRotate(filename string, rotate int, newLog string) {
	// May compress new log file here

	if DEBUG_ROTATE { fmt.Println(filename, "Rename", newLog, "already") }
	
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
	
	for ; n > 1; n-- {
		prev := path + fmt.Sprintf(".%03d", n - 1) + ext

		if DEBUG_ROTATE { fmt.Println(prev, "Rename", slot) }

		os.Rename(prev, slot)
		slot = prev
	}
	
	if DEBUG_ROTATE { fmt.Println(newLog, "Rename", path + ".001" + ext) }

	os.Rename(newLog, path + ".001" + ext)
}