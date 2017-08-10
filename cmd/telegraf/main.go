package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/influxdata/tgconfig/agent"
)

var fDebug = flag.Bool("debug", false, "turn on debug logging")

func main() {
	// Parse cli flags; these can never be modified, any other piece of
	// configuration can be updated.  Because of the immutability, many of the
	// options should be changeable via the agent config.
	flag.Usage = func() { os.Exit(0) }
	flag.Parse()
	args := flag.Args()

	flags := &agent.Flags{
		Debug: *fDebug,
		Args:  args,
	}

	agent := agent.NewAgent(flags)
	err := agent.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
