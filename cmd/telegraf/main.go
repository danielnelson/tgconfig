package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/influxdata/tgconfig/agent"

	// init inputs
	_ "github.com/influxdata/tgconfig/plugins/inputs/all"
)

var fDebug = flag.Bool("debug", false, "turn on debug logging")
var fRunTimeout = flag.Int("run-timeout", 0, "run for this many seconds")

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

	fmt.Println(*fRunTimeout)
	if fRunTimeout != nil {
		flags.RunTimeout = time.Duration(*fRunTimeout) * time.Second
	}

	agent, err := agent.NewAgent(flags)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = agent.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}
