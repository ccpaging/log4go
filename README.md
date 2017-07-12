# log4go colored

Please see http://log4go.googlecode.com/

Installation:

- Run `go get github.com/ccpaging/log4go`

- Run `go get github.com/daviddengcn/go-colortext`

OR

- Run `go install github.com/ccpaging/log4go`

- Run `go install github.com/daviddengcn/go-colortext`

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

- ccpaging
  For providing awesome patches to bring colored log4go up to the latest Go spec

Reference:

1. <https://github.com/alecthomas/log4go/>
2. <https://github.com/ngmoco/timber>
3. <https://github.com/siddontang/go/tree/master/log>
4. <htps://github.com/sirupsen/logrus>
