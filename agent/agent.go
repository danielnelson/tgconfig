package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"time"

	telegraf "github.com/influxdata/tgconfig"
	"github.com/influxdata/tgconfig/models"
	"github.com/influxdata/tgconfig/plugins/inputs"
	"github.com/influxdata/tgconfig/plugins/loaders"
	"github.com/influxdata/tgconfig/plugins/loaders/toml"
	"github.com/influxdata/tgconfig/plugins/outputs"
	"github.com/influxdata/tgconfig/plugins/parsers"
)

// Agent represents the main event loop
type Agent struct {
	flags *Flags
}

// Flags are the initialization options that cannot be changed
//
// The Agent also loads an AgentConfig which can be modified during runtime.
type Flags struct {
	Debug bool
	Args  []string
}

func NewAgent(flags *Flags) *Agent {
	return &Agent{flags}
}

func createBuiltinLoader(path string, registry telegraf.FactoryRegistry) (*models.RunningLoader, error) {
	config := &telegraf.LoaderConfig{
		Config:       &telegraf.CommonLoaderConfig{},
		PluginConfig: &toml.Config{Path: path},
	}

	return models.NewRunningLoader("toml", config, registry)
}

func createPlugin(config interface{}, factory interface{}) interface{} {
	vfactory := reflect.ValueOf(factory)

	// Call factory with the config struct
	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(config)
	plugin := vfactory.Call(in)[0].Interface()
	return plugin
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

	// make factories own configs
	registry, _ := models.NewFactories(
		loaders.Loaders,
		inputs.Inputs,
		outputs.Outputs,
		parsers.Parsers,
	)
	// call once or many times?
	configr := registry.GetConfigRegistry()

	builtinLoader, err := createBuiltinLoader(configfile, registry)
	if err != nil {
		return err
	}

	// dealing with recursion:
	// - only local toml can contain more config plugins? yes but...
	// - could do a top level parse and pass remaining to plugins. could only do with top toml
	// don't provide a way to return config plugins.
	// a plugin could theoretically chain load or whatever for redirection, but it has to load
	// all the plugins.

	var wg sync.WaitGroup

	ctx := context.Background()
	ctx, sigcancel := context.WithCancel(ctx)

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case sig := <-signals:
			if sig == os.Interrupt {
				fmt.Println("interrupt: agent")
				break
			}
		case <-ctx.Done():
			break
		}
		signal.Stop(signals)
		sigcancel()
	}()

	for {
		var ris = make([]*models.RunningInput, 0)
		var ros = make([]*models.RunningOutput, 0)
		var rls = make([]*models.RunningLoader, 0)

		// !! Run for maximum amount of time; this is for development reasons but
		// maybe it should become an official option?  Although we probably
		// should exclude config loading time.
		ctx, cancel := context.WithTimeout(ctx, 200*time.Second)
		defer cancel()

		conf, err := builtinLoader.Load(ctx, configr)
		if err != nil {
			return err
		}

		rls = append(rls, builtinLoader)

		fmt.Printf(FormatPlugin(builtinLoader))

		// Begin monitoring all plugins for changes, by monitoring before
		// loading we ensure the config can never be stale.
		//
		// !! We don't want to continue until all are started, but the current
		// interface doesn't allow us to know when this happened.
		// var watcher = newMonitor()
		var watcher = newWatcher()
		watcher.WatchLoader(ctx, builtinLoader)

		for name, configs := range conf.Inputs {
			for _, config := range configs {
				ri, err := models.NewRunningInput(name, config, registry)
				if err != nil {
					// what do
					return err
				}
				ris = append(ris, ri)

				fmt.Printf(FormatPlugin(ri))
			}
		}
		for name, configs := range conf.Outputs {
			for _, config := range configs {
				ro, err := models.NewRunningOutput(name, config, registry)
				if err != nil {
					// what do
				}
				ros = append(ros, ro)

				fmt.Printf(FormatPlugin(ro))
			}
		}

		for name, configs := range conf.Loaders {
			var conf *telegraf.Config
			for _, config := range configs {
				rl, err := models.NewRunningLoader(name, config, registry)
				if err != nil {
					// what do
				}
				rls = append(rls, rl)

				watcher.WatchLoader(ctx, rl)

				conf, err = rl.Load(ctx, configr)
				if err != nil {
					return err
				}

				fmt.Printf(FormatPlugin(rl))
			}

			for name, configs := range conf.Inputs {
				for _, config := range configs {
					ri, err := models.NewRunningInput(name, config, registry)
					if err != nil {
						// what do
						return err
					}
					ris = append(ris, ri)

					fmt.Printf(FormatPlugin(ri))
				}
			}
			for name, configs := range conf.Outputs {
				for _, config := range configs {
					ro, err := models.NewRunningOutput(name, config, registry)
					if err != nil {
						// what do
					}
					ros = append(ros, ro)

					fmt.Printf(FormatPlugin(ro))
				}
			}
		}

		// !! Start Pipeline
		// Wait for Watch to complete
		watcher.Wait()
		// !! Stop Pipeline

		if ctx.Err() == context.Canceled {
			fmt.Println("cancelled: agent")
			break
		}
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("finished timed run: agent")
			break
		}

		fmt.Println("reloading")
	}

	fmt.Println("Run -- finished")
	sigcancel()
	wg.Wait()
	return nil
}

type Watcher struct {
	wg      sync.WaitGroup
	cancels []context.CancelFunc
	done    chan struct{}
	once    sync.Once
}

func newWatcher() *Watcher {
	return &Watcher{
		done: make(chan struct{}),
	}
}

func (m *Watcher) WatchLoader(ctx context.Context, loader *models.RunningLoader) error {
	ctx, cancel := context.WithCancel(ctx)
	m.cancels = append(m.cancels, cancel)

	waiter, err := loader.Watch(ctx)
	if err != nil {
		return err
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		err := waiter.Wait()

		if ctx.Err() == context.Canceled {
			fmt.Printf("cancelled: %T\n", loader)
		} else if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("timeout: %T\n", loader)
		} else if err != nil {
			fmt.Printf("%s: %T\n", err, loader)
		} else {
			fmt.Printf("monitor completed without error: %T\n", loader)
		}
		m.once.Do(func() { close(m.done) })
	}()
	return nil
}

func (m *Watcher) Wait() error {
	select {
	case <-m.done:
		for _, cancel := range m.cancels {
			cancel()
		}

	}

	m.wg.Wait()
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

func FormatPlugin(p interface{}) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetIndent("", "    ")
	err := enc.Encode(p)
	if err != nil {
		fmt.Println(err)
	}
	return b.String()

}
