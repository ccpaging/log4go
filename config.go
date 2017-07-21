// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
	"strings"
	"encoding/json"
)

type FilterProp struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type FilterConfig struct {
	Enabled  string        `xml:"enabled,attr"`
	Tag      string        `xml:"tag"`
	Level    string        `xml:"level"`
	Type     string        `xml:"type"`
	Properties []FilterProp `xml:"property"`
}

type LogConfig struct {
	Filters []FilterConfig `xml:"filter"`
}

func (log Logger) LoadConfiguration(filename string) {
	if len(filename) <= 0 {
		return
	}

	// Open the LogConfiguration file
	fd, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Could not open %q for reading. %s\n", filename, err)
		os.Exit(1)
	}
	defer fd.Close()

	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Could not read %q. %s\n", filename, err)
		os.Exit(1)
	}

	log.LoadConfigBuf(buf)
	return
}

func (log Logger) LoadConfigBuf(contents []byte) {
	jc := new(LogConfig)
	if err := json.Unmarshal(contents, jc); err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Could not parse Json LogConfiguration. %s\n", err)
		os.Exit(1)
	}

	log.Close()
	for _, fc := range jc.Filters {
		bad, enabled, lvl := log.CheckFilterConfig(fc)

		// Just so all of the required attributes are errored at the same time if missing
		if bad {
			os.Exit(1)
		}

		lw, good := log.MakeLogWriter(fc, enabled)

		// Just so all of the required params are errored at the same time if wrong
		if !good {
			os.Exit(1)
		}

		// If we're disabled (syntax and correctness checks only), don't add to logger
		if !enabled {
			continue
		}

		if lw == nil {
			fmt.Fprintf(os.Stderr, "LoadConfiguration: LogWriter is nil. %v\n", fc)
			os.Exit(1)
		}

		log.AddFilter(fc.Tag, lvl, lw)
	}
}

func (log Logger) CheckFilterConfig(fc FilterConfig) (bad bool, enabled bool, lvl Level) {
	bad, enabled, lvl = false, false, INFO

	// Check required children
	if len(fc.Enabled) == 0 {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Required attribute %s\n", "enabled")
		bad = true
	} else {
		enabled = fc.Enabled != "false"
	}
	if len(fc.Tag) == 0 {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Required child <%s>\n", "tag")
		bad = true
	}
	if len(fc.Type) == 0 {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Required child <%s>\n", "type")
		bad = true
	}
	if len(fc.Level) == 0 {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Required child <%s>\n", "level")
		bad = true
	}

	switch fc.Level {
	case "FINEST":
		lvl = FINEST
	case "FINE":
		lvl = FINE
	case "DEBUG":
		lvl = DEBUG
	case "TRACE":
		lvl = TRACE
	case "INFO":
		lvl = INFO
	case "WARNING":
		lvl = WARNING
	case "ERROR":
		lvl = ERROR
	case "CRITICAL":
		lvl = CRITICAL
	default:
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Required child <%s> for filter has unknown value. %s\n", "level", fc.Level)
		bad = true
	}
	return bad, enabled, lvl
}

func (log Logger) MakeLogWriter(fc FilterConfig, enabled bool) (LogWriter, bool) {
	var (
		lw LogWriter
		good bool
	)
	switch fc.Type {
	case "console":
		lw, good = propToConsoleLogWriter(fc.Properties, enabled)
	case "file":
		lw, good = log.PropToFileLogWriter(fc.Properties, enabled)
	case "socket":
		lw, good = propToSocketLogWriter(fc.Properties, enabled)
	default:
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Could not load LogConfiguration. Unknown filter type \"%s\"\n", fc.Type)
		return nil, false
	}
	return lw, good
}

