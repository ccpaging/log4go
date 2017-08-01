// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"net"
	"os"
)

var (
	// Default rotate cycle in seconds
	DefaultSockProto string = "udp"

	// Default rotate delay since midnight in seconds
	DefaultSockEndPoint string = "127.0.0.1:12124"
)

// This log writer sends output to a socket
type SocketLogWriter struct {
	sock 	net.Conn
	proto	string
	hostport string
	format 	string
}

func (w *SocketLogWriter) Close() {
	if w.sock != nil {
		w.sock.Close()
	}
}

func NewSocketLogWriter(proto, hostport string) *SocketLogWriter {
	s := &SocketLogWriter{
		sock:	nil,
		proto:	proto,
		hostport:	hostport,
		format: FORMAT_DEFAULT,
	}
	if s.proto == "" {
		s.proto = DefaultSockProto
	}
	if s.hostport == "" {
		s.hostport = DefaultSockEndPoint
	}
	return s
}

func (s *SocketLogWriter) LogWrite(rec *LogRecord) {
	var err error
	if s.sock == nil {
		s.sock, err = net.Dial(s.proto, s.hostport)
		if err != nil {
			fmt.Fprintf(os.Stderr, "SocketLogWriter(%s): %v\n", s.hostport, err)
			if s.sock != nil {
				s.sock.Close()
				s.sock = nil
			}
			return
		}
	}

	_, err = s.sock.Write([]byte(FormatLogRecord(s.format, rec)))
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "SocketLogWriter(%s): %v\n", s.hostport, err)
	s.sock.Close()
	s.sock = nil
}

// Set option. chainable
func (s *SocketLogWriter) Set(name string, v interface{}) *SocketLogWriter {
	s.SetOption(name, v)
	return s
}

// Set option. checkable
func (s *SocketLogWriter) SetOption(name string, v interface{}) error {
	var ok bool
	switch name {
	case "protocol":
		if s.proto, ok = v.(string); !ok {
			return ErrBadValue
		}
	case "endpoint":
		if s.hostport, ok = v.(string); !ok {
			return ErrBadValue
		}
		if len(s.hostport) <= 0 {
			return ErrBadValue
		} 
	case "format":
		if s.format, ok = v.(string); !ok {
			return ErrBadValue
		}
	default:
		return ErrBadOption
	}
	return nil
}
