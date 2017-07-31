// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package jsonlog

import (
	"fmt"
	"os"
	"net"
	"encoding/json"
	l4g "github.com/ccpaging/log4go"
)

// This log writer sends output to a socket
type JsonLogWriter struct {
	sock 	net.Conn
	proto	string
	hostport string
}

func (w *JsonLogWriter) Close() {
	if w.sock != nil {
		w.sock.Close()
	}
}

func NewLogWriter(proto, hostport string) *JsonLogWriter {
	s := &JsonLogWriter{
		sock:	nil,
		proto:	proto,
		hostport:	hostport,
	}
	return s
}

func (s *JsonLogWriter) LogWrite(rec *l4g.LogRecord) {
	// Marshall into JSON
	js, err := json.Marshal(rec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JsonLogWriter(%s): %v\n", s.hostport, err)
		return
	}

	if s.sock == nil {
		s.sock, err = net.Dial(s.proto, s.hostport)
		if err != nil {
			fmt.Fprintf(os.Stderr, "JsonLogWriter(%s): %v\n", s.hostport, err)
			if s.sock != nil {
				s.sock.Close()
				s.sock = nil
			}
			return
		}
	}

	_, err = s.sock.Write(js)
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "JsonLogWriter(%s): %v\n", s.hostport, err)
	s.sock.Close()
	s.sock = nil
}
