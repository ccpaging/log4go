package main

import (
	l4g "github.com/ccpaging/log4go"
	"encoding/json"
	"os"
	"fmt"
)

func main() {
	// Load the configuration (isn't this easy?)
	l4g.LoadConfiguration("config.xml")

	// And now we're ready!
	l4g.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	l4g.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	l4g.Info("About that time, eh chaps?")

	// l4g.Close()

	filename := "config.json"
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
	
    l4g.LoadConfigBuf("config.json", c.LogConfig)
	//l4g.LoadConfiguration("config.json")

	// And now we're ready!
	l4g.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	l4g.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	l4g.Info("About that time, eh chaps?")

	l4g.Close()
}

