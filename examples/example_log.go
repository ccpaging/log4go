// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	log "github.com/ccpaging/log4go"
)

func main() {
	log.SetFlags(log.Lshortfile)
	log.Output(log.DefaultCallerSkip, "Hello, log file!")
	log.Print("Hello, log file!")
	log.Close()

	var buf bytes.Buffer
	logger := log.New(&buf, "logger: ", log.Lshortfile)
	logger.Print("Hello, log file!")
	// logger.Panic("Hello, log file!")
	logger.Close()

	fmt.Print(&buf)
	// Output:
	// logger: example.go:21: Hello, log file!
}
