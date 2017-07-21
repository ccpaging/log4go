package main

import (
	"time"
)

import l4g "github.com/ccpaging/log4go"

func main() {
	log := l4g.NewLogger()
	log.AddFilter("network", l4g.FINEST, l4g.NewSocketLogWriter("udp", "127.0.0.1:12124"))
	log.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())

	// Run `nc -u -l -p 12124` or similar before you run this to see the following message
	log.Info("The time is now: %s", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Log(l4g.DEBUG, "myApp", "Send a log message with manual level, source, and message.")

	for i := 0; i < 5; i++ {
		time.Sleep(3 * time.Second)
		log.Log(l4g.DEBUG, "myApp", "Send a log message with manual level, source, and message.")
	}
	// This makes sure the output stream buffer is written
	log.Close()
}
