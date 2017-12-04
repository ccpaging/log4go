package main

import (
	"path/filepath"
	"fmt"
	"os"
	l4g "github.com/ccpaging/log4go"
	"github.com/ccpaging/log4go/xml"
)

var oldfiles string = "_*.*"

func main() {
	l4g.Close()

	// Load the configuration (isn't this easy?)
	log := l4g.GetGlobalLogger()

	xmlog.LoadConfiguration(log, "config.xml")

	// And now we're ready!
	l4g.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	l4g.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	l4g.Info("About that time, eh chaps?")

	l4g.Close()

	files, _ := filepath.Glob(oldfiles)
	fmt.Printf("%d files match %s\n", len(files), oldfiles) // contains a list of all files in the current directory
	for _, f := range files {
		fmt.Printf("Remove %s\n", f)
		os.Remove(f)
	}
}

