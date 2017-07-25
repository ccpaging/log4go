// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
	"time"
)

const testLogFile = "_logtest.log"
const benchLogFile = "_benchlog.log"

var now time.Time = time.Unix(0, 1234567890123456789).In(time.UTC)

func newLogRecord(lvl Level, src string, msg string) *LogRecord {
	return &LogRecord{
		Level:   lvl,
		Source:  src,
		Created: now,
		Message: msg,
	}
}

func TestELog(t *testing.T) {
	fmt.Printf("Testing %s\n", L4G_VERSION)
	lr := newLogRecord(CRITICAL, "source", "message")
	if lr.Level != CRITICAL {
		t.Errorf("Incorrect level: %d should be %d", lr.Level, CRITICAL)
	}
	if lr.Source != "source" {
		t.Errorf("Incorrect source: %s should be %s", lr.Source, "source")
	}
	if lr.Message != "message" {
		t.Errorf("Incorrect message: %s should be %s", lr.Source, "message")
	}
}

var formatTests = []struct {
	Test    string
	Record  *LogRecord
	Formats map[string]string
}{
	{
		Test: "Standard formats",
		Record: &LogRecord{
			Level:   ERROR,
			Source:  "source",
			Message: "message",
			Created: now,
		},
		Formats: map[string]string{
			// TODO(kevlar): How can I do this so it'll work outside of PST?
			FORMAT_DEFAULT: "[2009/02/13 23:31:30 UTC] [EROR] (source) message\n",
			FORMAT_SHORT:   "[23:31 13/02/09] [EROR] message\n",
			FORMAT_ABBREV:  "[EROR] message\n",
		},
	},
}

func TestFormatLogRecord(t *testing.T) {
	for _, test := range formatTests {
		name := test.Test
		for fmt, want := range test.Formats {
			if got := FormatLogRecord(fmt, test.Record); got != want {
				t.Errorf("%s - %s:", name, fmt)
				t.Errorf("   got %q", got)
				t.Errorf("  want %q", want)
			}
		}
	}
}

var logRecordWriteTests = []struct {
	Test    string
	Record  *LogRecord
	Console string
}{
	{
		Test: "Normal message",
		Record: &LogRecord{
			Level:   CRITICAL,
			Source:  "source",
			Message: "message",
			Created: now,
		},
		Console: "[23:31:30 UTC 2009/02/13] [CRIT] [source] message",
	},
}

