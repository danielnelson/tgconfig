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

	"github.com/influxdata/tgconfig/plugins/parsers"

	"github.com/BurntSushi/toml"

	telegraf "github.com/influxdata/tgconfig"
	"github.com/influxdata/tgconfig/models"
)

const (
	Name = "toml"
)

func New(config *Config) (telegraf.Loader, error) {
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
	Loaders map[string][]toml.Primitive
}

func (c *Toml) Load(ctx context.Context, plugins *telegraf.Plugins) (*telegraf.Config, error) {
	var (
		conf telegrafConfig
		md   toml.MetaData
		err  error
	)

	if md, err = toml.DecodeFile(c.Config.Path, &conf); err != nil {
		return nil, err
	}

	ri, err := LoadInputs(md, plugins, conf.Inputs)
	if err != nil {
		return nil, err
	}

	ro, err := LoadOutputs(md, plugins, conf.Outputs)
	if err != nil {
		return nil, err
	}

	rl, err := LoadLoaders(md, plugins, conf.Loaders)
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
		Loaders: rl,
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

func LoadInputs(md toml.MetaData, plugins *telegraf.Plugins, inputs map[string][]toml.Primitive) (
	[]*telegraf.InputPlugin,
	error,
) {
	ri := make([]*telegraf.InputPlugin, 0)

	loaders, err := models.Check(plugins.Inputs)
	if err != nil {
		return nil, err
	}

	for name, primitive := range inputs {
		loader, ok := loaders[name]
		if !ok {
			return nil, fmt.Errorf("unknown plugin: %s", name)
		}

		for _, p := range primitive {
			// Parse Input level configuration
			inputConfig := &telegraf.InputConfig{}
			if err := md.PrimitiveDecode(p, inputConfig); err != nil {
				return nil, err
			}

			plugin := loadPlugin(md, p, loader)

			// Parser injection is done for backwards compatibility, new
			// plugins should add the ParserConfig to the plugins config and
			// call NewParser themselves.
			//
			// Not sure if this is actually possible though.  I think we will
			// at the very least have to introduce a new toml syntax.
			//
			// Ideally, the Config could just have a Parser interface and it
			// would be filled before calling New:
			//
			// type Config struct {
			//     Parser telegraf.ParserConfig
			// }
			//
			// func New(config *Config) (telegraf.Input, error) {
			//     // maybe move this?
			//     parser := model.NewParser(config.Parser)
			//     return &Example{parser: parser}
			// }
			if plugin, ok := plugin.(telegraf.ParserInput); ok {
				parser, err := LoadParser(md, p)
				if err != nil {
					return nil, err
				}

				plugin.SetParser(parser)
			}

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

func LoadOutputs(md toml.MetaData, plugins *telegraf.Plugins, outputs map[string][]toml.Primitive) (
	[]*telegraf.OutputPlugin,
	error,
) {
	ro := make([]*telegraf.OutputPlugin, 0)

	loaders, err := models.Check(plugins.Outputs)
	if err != nil {
		return nil, err
	}

	for name, primitive := range outputs {
		loader, ok := loaders[name]
		if !ok {
			return nil, fmt.Errorf("unknown plugin: %s", name)
		}

		for _, p := range primitive {
			// Parse Output level configuration
			outputConfig := &telegraf.OutputConfig{}
			if err := md.PrimitiveDecode(p, outputConfig); err != nil {
				return nil, err
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

func LoadParser(md toml.MetaData, p toml.Primitive) (
	telegraf.Parser,
	error,
) {
	parsers, err := models.Check(parsers.Parsers)
	if err != nil {
		return nil, err
	}

	config := &models.ParserConfig{}
	if err := md.PrimitiveDecode(p, config); err != nil {
		return nil, err
	}

	if config.DataFormat == "" {
		config.DataFormat = "influx"
	}

	parser, ok := parsers[config.DataFormat]
	if !ok {
		return nil, fmt.Errorf("unknown parser: %q", config.DataFormat)
	}

	plugin := loadPlugin(md, p, parser)

	if plugin, ok := plugin.(telegraf.Parser); ok {
		return plugin, nil
	}

	return nil, fmt.Errorf("unexpected plugin type: %s", config.DataFormat)
}

func LoadLoaders(md toml.MetaData, plugins *telegraf.Plugins, configs map[string][]toml.Primitive) (
	[]*telegraf.LoaderPlugin,
	error,
) {
	cp := make([]*telegraf.LoaderPlugin, 0)

	loaders, err := models.Check(plugins.Loaders)
	if err != nil {
		return nil, err
	}

	for name, primitive := range configs {
		loader, ok := loaders[name]
		if !ok {
			return nil, fmt.Errorf("unknown plugin: %s", name)
		}

		for _, p := range primitive {
			plugin := loadPlugin(md, p, loader)

			switch plugin := plugin.(type) {
			case telegraf.Loader:
				clp := &telegraf.LoaderPlugin{Loader: plugin}
				cp = append(cp, clp)
			default:
				return nil, fmt.Errorf("unexpected plugin type: %s", name)
			}

		}
	}
	return cp, nil
}

func (c *Toml) Name() string {
	return Name
}

func (c *Toml) Monitor(ctx context.Context) error {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP)
	defer signal.Stop(signals)

	select {
	case <-signals:
		return telegraf.ReloadConfig
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Toml) MonitorC(ctx context.Context) <-chan error {
	out := make(chan error)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP)

	go func() {
		select {
		case sig := <-signals:
			if sig == syscall.SIGHUP {
				out <- telegraf.ReloadConfig
				break
			}
		case <-ctx.Done():
			out <- ctx.Err()
			break
		}

		signal.Stop(signals)
		close(out)
	}()

	return out
}

// Debugging
func (c *Toml) String() string {
	return "Config: toml"
}
