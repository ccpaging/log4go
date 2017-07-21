Writer Extension

log4go do not use any 3rd package to keep it simple.

Then you can program own writer like below.

```
// Define you own writer
type XXXLogWriter struct {
	format 	string
}

// This creates a new XXXLogWriter
func NewXXXLogWriter() *XXXLogWriter {
	lw := &XXXLogWriter{
		format: "[%T %D %Z] [%L] (%S) %M",
	}
	return lw
}

// Set the logging format (chainable).  Must be called before the first log
// message is written.
func (lw *XXXLogWriter) SetFormat(format string) *XXXLogWriter {
	lw.format = format
	return lw
}

func (lw *XXXLogWriter) Close() {
}

func (lw *XXXLogWriter) LogWrite(rec *l4g.LogRecord) {
	switch rec.Level {
		case l4g.CRITICAL:
		case l4g.ERROR:
		case l4g.WARNING:
		case l4g.INFO:
		case l4g.DEBUG:
		case l4g.TRACE:
		default:
	}
	fmt.Println(l4g.FormatLogRecord(lw.format, rec))
}
```

You may create a new writer extension. For example:

* Send error messages as a mail

* Make TCP/UDP server and let client pull the messages

* websocket

* nanomsg pub/sub

* Store log messages in MySQL

Even new configuration file format like INI style.