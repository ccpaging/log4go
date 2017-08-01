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
		fmt.Fprintf(os.Stderr, 
			"LoadConfiguration: Required child <%s> for filter has unknown value. %s\n", 
			"level", fc.Level)
		bad = true
	}
	return bad, enabled, lvl
}

func (log Logger) MakeLogWriter(fc FilterConfig, enabled bool) (LogWriter, bool) {
	var (
		lw LogWriter
	)
	switch fc.Type {
	case "console":
		lw = NewConsoleLogWriter()
	case "file":
		lw = NewFileLogWriter(DefaultFileName, 0)
	case "socket":
		lw = NewSocketLogWriter(DefaultSockProto, DefaultSockEndPoint)
	default:
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Unknown filter type \"%s\"\n", fc.Type)
		return nil, false
	}

	_, good := log.ConfigLogWriter(lw, fc.Properties)
	if !good {
		return nil, false
	}

	// If it's disabled, we're just checking syntax
	if !enabled {
		return nil, true
	}

	return lw, good
}

func (log Logger) ConfigLogWriter(lw LogWriter, props []FilterProp) (LogWriter, bool) {
	good := true
	for _, prop := range props {
		err := lw.SetOption(prop.Name, strings.Trim(prop.Value, " \r\n"))
		if err != nil {
			switch err {
			case ErrBadValue:
				fmt.Fprintf(os.Stderr, "LoadConfiguration: console filter, Bad value of \"%s\"\n", prop.Name)
				good = false
			case ErrBadOption:
				fmt.Fprintf(os.Stderr, "LoadConfiguration: console filter, Unknown property \"%s\"\n", prop.Name)
			default:
			}
		}
	}
	return lw, good
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