func TestConsoleLogWriter(t *testing.T) {
	console := new(ConsoleLogWriter)
	
	console.format = "[%T %z %D] [%L] [%S] %M"

	r, w := io.Pipe()
	console.out = w

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

func TestFileLogWriter(t *testing.T) {
	defer func(buflen int) {
		DefaultBufferLength = buflen
	}(DefaultBufferLength)
	DefaultBufferLength = 0

	w := NewFileLogWriter(testLogFile, false)
	if w == nil {
		t.Fatalf("Invalid return: w should not be nil")
	}
	defer os.Remove(testLogFile)

	w.LogWrite(newLogRecord(CRITICAL, "source", "message"))
	w.Close()
	runtime.Gosched()

	if contents, err := ioutil.ReadFile(testLogFile); err != nil {
		t.Errorf("read(%q): %s", testLogFile, err)
	} else if len(contents) != 50 {
		t.Errorf("malformed filelog: %q (%d bytes)", string(contents), len(contents))
	}
}

func TestCacheFileLogWriter(t *testing.T) {
	defer func(buflen int) {
		DefaultBufferLength = buflen
	}(DefaultBufferLength)
	DefaultBufferLength = 0

	w := NewCacheFileLogWriter(testLogFile, 0)
	if w == nil {
		t.Fatalf("Invalid return: w should not be nil")
	}
	defer os.Remove(testLogFile)

	w.LogWrite(newLogRecord(CRITICAL, "source", "message"))
	w.Close()
	runtime.Gosched()

	if contents, err := ioutil.ReadFile(testLogFile); err != nil {
		t.Errorf("read(%q): %s", testLogFile, err)
	} else if len(contents) != 50 {
		t.Errorf("malformed filelog: %q (%d bytes)", string(contents), len(contents))
	}
}

func TestLogger(t *testing.T) {
	sl := NewDefaultLogger(WARNING)
	if sl == nil {
		t.Fatalf("NewDefaultLogger should never return nil")
	}
	if lw, exist := sl["stdout"]; lw == nil || exist != true {
		t.Fatalf("NewDefaultLogger produced invalid logger (DNE or nil)")
	}
	if sl["stdout"].Level != WARNING {
		t.Fatalf("NewDefaultLogger produced invalid logger (incorrect level)")
	}
	if len(sl) != 1 {
		t.Fatalf("NewDefaultLogger produced invalid logger (incorrect map count)")
	}

	//func (l *Logger) AddFilter(name string, level int, writer LogWriter) {}
	l := make(Logger)
	l.AddFilter("stdout", DEBUG, NewConsoleLogWriter())
	if lw, exist := l["stdout"]; lw == nil || exist != true {
		t.Fatalf("AddFilter produced invalid logger (DNE or nil)")
	}
	if l["stdout"].Level != DEBUG {
		t.Fatalf("AddFilter produced invalid logger (incorrect level)")
	}
	if len(l) != 1 {
		t.Fatalf("AddFilter produced invalid logger (incorrect map count)")
	}

	//func (l *Logger) Warn(format string, args ...interface{}) error {}
	if err := l.Warn("%s %d %#v", "Warning:", 1, []int{}); err.Error() != "Warning: 1 []int{}" {
		t.Errorf("Warn returned invalid error: %s", err)
	}

	//func (l *Logger) Error(format string, args ...interface{}) error {}
	if err := l.Error("%s %d %#v", "Error:", 10, []string{}); err.Error() != "Error: 10 []string{}" {
		t.Errorf("Error returned invalid error: %s", err)
	}

	//func (l *Logger) Critical(format string, args ...interface{}) error {}
	if err := l.Critical("%s %d %#v", "Critical:", 100, []int64{}); err.Error() != "Critical: 100 []int64{}" {
		t.Errorf("Critical returned invalid error: %s", err)
	}
	// Already tested or basically untestable
	//func (l *Logger) Log(level int, source, message string) {}
	//func (l *Logger) Logf(level int, format string, args ...interface{}) {}
	//func (l *Logger) intLogf(level int, format string, args ...interface{}) string {}
	//func (l *Logger) Finest(format string, args ...interface{}) {}
	//func (l *Logger) Fine(format string, args ...interface{}) {}
	//func (l *Logger) Debug(format string, args ...interface{}) {}
	//func (l *Logger) Trace(format string, args ...interface{}) {}
	//func (l *Logger) Info(format string, args ...interface{}) {}
}

func TestLogOutput(t *testing.T) {
	const (
		expected = "fdf3e51e444da56b4cb400f30bc47424"
	)

	// Unbuffered output
	defer func(buflen int) {
		DefaultBufferLength = buflen
	}(DefaultBufferLength)
	DefaultBufferLength = 0

	l := make(Logger)

	// Delete and open the output log without a timestamp (for a constant md5sum)
	l.AddFilter("file", FINEST, NewFileLogWriter(testLogFile, false).SetFormat("[%L] %M"))
	defer os.Remove(testLogFile)

	// Send some log messages
	l.Log(CRITICAL, "testsrc1", fmt.Sprintf("This message is level %d", int(CRITICAL)))
	l.Logf(ERROR, "This message is level %v", ERROR)
	l.Logf(WARNING, "This message is level %s", WARNING)
	l.Logc(INFO, func() string { return "This message is level INFO" })
	l.Trace("This message is level %d", int(TRACE))
	l.Debug("This message is level %s", DEBUG)
	l.Fine(func() string { return fmt.Sprintf("This message is level %v", FINE) })
	l.Finest("This message is level %v", FINEST)
	l.Finest(FINEST, "is also this message's level")

	l.Close()

	contents, err := ioutil.ReadFile(testLogFile)
	if err != nil {
		t.Fatalf("Could not read output log: %s", err)
	}

	sum := md5.New()
	sum.Write(contents)
	if sumstr := hex.EncodeToString(sum.Sum(nil)); sumstr != expected {
		t.Errorf("--- Log Contents:\n%s---", string(contents))
		t.Fatalf("Checksum does not match: %s (expecting %s)", sumstr, expected)
	}
}

func TestCountMallocs(t *testing.T) {
	const N = 1
	var m runtime.MemStats
	getMallocs := func() uint64 {
		runtime.ReadMemStats(&m)
		return m.Mallocs
	}

	// Console logger
	sl := NewDefaultLogger(INFO)
	mallocs := 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.Log(WARNING, "here", "This is a WARNING message")
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per sl.Log((WARNING, \"here\", \"This is a log message\"): %d\n", mallocs/N)

	// Console logger formatted
	mallocs = 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.Logf(WARNING, "%s is a log message with level %d", "This", WARNING)
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per sl.Logf(WARNING, \"%%s is a log message with level %%d\", \"This\", WARNING): %d\n", mallocs/N)

	// Console logger (not logged)
	sl = NewDefaultLogger(INFO)
	mallocs = 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.Log(DEBUG, "here", "This is a DEBUG log message")
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per unlogged sl.Log((WARNING, \"here\", \"This is a log message\"): %d\n", mallocs/N)

	// Console logger formatted (not logged)
	mallocs = 0 - getMallocs()
	for i := 0; i < N; i++ {
		sl.Logf(DEBUG, "%s is a log message with level %d", "This", DEBUG)
	}
	mallocs += getMallocs()
	fmt.Printf("mallocs per unlogged sl.Logf(WARNING, \"%%s is a log message with level %%d\", \"This\", WARNING): %d\n", mallocs/N)
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

	log := NewLogger()
	log.LoadConfiguration(configfile)
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
	if _, ok := log["stdout"].LogWriter.(*ConsoleLogWriter); !ok {
		t.Fatalf("JsonConfig: Expected stdout to be ConsoleLogWriter, found %T", log["stdout"].LogWriter)
	}
	if _, ok := log["file"].LogWriter.(*FileLogWriter); !ok {
		t.Fatalf("JsonConfig: Expected file to be *FileLogWriter, found %T", log["file"].LogWriter)
	}

	// Make sure levels are set
	if lvl := log["stdout"].Level; lvl != DEBUG {
		t.Errorf("JsonConfig: Expected stdout to be set to level %d, found %d", DEBUG, lvl)
	}
	if lvl := log["file"].Level; lvl != FINEST {
		t.Errorf("JsonConfig: Expected file to be set to level %d, found %d", FINEST, lvl)
	}

	// Make sure the w is open and points to the right file
	if fname := log["file"].LogWriter.(*FileLogWriter).file.Name(); fname != "_test.log" {
		t.Errorf("JsonConfig: Expected file to have opened %s, found %s", "test.log", fname)
	}

	// Save Json config file
	err = os.Rename(configfile, "examples/singleconfig.json") // Keep this so that an example with the documentation is available
	if err != nil {
		t.Fatalf("Could not rename %s: %s", configfile, err)
	}
	os.Remove(configfile)
}

func BenchmarkFormatLogRecord(b *testing.B) {
	const updateEvery = 1
	rec := &LogRecord{
		Level:   CRITICAL,
		Created: now,
		Source:  "source",
		Message: "message",
	}
	for i := 0; i < b.N; i++ {
		rec.Created = rec.Created.Add(1 * time.Second / updateEvery)
		if i%2 == 0 {
			FormatLogRecord(FORMAT_DEFAULT, rec)
		} else {
			FormatLogRecord(FORMAT_SHORT, rec)
		}
	}
}

func BenchmarkConsoleLog(b *testing.B) {
	/* This doesn't seem to work on OS X
	sink, err := os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}
	if err := syscall.Dup2(int(sink.Fd()), syscall.Stdout); err != nil {
		panic(err)
	}
	*/

	stdout = ioutil.Discard
	sl := NewDefaultLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Log(WARNING, "here", "This is a log message")
	}
}

func BenchmarkConsoleNotLogged(b *testing.B) {
	sl := NewDefaultLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Log(DEBUG, "here", "This is a log message")
	}
}

func BenchmarkConsoleUtilLog(b *testing.B) {
	sl := NewDefaultLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
}

func BenchmarkConsoleUtilNotLog(b *testing.B) {
	sl := NewDefaultLogger(INFO)
	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
}

func BenchmarkFileLog(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter(benchLogFile, false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Log(WARNING, "here", "This is a log message")
	}
	b.StopTimer()
	sl.Close()
	os.Remove(benchLogFile)
}

func BenchmarkFileNotLogged(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter(benchLogFile, false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Log(DEBUG, "here", "This is a log message")
	}
	b.StopTimer()
	sl.Close()
	os.Remove(benchLogFile)
}

func BenchmarkFileUtilLog(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter(benchLogFile, false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
	b.StopTimer()
	sl.Close()
	os.Remove(benchLogFile)
}

func BenchmarkFileUtilNotLog(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("file", INFO, NewFileLogWriter(benchLogFile, false))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
	b.StopTimer()
	sl.Close()
	os.Remove(benchLogFile)
}

func BenchmarkCacheFileLog(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("cachefile", INFO, NewCacheFileLogWriter(benchLogFile, 0))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Log(WARNING, "here", "This is a log message")
	}
	b.StopTimer()
	sl.Close()
	os.Remove(benchLogFile)
}

func BenchmarkCacheFileNotLogged(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("cachefile", INFO, NewCacheFileLogWriter(benchLogFile, 0))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Log(DEBUG, "here", "This is a log message")
	}
	b.StopTimer()
	sl.Close()
	os.Remove(benchLogFile)
}

