package main

import (
	"encoding/json"
	"os"
	"fmt"
	l4g "github.com/ccpaging/log4go"
	"github.com/ccpaging/log4go/json"
)

var filename string = "logconfig.json"

func main() {
	l4g.Close()

	// Load the configuration (isn't this easy?)
	log := l4g.GetLogger()

	fd, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("Can't load json config file: %s %v", filename, err))
	}
	defer fd.Close()

	type Config struct {
		LogConfig json.RawMessage
	}

	c := Config{}
	err = json.NewDecoder(fd).Decode(&c)
	if err != nil {
		panic(fmt.Sprintf("Can't parse json config file: %s %v", filename, err))
	}
	
	jsonlog.LoadConfigBuf(log, c.LogConfig)

	//l4g.LoadConfiguration(filename)

	// And now we're ready!
	l4g.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	l4g.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	l4g.Info("About that time, eh chaps?")

	l4g.Close()

	os.Remove("_test.log")
}