func propToConsoleLogWriter(props []FilterProp, enabled bool) (*ConsoleLogWriter, bool) {
	format := "[%D %T] [%L] (%S) %M"
	// Parse properties
	for _, prop := range props {
		switch prop.Name {
		case "format":
			format = strings.Trim(prop.Value, " \r\n")
		default:
			fmt.Fprintf(os.Stderr, "LoadConfiguration Warning: Unknown property \"%s\" for console filter\n", prop.Name)
		}
	}

	// If it's disabled, we're just checking syntax
	if !enabled {
		return nil, true
	}

	clw := NewConsoleLogWriter()
	clw.SetFormat(format)
	return clw, true
}

// Parse a duration string.
// A duration string is a possibly signed sequence of decimal numbers,
// each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us", "ms", "s", "m", "h".
func StrToTimeDuration(str string) int64 {
	dur, _ := time.ParseDuration(str)
	return int64(dur/time.Millisecond)
}

// Parse a number with K/M/G suffixes based on thousands (1000) or 2^10 (1024)
func StrToNumSuffix(str string, mult int) int {
	num := 1
	if len(str) > 1 {
		switch str[len(str)-1] {
		case 'G', 'g':
			num *= mult
			fallthrough
		case 'M', 'm':
			num *= mult
			fallthrough
		case 'K', 'k':
			num *= mult
			str = str[0 : len(str)-1]
		}
	}
	parsed, _ := strconv.Atoi(str)
	return parsed * num
}

func (log Logger) PropToFileLogWriter(props []FilterProp, enabled bool) (*FileLogWriter, bool) {
	file := ""
	format := "[%D %T] [%L] (%S) %M"
	maxlines := 0
	maxsize := 0
	daily := false
	rotate := false
	maxbackup := 999
	maxdays := 0

	// Parse properties
	for _, prop := range props {
		switch prop.Name {
		case "filename":
			file = strings.Trim(prop.Value, " \r\n")
		case "format":
			format = strings.Trim(prop.Value, " \r\n")
		case "maxlines":
			maxlines = StrToNumSuffix(strings.Trim(prop.Value, " \r\n"), 1000)
		case "maxsize":
			maxsize = StrToNumSuffix(strings.Trim(prop.Value, " \r\n"), 1024)
		case "maxdays":
			maxdays = StrToNumSuffix(strings.Trim(prop.Value, " \r\n"), 0)
		case "daily":
			daily = strings.Trim(prop.Value, " \r\n") != "false"
		case "rotate":
			rotate = strings.Trim(prop.Value, " \r\n") != "false"
		case "maxBackup":
			maxbackup = StrToNumSuffix(strings.Trim(prop.Value, " \r\n"), 999)
		default:
			fmt.Fprintf(os.Stderr, "LoadConfiguration Warning: Unknown property \"%s\" for file filter\n", prop.Name)
		}
	}

	// Check properties
	if len(file) == 0 {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Required property \"%s\" for file filter missing\n", "filename")
		return nil, false
	}

	// If it's disabled, we're just checking syntax
	if !enabled {
		return nil, true
	}

	flw := NewFileLogWriter(file, rotate)
	if flw == nil {
		return nil, false
	}
	flw.SetFormat(format)
	flw.SetRotateLines(maxlines)
	flw.SetRotateSize(maxsize)
	flw.SetRotateDays(maxdays)
	flw.SetRotateDaily(daily)
	flw.SetRotateBackup(maxbackup)
	return flw, true
}

func propToSocketLogWriter(props []FilterProp, enabled bool) (*SocketLogWriter, bool) {
	endpoint := ""
	protocol := "udp"

	// Parse properties
	for _, prop := range props {
		switch prop.Name {
		case "endpoint":
			endpoint = strings.Trim(prop.Value, " \r\n")
		case "protocol":
			protocol = strings.Trim(prop.Value, " \r\n")
		default:
			fmt.Fprintf(os.Stderr, "LoadConfiguration Warning: Unknown property \"%s\" for file filter\n", prop.Name)
		}
	}

	// Check properties
	if len(endpoint) == 0 {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Required property \"%s\" for file filter missing\n", "endpoint")
		return nil, false
	}

	// If it's disabled, we're just checking syntax
	if !enabled {
		return nil, true
	}

	return NewSocketLogWriter(protocol, endpoint), true
}
