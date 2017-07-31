// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package log4go

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"errors"
)

// Various error codes.
var (
	ErrBadOption   = errors.New("invalid or unsupported option")
	ErrBadValue    = errors.New("invalid option value")
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
	clw := NewConsoleLogWriter()
	// Parse properties
	for _, prop := range props {
		err := clw.SetOption(prop.Name, strings.Trim(prop.Value, " \r\n"))
		if err != nil { 
			fmt.Fprintf(os.Stderr, "Console filter Warning: \"%s\", %v\n", prop.Name, err)
		}
	}

	// If it's disabled, we're just checking syntax
	if !enabled {
		return nil, true
	}

	return clw, true
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

// cycle, delay0, time.Duration. Parse a duration string.
// A duration string is a possibly signed sequence of decimal numbers,
// each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us", "ms", "s", "m", "h".
func (log Logger) PropToFileLogWriter(props []FilterProp, enabled bool) (*FileLogWriter, bool) {
	filename := ""
	rotate := 0
	cycle := "24h"
	delay0 := "0h"
	format := "[%D %T] [%L] (%S) %M"
	flush := 0
	maxsize := "10M"

	// Parse properties
	for _, prop := range props {
		switch prop.Name {
		case "filename":
			filename = strings.Trim(prop.Value, " \r\n")
		case "rotate":
			rotate = StrToNumSuffix(strings.Trim(prop.Value, " \r\n"), 1)
		case "cycle":
			maxsize = strings.Trim(prop.Value, " \r\n")
		case "delay0":
			delay0 = strings.Trim(prop.Value, " \r\n")
		case "format":
			format = strings.Trim(prop.Value, " \r\n")
		case "flush":
			flush = StrToNumSuffix(strings.Trim(prop.Value, " \r\n"), 1024)
		case "maxsize":
			maxsize = strings.Trim(prop.Value, " \r\n")
		default:
			fmt.Fprintf(os.Stderr, "LoadConfiguration Warning: Unknown property \"%s\" for file filter\n", prop.Name)
		}
	}

	// Check properties
	if len(filename) == 0 {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Required property \"%s\" for file filter missing\n", "filename")
		return nil, false
	}

	// If it's disabled, we're just checking syntax
	if !enabled {
		return nil, true
	}

	flw := NewFileLogWriter(filename, rotate).Set("cycle", cycle).Set("delay0", delay0)
	if flw == nil {
		return nil, false
	}
	flw.SetOption("format", format)
	flw.SetOption("flush", flush)
	flw.SetOption("maxsize", maxsize)
	return flw, true
}

func propToSocketLogWriter(props []FilterProp, enabled bool) (*SocketLogWriter, bool) {
	endpoint := ""
	protocol := "udp"
	format := "[%D %T] [%L] (%S) %M"

	// Parse properties
	for _, prop := range props {
		switch prop.Name {
		case "endpoint":
			endpoint = strings.Trim(prop.Value, " \r\n")
		case "protocol":
			protocol = strings.Trim(prop.Value, " \r\n")
		case "format":
			format = strings.Trim(prop.Value, " \r\n")
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

	return NewSocketLogWriter(protocol, endpoint).Set("format", format), true
}