func BenchmarkCacheFileUtilLog(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("cachefile", INFO, NewCacheFileLogWriter(benchLogFile, 0))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Info("%s is a log message", "This")
	}
	b.StopTimer()
	sl.Close()
	os.Remove(benchLogFile)
}

func BenchmarkCacheFileUtilNotLog(b *testing.B) {
	sl := make(Logger)
	b.StopTimer()
	sl.AddFilter("cachefile", INFO, NewCacheFileLogWriter(benchLogFile, 0))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		sl.Debug("%s is a log message", "This")
	}
	b.StopTimer()
	sl.Close()
	os.Remove(benchLogFile)
}

// Benchmark results (darwin amd64 6g)
//elog.BenchmarkConsoleLog           100000       22819 ns/op
//elog.BenchmarkConsoleNotLogged    2000000         879 ns/op
//elog.BenchmarkConsoleUtilLog        50000       34380 ns/op
//elog.BenchmarkConsoleUtilNotLog   1000000        1339 ns/op
//elog.BenchmarkFileLog              100000       26497 ns/op
//elog.BenchmarkFileNotLogged       2000000         821 ns/op
//elog.BenchmarkFileUtilLog           50000       33945 ns/op
//elog.BenchmarkFileUtilNotLog      1000000        1258 ns/op
