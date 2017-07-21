package main

import (
	"time"

	log "github.com/ccpaging/log4go"
	"github.com/ccpaging/log4go/colorlog"
)


func main() {
 	// AddFilter will close LogWriter with name "stdout" automatic
	// "stdout" is default ConsoleLogWriter added to log.Global
	// As old version. Have to close all exist LogWriter.
	// log.Close()
	log.AddFilter("stdout", log.DEBUG, colorlog.NewColorLogWriter())
	log.Debug("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Warn("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))

	// This makes sure the filters is running
	time.Sleep(200 * time.Millisecond)
}
