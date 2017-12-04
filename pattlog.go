// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"bytes"
	"fmt"
	"strings"
)

var (
	FORMAT_DEFAULT = "[%D %T %z] [%L] (%S) %M"
	FORMAT_SHORT   = "[%t %d] [%L] %M"
	FORMAT_ABBREV  = "[%L] %M"
	FORMAT_UTC	   = false
)

type formatCacheType struct {
	LastUpdateSeconds    int64
	longTime, shortTime string
	longZone, shortZone string
	longDate, shortDate   string
}

var formatCache = &formatCacheType{}

// Known format codes:
// %T - Time (15:04:05)
// %t - Time (15:04)
// %Z - Zone (-0700)
// %z - Zone (MST)
// %D - Date (2006/01/02)
// %d - Date (01/02/06)
// %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
// %S - Source
// %s - Short Source
// %x - Extra Short Source: just file without .go suffix
// %M - Message
// Ignores unknown formats
// Recommended: "[%D %T] [%L] (%S) %M"
func FormatLogRecord(format string, rec *LogRecord) string {
	if rec == nil {
		return "<nil>"
	}
	if len(format) == 0 {
		return ""
	}

	t := rec.Created
	if FORMAT_UTC {
		t = t.UTC()
	}

	out := bytes.NewBuffer(make([]byte, 0, 64))
	secs := t.UnixNano() / 1e9

	cache := *formatCache
	if cache.LastUpdateSeconds != secs {
		month, day, year := t.Month(), t.Day(), t.Year()
		hour, minute, second := t.Hour(), t.Minute(), t.Second()
		updated := &formatCacheType{
			LastUpdateSeconds: secs,
			shortTime:         fmt.Sprintf("%02d:%02d", hour, minute),
			longTime:          fmt.Sprintf("%02d:%02d:%02d", hour, minute, second),
			shortZone:         t.Format("MST"),
			longZone:          t.Format("-0700"),
			shortDate:         fmt.Sprintf("%02d/%02d/%02d", day, month, year%100),
			longDate:          fmt.Sprintf("%04d/%02d/%02d", year, month, day),
		}
		cache = *updated
		formatCache = updated
	}

	// Split the string into pieces by % signs
	pieces := bytes.Split([]byte(format), []byte{'%'})

	// Iterate over the pieces, replacing known formats
	for i, piece := range pieces {
		if i > 0 && len(piece) > 0 {
			switch piece[0] {
			case 'T':
				out.WriteString(cache.longTime)
			case 't':
				out.WriteString(cache.shortTime)
			case 'Z':
				out.WriteString(cache.longZone)
			case 'z':
				out.WriteString(cache.shortZone)
			case 'D':
				out.WriteString(cache.longDate)
			case 'd':
				out.WriteString(cache.shortDate)
			case 'L':
				out.WriteString(levelStrings[rec.Level])
			case 'S':
				out.WriteString(rec.Source)
			case 's':
				out.WriteString(rec.Source[strings.LastIndex(rec.Source, "/")+1:])
			case 'x':
				fn := rec.Source[strings.LastIndex(rec.Source, "/")+1:]
				out.WriteString(fn[strings.LastIndex(fn, ".")+1:])
			case 'M':
				out.WriteString(rec.Message)
			}
			if len(piece) > 1 {
				out.Write(piece[1:])
			}
		} else if len(piece) > 0 {
			out.Write(piece)
		}
	}
	out.WriteByte('\n')

	return out.String()
}

