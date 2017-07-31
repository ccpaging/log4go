package colorlog

import (
	"io"
	"time"

	l4g "github.com/ccpaging/log4go"
	"testing"
)

var now time.Time = time.Unix(0, 1234567890123456789).In(time.UTC)

var logRecordWriteTests = []struct {
	Test    string
	Record  *l4g.LogRecord
	Console string
}{
	{
		Test: "Normal message",
		Record: &l4g.LogRecord{
			Level:   l4g.CRITICAL,
			Source:  "source",
			Message: "message",
			Created: now,
		},
		Console: "[23:31:30 UTC 2009/02/13] [CRIT] [source] message",
	},
}

func Test(t *testing.T) {
	console := new(ColorLogWriter)
	
	console.format = "[%T %z %D] [%L] [%S] %M"

	r, w := io.Pipe()
	console.iow = w

	defer console.Close()

	buf := make([]byte, 1024)

	for _, test := range logRecordWriteTests {
		name := test.Test

		// Pipe write and read must be in diff routines otherwise cause dead lock
		go console.LogWrite(test.Record)
		
		n, _ := r.Read(buf)
		if got, want := string(buf[:n]), test.Console; got != (want+"\n") {
			t.Errorf("%s:  got %q", name, got)
			t.Errorf("%s: want %q", name, want)
		}
	}
}