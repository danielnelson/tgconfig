package agent

import (
	"fmt"

	telegraf "github.com/influxdata/tgconfig"
	"github.com/influxdata/tgconfig/models"
	"github.com/influxdata/tgconfig/plugins/configs"
	"github.com/influxdata/tgconfig/plugins/configs/toml"
	"github.com/influxdata/tgconfig/plugins/inputs"
	"github.com/influxdata/tgconfig/plugins/outputs"
)

// Agent is the primary Telegraf struct
type Agent struct {
	flags *Flags
}

type Flags struct {
	Debug bool
	Args  []string
}

func NewAgent(flags *Flags) *Agent {
	return &Agent{flags}
}

// Run starts the main event loop
func (a *Agent) Run() error {
	// Load the base configuration; required and always using the toml config
	// plugin.  This file might contain as little as another config plugin.
	// Global tags need to be passed along.
	var configfile string
	if len(a.flags.Args) > 0 {
		configfile = a.flags.Args[0]
	}

	primaryConfigLoader, err := toml.New(&toml.Config{Path: configfile})
	if err != nil {
		return err
	}

	// dealing with recursion:
	// - only local toml can contain more config plugins? yes but...
	// - could do a top level parse and pass remaining to plugins. could only do with top toml
	// don't provide a way to return config plugins.
	// a plugin could theoretically chain load or whatever for redirection, but it has to load
	// all the plugins.

	plugins := &telegraf.Plugins{
		Configs: configs.Configs,
		Inputs:  inputs.Inputs,
		Outputs: outputs.Outputs,
	}

	// deal with multiple monitors
	// ctx := context.Background()
	for {
		var ris = make([]*models.RunningInput, 0)
		var ros = make([]*models.RunningOutput, 0)
		var rcs = make([]*models.RunningConfig, 0)

		conf, err := primaryConfigLoader.Load(plugins)
		if err != nil {
			return err
		}

		rc := &models.RunningConfig{
			&telegraf.ConfigLoaderPlugin{primaryConfigLoader}}
		fmt.Println(rc.String())
		rcs = append(rcs, rc)

		// Debugging
		for _, input := range conf.Inputs {
			ri := &models.RunningInput{input}
			fmt.Println(ri.String())
			ris = append(ris, ri)
		}

		for _, output := range conf.Outputs {
			ro := &models.RunningOutput{output}
			fmt.Println(ro.String())
			ros = append(ros, ro)
		}

		for _, loader := range conf.Configs {
			conf, err := loader.Load(plugins)
			if err != nil {
				return err
			}

			rc := &models.RunningConfig{loader}
			fmt.Println(rc.String())
			rcs = append(rcs, rc)

			// Debugging
			for _, input := range conf.Inputs {
				ri := &models.RunningInput{input}
				fmt.Println(ri.String())
				ris = append(ris, ri)
			}

			// Debugging
			for _, output := range conf.Outputs {
				ro := &models.RunningOutput{output}
				fmt.Println(ro.String())
				ros = append(ros, ro)
			}
		}

		// ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		// defer cancel()

		// err = primaryConfigLoader.Monitor(ctx)
		// if err != nil {
		// 	fmt.Println(err)
		// 	break
		// }

		// if ctx.Err() == context.DeadlineExceeded {
		// 	fmt.Println("finished timed run")
		// 	break
		// }

		// fmt.Println("reloading")

		break
	}

	// run stuff

	fmt.Println("Run -- finished")
	return nil
}

// Restart triggers a plugin reload
func (a *Agent) Reload() error {
	// Pause Inputs
	// Flush Outputs
	// Run Inputs acks
	// Clear all inputs, processors, aggregators, outputs
	// Do all that but also keep buffers
	// Reload config and start
	return nil
}

// Shutdown stops the Agent
func (a *Agent) Shutdown() {
	// Pause Inputs
	// Flush Outputs
	// Run Inputs acks
	// Clear all inputs, processors, aggregators, outputs
	// Stop
}
