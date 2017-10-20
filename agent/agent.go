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
	flags      *Flags
	registry   telegraf.Registry
	mainLoader *models.RunningLoader
}

// Flags are the initialization options that cannot be changed
//
// The Agent also loads an AgentConfig which can be modified during runtime.
type Flags struct {
	Debug      bool
	RunTimeout time.Duration
	Args       []string
}

func NewAgent(flags *Flags) (*Agent, error) {
	registry, err := models.NewRegistry(
		loaders.Loaders,
		inputs.Inputs,
		outputs.Outputs,
		parsers.Parsers,
	)
	if err != nil {
		return nil, err
	}

	// Load the base configuration; required and always using the toml config
	// plugin.  This file might contain as little as another config plugin.
	// Global tags need to be passed along.
	var configfile string
	if len(flags.Args) > 0 {
		configfile = flags.Args[0]
	}

	mainLoader, err := createMainLoader(configfile, registry)
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		flags:      flags,
		registry:   registry,
		mainLoader: mainLoader,
	}

	return agent, nil
}

func createMainLoader(path string, registry telegraf.Registry) (*models.RunningLoader, error) {
	config := &telegraf.LoaderConfig{
		Config:       &telegraf.CommonLoaderConfig{},
		PluginConfig: &toml.Config{Path: path},
	}

	loaders, err := models.NewRunningLoaders("toml", config, registry)
	if err != nil {
		return nil, err
	}
	return loaders[0], nil
}

func createPlugin(config interface{}, factory interface{}) interface{} {
	vfactory := reflect.ValueOf(factory)

	// Call factory with the config struct
	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(config)
	plugin := vfactory.Call(in)[0].Interface()
	return plugin
}

type Pipeline struct {
	Acc     telegraf.Accumulator
	Inputs  []*models.RunningInput
	Outputs []*models.RunningOutput
	Loaders []*models.RunningLoader
}

func NewPipeline() *Pipeline {
	p := &Pipeline{}
	p.Acc = &models.Accumulator{}
	p.Inputs = make([]*models.RunningInput, 0)
	p.Outputs = make([]*models.RunningOutput, 0)
	p.Loaders = make([]*models.RunningLoader, 0)
	return p
}

func (p *Pipeline) AddInputs(inputs ...*models.RunningInput) {
	p.Inputs = append(p.Inputs, inputs...)
}

func (p *Pipeline) AddOutputs(outputs ...*models.RunningOutput) {
	p.Outputs = append(p.Outputs, outputs...)
}

func (p *Pipeline) AddLoaders(loaders ...*models.RunningLoader) {
	p.Loaders = append(p.Loaders, loaders...)
}

// Run starts the main event loop
func (a *Agent) Run() error {
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

	// Might want another timeout for run-timeout after loaded
	if a.flags.RunTimeout > time.Second*0 {
		fmt.Printf("Setting run timeout: %s\n", a.flags.RunTimeout)
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.flags.RunTimeout)
		defer cancel()
	}

	for {
		var watcher = NewWatcher()
		pipeline, err := a.LoadPipeline(ctx, watcher)
		if err != nil {
			fmt.Println(err)
			break
		}

		for _, input := range pipeline.Inputs {
			fmt.Printf(FormatPlugin(input))
		}
		for _, output := range pipeline.Outputs {
			fmt.Printf(FormatPlugin(output))
		}
		for _, loader := range pipeline.Loaders {
			fmt.Printf(FormatPlugin(loader))
		}

		// !! Start Pipeline
		// Wait for Watch to complete
		watcher.Wait()
		fmt.Println("Watch Triggered")
		// !! Stop Pipeline

		if ctx.Err() == context.Canceled {
			fmt.Println("cancelled: agent")
		}
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("finished timed run: agent")
		}
	}

	fmt.Println("Run -- finished")
	sigcancel()
	wg.Wait()
	return nil
}

func (a *Agent) LoadPipeline(ctx context.Context, watcher *watcher) (*Pipeline, error) {
	var pipeline = NewPipeline()

	configreg := a.registry.GetConfigRegistry()

	// Place a watch on the main loader before loading, ensuring that we don't
	// miss any updates.
	watcher.WatchLoader(ctx, a.mainLoader)

	fmt.Printf("Loading: %s\n", a.mainLoader.Name)
	conf, err := a.mainLoader.Load(ctx, configreg)
	if err != nil {
		return nil, err
	}
	pipeline.AddLoaders(a.mainLoader)

	for name, configs := range conf.Inputs {
		for _, config := range configs {
			inputs, err := models.NewRunningInputs(name, config, a.registry)
			if err != nil {
				return nil, err
			}
			pipeline.AddInputs(inputs...)
		}
	}

	for name, configs := range conf.Outputs {
		for _, config := range configs {
			outputs, err := models.NewRunningOutputs(name, config, a.registry)
			if err != nil {
				return nil, err
			}
			pipeline.AddOutputs(outputs...)
		}
	}

	for name, configs := range conf.Loaders {
		var conf *telegraf.Config
		for _, config := range configs {
			loaders, err := models.NewRunningLoaders(name, config, a.registry)
			if err != nil {
				return nil, err
			}
			pipeline.AddLoaders(loaders...)

			for _, loader := range loaders {
				watcher.WatchLoader(ctx, loader)

				fmt.Printf("Loading: %s\n", name)
				conf, err = loader.Load(ctx, configreg)
				if err != nil {
					return nil, err
				}
			}
		}

		for name, configs := range conf.Inputs {
			for _, config := range configs {
				inputs, err := models.NewRunningInputs(name, config, a.registry)
				if err != nil {
					return nil, err
				}
				pipeline.AddInputs(inputs...)
			}
		}

		for name, configs := range conf.Outputs {
			for _, config := range configs {
				outputs, err := models.NewRunningOutputs(name, config, a.registry)
				if err != nil {
					return nil, err
				}
				pipeline.AddOutputs(outputs...)
			}
		}
	}

	return pipeline, nil
}

type watcher struct {
	wg      sync.WaitGroup
	cancels []context.CancelFunc
	done    chan struct{}
	once    sync.Once
}

func NewWatcher() *watcher {
	return &watcher{
		done: make(chan struct{}),
	}
}

func (m *watcher) WatchLoader(ctx context.Context, loader *models.RunningLoader) error {
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

		var name = loader.Name

		if ctx.Err() == context.Canceled {
			fmt.Printf("cancelled: %s\n", name)
		} else if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("timeout: %s\n", name)
		} else if err != nil {
			fmt.Printf("%v: %s\n", err, name)
		} else {
			fmt.Printf("monitor completed without error: %s\n", name)
		}
		m.once.Do(func() { close(m.done) })
	}()
	return nil
}

func (m *watcher) Wait() error {
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
	return nil

	// Alternative way:
	// load_new_configs()
	// if error:
	//    # note: includes errors on New
	//    print error and abandon reload
	//
	// try connect new outputs
	// if error:
	//    print error and abandon reload
	//
	// swap_runtime_state()
	// pause_inputs() # doesn't exist today -- future ack work
	//
	// async start new inputs()
	//
	// flush_aggregators()
	// flush_processors()
	// flush_outputs()
	// run_input_acks() # doesn't exist today -- future ack work
	// delete_old_runtime_state()

	// Needs plenty of thought about what we do when the config cannot be
	// loaded.

	// Remain cancellable...
}

// Shutdown stops the Agent
func (a *Agent) Shutdown() {
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
