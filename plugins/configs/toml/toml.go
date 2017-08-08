package toml

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"

	telegraf "github.com/influxdata/tgconfig"
	"github.com/influxdata/tgconfig/models"
)

const (
	Name = "toml"
)

func New(config *Config) (telegraf.ConfigLoader, error) {
	return &Toml{*config}, nil
}

type Toml struct {
	Config Config
}

type Config struct {
	// Path is the main config file
	Path string
	// Directory is an directory containing config snippets
	Directory string
}

type telegrafConfig struct {
	Agent   telegraf.AgentConfig
	Inputs  map[string][]toml.Primitive
	Outputs map[string][]toml.Primitive
	Configs map[string][]toml.Primitive
}

func (c *Toml) Load(plugins *telegraf.Plugins) (*telegraf.Config, error) {
	var (
		conf telegrafConfig
		md   toml.MetaData
		err  error
	)

	if md, err = toml.DecodeFile(c.Config.Path, &conf); err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("Agent Interval: %d\n", conf.Agent.Interval)

	ri, err := loadInputs(md, plugins, conf)
	if err != nil {
		return nil, err
	}

	ro, err := loadOutputs(md, plugins, conf)
	if err != nil {
		return nil, err
	}

	rc, err := loadConfigs(md, plugins, conf)
	if err != nil {
		return nil, err
	}

	// Only after the entire file is parsed can we report unrecognized plugins.
	for _, item := range md.Undecoded() {
		// Recursive config plugin loading is not allowed.
		if strings.HasPrefix(item.String(), "configs.") {
			continue
		}
		return nil, fmt.Errorf("undecoded toml key: %s", item)
	}

	config := &telegraf.Config{
		Agent:   conf.Agent,
		Inputs:  ri,
		Outputs: ro,
		Configs: rc,
	}

	return config, nil
}

func loadPlugin(md toml.MetaData, p toml.Primitive, factory interface{}) interface{} {
	vfactory := reflect.ValueOf(factory)

	// Get the Type of the first and only argument
	configType := vfactory.Type().In(0)

	// Create a new config struct
	config := reflect.New(configType.Elem()).Interface()

	// Parse TOML into config struct
	if err := md.PrimitiveDecode(p, config); err != nil {
		log.Fatal(err)
	}

	// Call factory with the config struct
	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(config)
	plugin := vfactory.Call(in)[0].Interface()
	return plugin
}

func loadInputs(md toml.MetaData, plugins *telegraf.Plugins, conf telegrafConfig) (
	[]*telegraf.InputPlugin,
	error,
) {
	ri := make([]*telegraf.InputPlugin, 0)

	loaders, err := models.Check(plugins.Inputs)
	if err != nil {
		return nil, err
	}

	for name, primitive := range conf.Inputs {
		loader, ok := loaders[name]
		if !ok {
			return nil, fmt.Errorf("unknown plugin: %s", name)
		}

		for _, p := range primitive {
			// Parse Input level configuration
			inputConfig := &telegraf.InputConfig{}
			if err := md.PrimitiveDecode(p, inputConfig); err != nil {
				log.Fatal(err)
			}

			plugin := loadPlugin(md, p, loader)

			switch plugin := plugin.(type) {
			case telegraf.Input:
				ip := &telegraf.InputPlugin{Input: plugin, Config: inputConfig}
				ri = append(ri, ip)
			default:
				return nil, fmt.Errorf("unexpected plugin type: %s", name)
			}

		}
	}
	return ri, nil
}

func loadOutputs(md toml.MetaData, plugins *telegraf.Plugins, conf telegrafConfig) (
	[]*telegraf.OutputPlugin,
	error,
) {
	ro := make([]*telegraf.OutputPlugin, 0)

	loaders, err := models.Check(plugins.Outputs)
	if err != nil {
		return nil, err
	}

	for name, primitive := range conf.Outputs {
		loader, ok := loaders[name]
		if !ok {
			return nil, fmt.Errorf("unknown plugin: %s", name)
		}

		for _, p := range primitive {
			// Parse Output level configuration
			outputConfig := &telegraf.OutputConfig{}
			if err := md.PrimitiveDecode(p, outputConfig); err != nil {
				log.Fatal(err)
			}

			plugin := loadPlugin(md, p, loader)

			switch plugin := plugin.(type) {
			case telegraf.Output:
				op := &telegraf.OutputPlugin{Output: plugin, Config: outputConfig}
				ro = append(ro, op)
			default:
				return nil, fmt.Errorf("unexpected plugin type: %s", name)
			}

		}
	}
	return ro, nil
}

func loadConfigs(md toml.MetaData, plugins *telegraf.Plugins, conf telegrafConfig) (
	[]*telegraf.ConfigLoaderPlugin,
	error,
) {
	cp := make([]*telegraf.ConfigLoaderPlugin, 0)

	loaders, err := models.Check(plugins.Configs)
	if err != nil {
		return nil, err
	}

	for name, primitive := range conf.Configs {
		loader, ok := loaders[name]
		if !ok {
			return nil, fmt.Errorf("unknown plugin: %s", name)
		}

		for _, p := range primitive {
			plugin := loadPlugin(md, p, loader)

			switch plugin := plugin.(type) {
			case telegraf.ConfigLoader:
				clp := &telegraf.ConfigLoaderPlugin{ConfigLoader: plugin}
				cp = append(cp, clp)
			default:
				return nil, fmt.Errorf("unexpected plugin type: %s", name)
			}

		}
	}
	return cp, nil
}

func (c *Toml) Monitor(ctx context.Context) error {
	// todo: debounce signals
	// todo: don't miss signals, need to run this all the time, need stop?
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP)
	select {
	case sig := <-signals:
		if sig == os.Interrupt {
			return fmt.Errorf("interrupted")
		}
		if sig == syscall.SIGHUP {
			// reload
			return nil
		}
	case <-ctx.Done():
		return nil
	}
	return nil
}

// Debugging
func (c *Toml) String() string {
	return "Config: toml"
}
