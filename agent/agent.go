package agent

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	telegraf "github.com/influxdata/tgconfig"
	"github.com/influxdata/tgconfig/models"
	"github.com/influxdata/tgconfig/plugins/inputs"
	"github.com/influxdata/tgconfig/plugins/loaders"
	"github.com/influxdata/tgconfig/plugins/loaders/toml"
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

	primaryLoader, err := toml.New(&toml.Config{Path: configfile})
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
		Loaders: loaders.Loaders,
		Inputs:  inputs.Inputs,
		Outputs: outputs.Outputs,
	}

	ctx := context.Background()
	for {
		var ris = make([]*models.RunningInput, 0)
		var ros = make([]*models.RunningOutput, 0)
		var rls = make([]*models.RunningLoader, 0)

		conf, err := primaryLoader.Load(ctx, plugins)
		if err != nil {
			return err
		}

		rl := &models.RunningLoader{
			LoaderPlugin: &telegraf.LoaderPlugin{Loader: primaryLoader}}
		fmt.Println(rl.String())
		rls = append(rls, rl)

		// Debugging
		for _, input := range conf.Inputs {
			ri := &models.RunningInput{InputPlugin: input}
			fmt.Println(ri.String())
			ris = append(ris, ri)
		}

		for _, output := range conf.Outputs {
			ro := &models.RunningOutput{OutputPlugin: output}
			fmt.Println(ro.String())
			ros = append(ros, ro)
		}

		for _, loader := range conf.Loaders {
			conf, err := loader.Load(ctx, plugins)
			if err != nil {
				return err
			}

			rc := &models.RunningLoader{LoaderPlugin: loader}
			fmt.Println(rc.String())
			rls = append(rls, rl)

			// Debugging
			for _, input := range conf.Inputs {
				ri := &models.RunningInput{InputPlugin: input}
				fmt.Println(ri.String())
				ris = append(ris, ri)
			}

			// Debugging
			for _, output := range conf.Outputs {
				ro := &models.RunningOutput{OutputPlugin: output}
				fmt.Println(ro.String())
				ros = append(ros, ro)
			}
		}

		ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		sigctx, sigcancel := context.WithCancel(ctx)
		defer sigcancel()
		go func(ctx context.Context) {
			signals := make(chan os.Signal)
			signal.Notify(signals, os.Interrupt)
			select {
			case sig := <-signals:
				if sig == os.Interrupt {
					fmt.Println("interrupt: agent")
					cancel()
					break
				}
			case <-ctx.Done():
				cancel()
				break
			}
			signal.Stop(signals)
		}(sigctx)

		// Maybe we should begin monitoring before loading (except for
		// primary).
		Monitor(ctx, rls)

		if ctx.Err() == context.Canceled {
			fmt.Println("cancelled: agent")
			break
		}
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("finished timed run: agent")
			break
		}

		sigcancel()

		fmt.Println("reloading")
	}

	fmt.Println("Run -- finished")
	return nil
}

func Monitor(ctx context.Context, rcs []*models.RunningLoader) error {
	var wg sync.WaitGroup
	var cancels []context.CancelFunc
	var once sync.Once

	done := make(chan struct{}, len(rcs))

	for _, rc := range rcs {
		ctx, cancel := context.WithCancel(ctx)
		cancels = append(cancels, cancel)

		wg.Add(1)
		go func(rc *models.RunningLoader) {
			defer wg.Done()
			err := rc.Monitor(ctx)

			if ctx.Err() == context.Canceled {
				fmt.Printf("cancelled: %s\n", rc.Name())
			} else if ctx.Err() == context.DeadlineExceeded {
				fmt.Printf("timeout: %s\n", rc.Name())
			} else if err == telegraf.ReloadConfig {
				fmt.Printf("%s: %s\n", err, rc.Name())
			} else if err != nil {
				fmt.Println(err)
			}
			once.Do(func() { close(done) })
		}(rc)
	}

	select {
	case <-done:
	}

	for _, cancel := range cancels {
		cancel()
	}

	wg.Wait()
	return nil
}

func MonitorC(ctx context.Context, rcs []*models.RunningLoader) error {
	var wg sync.WaitGroup
	var cancels []context.CancelFunc

	// ugg, what if someone sends more than one message?
	// do i need a semaphore?
	done := make(chan error, len(rcs))

	for _, rc := range rcs {
		wg.Add(1)
		ctx, cancel := context.WithCancel(ctx)
		cancels = append(cancels, cancel)
		go func(rc *models.RunningLoader) {
			defer wg.Done()
			select {
			case err := <-rc.MonitorC(ctx):
				if ctx.Err() == context.Canceled {
					fmt.Printf("cancelled: %s\n", rc.Name())
				} else if ctx.Err() == context.DeadlineExceeded {
					fmt.Printf("timeout: %s\n", rc.Name())
				} else if err == telegraf.ReloadConfig {
					fmt.Printf("%s: %s\n", err, rc.Name())
				} else if err != nil {
					fmt.Println(err)
				}
				done <- err
				return
			}
		}(rc)
	}

	select {
	case <-done:
		for _, c := range cancels {
			c()
		}
	}

	wg.Wait()
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
