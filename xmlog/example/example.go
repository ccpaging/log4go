package main

import (
	l4g "github.com/ccpaging/log4go"
	"github.com/ccpaging/log4go/xmlog"
)

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
}

