package main

import (
	"time"
)

import l4g "github.com/ccpaging/log4go"

func main() {
	l4g.DefaultCallerSkip = 2

	log := l4g.NewLogger()
	defer log.Close()
	log.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	//time.Sleep(200 * time.Millisecond)
	// This makes sure the output stream buffer is written
	log.Close()
}
