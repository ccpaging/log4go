package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"encoding/json"
	log "github.com/ccpaging/log4go"
)



var (
	port = flag.String("p", "12124", "Port number to listen on")
)

func handleListener(listener *net.UDPConn){
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

	// fmt.Println(string(buffer[:buflen]))

	rec, err := Decode(buffer[:buflen])
	if err != nil {
		fmt.Printf("Err: %v, [%s]\n", err, string(buffer[:buflen]))
	}
	// fmt.Println(rec)
	log.Log(rec.Level, rec.Source, rec.Message)
}

func Decode(data []byte) (*log.LogRecord, error) {
	var rec log.LogRecord
	
	// Make the log record
	err := json.Unmarshal(data, &rec)
	if err != nil {
		return nil, err
	}
	
	return &rec, nil
}

func checkError(err error) {
	if err != nil {
		fmt.Printf("Erroring out: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	// Bind to the port
	bind, err := net.ResolveUDPAddr("udp4", "0.0.0.0:" + *port)
	checkError(err)

	fmt.Printf("Listening to port %s...\n", *port)
	
	// Create listener
	listener, err := net.ListenUDP("udp", bind)
	checkError(err)

	for {
		handleListener(listener)
	}

	// This makes sure the output stream buffer is written
	log.Close()
}
