// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package xmlog

import (
	l4g "github.com/ccpaging/log4go"
)

// NewXMLLogWriter is a utility method for creating a FileLogWriter set up to
// output XML record log messages instead of line-based ones.
func NewLogWriter(fname string, rotate int) *l4g.FileLogWriter {
	return l4g.NewFileLogWriter(fname, rotate).Set("format", 
`	<record level="%L">
		<timestamp>%D %T</timestamp>
		<source>%S</source>
		<message>%M</message>
	</record>`).Set("head","<log created=\"%D %T\">").Set("foot", "</log>")
}
