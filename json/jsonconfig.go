// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package jsonlog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	l4g "github.com/ccpaging/log4go"
)

func LoadConfiguration(log l4g.Logger, filename string) {
	if len(filename) <= 0 {
		return
	}

	// Open the configuration file
	fd, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfig: Error: Could not open %q for reading: %s\n", filename, err)
		os.Exit(1)
	}
	defer fd.Close()

	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfig: Error: Could not read %q: %s\n", filename, err)
		os.Exit(1)
	}

	LoadConfigBuf(log, buf)
}

// Parse JSON configuration; see examples/logconfig.json for documentation
func LoadConfigBuf(log l4g.Logger, contents []byte) {
	jc := new(l4g.LogConfig)
	if err := json.Unmarshal(contents, jc); err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfiguration: Could not parse Json LogConfiguration. %s\n", err)
		os.Exit(1)
	}
	l4g.Close()

	var (
		lw l4g.LogWriter
		good bool
	)
	for _, fc := range jc.Filters {
		bad, enabled, lvl := log.CheckFilterConfig(fc)
	
		// Just so all of the required attributes are errored at the same time if missing
		if bad {
			os.Exit(1)
		}
	
		if fc.Type == "json" {
			lw = NewLogWriter(l4g.DefaultSockProto, l4g.DefaultSockEndPoint)
			_, good = log.ConfigLogWriter(lw, fc.Properties)
		} else {
			lw, good = log.MakeLogWriter(fc, enabled)
		}
	
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
