# log4go

In Production Environment, please consider release Version 5 in the near future. ):

Forked from http://log4go.googlecode.com/

* Sync write, Structured

* Log writer extendable

* Format message with date, time, zone, source, line number

* Fast and buffered log file writer with rotate

* File configuration extendable

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
    defer log.Close()

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
6. <https://github.com/cihub/seelog>
7. <https://github.com/golang/glog>
