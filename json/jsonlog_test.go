package jsonlog

import (
	"os"
	"time"
	"runtime"
	// "io/ioutil"
	"fmt"

	l4g "github.com/ccpaging/log4go"
	"testing"
)

const testLogFile = "_logtest.log"

var now time.Time = time.Unix(0, 1234567890123456789).In(time.UTC)

func newLogRecord(lvl l4g.Level, src string, msg string) *l4g.LogRecord {
	return &l4g.LogRecord{
		Level:   lvl,
		Source:  src,
		Created: now,
		Message: msg,
	}
}

func TestJsonLogWriter(t *testing.T) {
	defer func(buflen int) {
		l4g.DefaultBufferLength = buflen
	}(l4g.DefaultBufferLength)
	l4g.DefaultBufferLength = 0

	w := NewLogWriter("udp", "127.0.0.1:12124")
	if w == nil {
		t.Fatalf("Invalid return: w should not be nil")
	}
	defer os.Remove(testLogFile)

	w.LogWrite(newLogRecord(l4g.CRITICAL, "source", "message"))
	w.Close()
	runtime.Gosched()

	/*
	if contents, err := ioutil.ReadFile(testLogFile); err != nil {
		t.Errorf("read(%q): %s", testLogFile, err)
	} else if len(contents) != 170 {
		t.Errorf("malformed xmlog: %q (%d bytes)", string(contents), len(contents))
	}
	*/
}

func TestJsonConfig(t *testing.T) {
	const (
		configfile = "_example.json"
	)

	fd, err := os.Create(configfile)
	if err != nil {
		t.Fatalf("Could not open %s for writing: %s", configfile, err)
	}

	fmt.Fprintln(fd, 
`{
	"filters": [
	{
	  "enabled": "true",
	  "tag": "stdout",
	  "type": "console",
	  "level": "DEBUG",
	  "properties": [
		{
		  "name": "format",
		  "value": "[%D %T] [%L] (%S) %M"
		}
	  ]
	},
	{
	  "enabled": "true",
	  "tag": "file",
	  "type": "file",
	  "level": "FINEST",
	  "properties": [
		{
		  "name": "filename",
		  "value": "_test.log"
		}
	  ]
	},
	{
	  "enabled": "true",
	  "tag": "socket",
	  "type": "socket",
	  "level": "FINEST",
	  "properties": [
		{
		  "name": "protocol",
		  "value": "udp"
		},
		{
		  "name": "endpoint",
		  "value": "127.0.0.1:12124"
		}
	  ]
	}
	]
}`)
	fd.Close()

	log := l4g.NewLogger()
	LoadConfiguration(log, configfile)
	defer os.Remove("_test.log")
	defer log.Close()

	// Make sure we got all loggers
	if len(log) != 3 {
		t.Fatalf("JsonConfig: Expected 3 filters, found %d", len(log))
	}

	// Make sure they're the right keys
	if _, ok := log["stdout"]; !ok {
		t.Errorf("JsonConfig: Expected stdout logger")
	}
	if _, ok := log["file"]; !ok {
		t.Fatalf("JsonConfig: Expected file logger")
	}

	// Make sure they're the right type
	if _, ok := log["stdout"].LogWriter.(*l4g.ConsoleLogWriter); !ok {
		t.Fatalf("JsonConfig: Expected stdout to be ConsoleLogWriter, found %T", log["stdout"].LogWriter)
	}
	if _, ok := log["file"].LogWriter.(*l4g.FileLogWriter); !ok {
		t.Fatalf("JsonConfig: Expected file to be *FileLogWriter, found %T", log["file"].LogWriter)
	}

	// Make sure levels are set
	if lvl := log["stdout"].Level; lvl != l4g.DEBUG {
		t.Errorf("JsonConfig: Expected stdout to be set to level %d, found %d", l4g.DEBUG, lvl)
	}
	if lvl := log["file"].Level; lvl != l4g.FINEST {
		t.Errorf("JsonConfig: Expected file to be set to level %d, found %d", l4g.FINEST, lvl)
	}

	// Make sure the w is open and points to the right file
	/*
	flw := log["file"].LogWriter.(*l4g.FileLogWriter)
	if fname, _ := flw.GetOption("filename"); fname != "_test.log" {
		t.Errorf("JsonConfig: Expected file to have opened %s, found %s", "test.log", fname)
	}
	*/
	/*
	if flw.file != nil {
		if fname := flw.file.Name(); fname != "_test.log" {
			t.Errorf("JsonConfig: Expected file to have opened %s, found %s", "test.log", fname)
		}
	}
	*/

	// Save Json config file
	err = os.Rename(configfile, "example/singleconfig.json") // Keep this so that an example with the documentation is available
	if err != nil {
		t.Fatalf("Could not rename %s: %s", configfile, err)
	}
	os.Remove(configfile)
}
