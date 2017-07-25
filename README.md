# log4go

Forked from http://log4go.googlecode.com/

Almost redesign.

* Sync write, Structured, Extendable

* Format message with date, time, zone, source, line number

* File config

* Fast log file writer with rotate

* Compatibility with golang `log`

Installation:

- Run `go get github.com/ccpaging/log4go`

OR

- Run `go install github.com/ccpaging/log4go`

Usage:

- Add the following import:

import log "github.com/ccpaging/log4go"

- Sample

```
package main

import (
	log "github.com/ccpaging/log4go"
)

func main() {
    log.Debug("This is Debug")
    log.Info("This is Info")

    // Compatibility with `log`
    log.Print("This is Print()")
    log.Println("This is Println()")
    log.Panic("This is Panic()")
}
```

Acknowledgements:

1. <https://github.com/alecthomas/log4go/>
2. <https://github.com/ngmoco/timber>
3. <https://github.com/siddontang/go/tree/master/log>
4. <https://github.com/sirupsen/logrus>
5. <https://github.com/YoungPioneers/blog4go>
