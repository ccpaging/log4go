package main

import (
	"time"
	log "github.com/ccpaging/log4go"
)

func main() {
	log.Print("This is Print()\n")
	log.Println("This is Println()")
	log.Printf("The time is now: %s\n", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Stderr("This is Stderr\n")
	log.Stderrf("The time is now: %s\n", time.Now().Format("15:04:05 MST 2006/01/02"))
	log.Panicf("The time is now: %s\n", time.Now().Format("15:04:05 MST 2006/01/02"))
}
