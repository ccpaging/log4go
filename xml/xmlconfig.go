// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package xmlog

import (
	"encoding/xml"
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

// Parse XML configuration; see examples/example.xml for documentation
func LoadConfigBuf(log l4g.Logger, contents []byte) {
	xc := new(l4g.LogConfig)
	if err := xml.Unmarshal(contents, xc); err != nil {
		fmt.Fprintf(os.Stderr, "LoadConfig: Could not parse XML configuration. %s\n", err)
		os.Exit(1)
	}

	l4g.Close()
	var (
		lw l4g.LogWriter
		flw *l4g.FileLogWriter
		good bool
	)
	for _, fc := range xc.Filters {
		bad, enabled, lvl := log.CheckFilterConfig(fc)
	
		// Just so all of the required attributes are errored at the same time if missing
		if bad {
			os.Exit(1)
		}
	
		if fc.Type == "xml" {
			flw, good = log.PropToFileLogWriter(fc.Properties, enabled)
			if flw != nil && good {
				flw.Set("format", 
`	<record level="%L">
		<timestamp>%D %T</timestamp>
		<source>%S</source>
		<message>%M</message>
	</record>`)
				flw.Set("head","<log created=\"%D %T\">").Set("foot", "</log>")
			}
			lw = flw
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
