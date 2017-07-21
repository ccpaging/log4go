package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
	"path/filepath"
)

import l4g "github.com/ccpaging/log4go"

const (
	filename = "_cflw.log"
	oldfiles = "_cflw.*.log"
)

func CheckTimer(cycle int64, delay0 int64) {
	fmt.Println("cycle:", cycle, "delay0:", delay0)
	nrt := time.Now()
	if delay0 < 0 { // Now + cycle
		nrt = nrt.Add(time.Duration(cycle) * time.Second)
	} else { // tomorrow midnight (Clock 0) + delay0
		tomorrow := nrt.Add(24 * time.Hour)
        nrt = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 
						0, 0, 0, 0, tomorrow.Location())
		nrt = nrt.Add(time.Duration(delay0) * time.Second)
	}
	fmt.Println("nrt:", nrt, "now:", time.Now())
	fmt.Println("First timer:", nrt.Sub(time.Now()))
}

// Print what was logged to the file (yes, I know I'm skipping error checking)
func PrintFile(fn string) {
	fd, _ := os.Open(fn)
	in := bufio.NewReader(fd)
	fmt.Print("Messages logged to file were: (line numbers not included)\n")
	for lineno := 1; ; lineno++ {
		line, err := in.ReadString('\n')
		if err == io.EOF {
			break
		}
		fmt.Printf("%3d:\t%s", lineno, line)
	}
	fd.Close()
}

func main() {
	fmt.Println("Every 10 minutes")
	CheckTimer(600, -1)
	fmt.Println("---\nEvery midnight")
	CheckTimer(86400, 0)
	fmt.Println("---\nEvery 3:00am")
	CheckTimer(86400, 10800)
	fmt.Println("---\nEvery weekly midnight")
	CheckTimer(86400 * 7, 0)

	// Get a new logger instance
	log := l4g.NewLogger()

	// Create a default logger that is logging messages of FINE or higher
	log.AddFilter("cachefile", l4g.FINE, l4g.NewCacheFileLogWriter(filename, 0))
	log.Finest("Everything is created now (notice that I will not be printing to the file)")
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Critical("Time to close out!")
	log.Close()

	PrintFile(filename)
	// Remove the file so it's not lying around
	os.Remove(filename)

	/* Can also specify manually via the following: (these are the defaults) */
	flw := l4g.NewCacheFileLogWriter(filename, 10)
	flw.SetFormat("[%D %T] [%L] (%x) %M")
	flw.SetRotateCycle(5, -1) // 5s
	flw.SetRotateSize(1024 * 5) // 5K
	log.AddFilter("cachefile", l4g.FINE, flw)

	// Log some experimental messages
	for j := 0; j < 15; j++ {
		time.Sleep(1 * time.Second)
		for i := 0; i < 200 / (j+1); i++ {
			log.Finest("Everything is created now (notice that I will not be printing to the file)")
			log.Info("%d. The time is now: %s", j, time.Now().Format("15:04:05 MST 2006/01/02"))
			log.Critical("Time to close out!")
		}
		time.Sleep(4 * time.Second)
	}
	// Close the log
	log.Close()

	PrintFile(filename)
	os.Remove(filename)

	files, _ := filepath.Glob(oldfiles)
    fmt.Printf("%d files match %s\n", len(files), oldfiles) // contains a list of all files in the current directory
    for _, f := range files {
		fmt.Printf("Remove %s\n", f)
		os.Remove(f)
    }
}
