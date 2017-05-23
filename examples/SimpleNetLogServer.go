package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

import l4g "github.com/ccpaging/log4go"

var (
	port = flag.String("p", "12124", "Port number to listen on")
)

func handleListener(log *l4g.Logger, listener *net.UDPConn){
    var buffer [4096]byte
    
	// read into a new buffer
	buflen, addr, err := listener.ReadFrom(buffer[0:])
    if err != nil{
		fmt.Println("[Error] [", addr, "] ", err)
        return
    }

	if buflen <= 0{
		fmt.Println("[Error] [", addr, "] ", "Empty packet")
        return
	}

	log.Json(buffer[:buflen])
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("Erroring out: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	log := l4g.NewLogger()
	log.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())

	// Bind to the port
	bind, err := net.ResolveUDPAddr("udp4", "0.0.0.0:" + *port)
	checkError(err)

	fmt.Printf("Listening to port %s...\n", *port)
	
	// Create listener
	listener, err := net.ListenUDP("udp", bind)
	checkError(err)

	for {
		handleListener(&log, listener)
	}

	// This makes sure the output stream buffer is written
	log.Close()
}
