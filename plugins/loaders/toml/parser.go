package toml

import (
	"fmt"
	"io"
	"log"
	"reflect"

	"github.com/influxdata/tgconfig/models"
	"github.com/influxdata/tgconfig/plugins/parsers"

	"github.com/BurntSushi/toml"

	telegraf "github.com/influxdata/tgconfig"
)

type parser struct {
	md       toml.MetaData
	registry *telegraf.ConfigRegistry
}

func NewParser(registry *telegraf.ConfigRegistry) *parser {
	return &parser{registry: registry}
}

func (p *parser) Parse(reader io.Reader) (*telegraf.Config, error) {
	var err error
	conf := struct {
		Agent   telegraf.AgentConfig
		Inputs  map[string][]toml.Primitive
		Outputs map[string][]toml.Primitive
		Loaders map[string][]toml.Primitive
	}{}

	if p.md, err = toml.DecodeReader(reader, &conf); err != nil {
		return nil, err
	}

	ri, err := p.loadInputs(conf.Inputs)
	if err != nil {
		return nil, err
	}

	ro, err := p.loadOutputs(conf.Outputs)
	if err != nil {
		return nil, err
	}

	rl, err := p.loadLoaders(conf.Loaders)
	if err != nil {
		return nil, err
	}

	// Now that we have tried to parse the entire file we report unrecognized plugins.
	for _, item := range p.md.Undecoded() {
		// Recursive config plugin loading is not allowed.
		// edit: Still error though?
		// if strings.HasPrefix(item.String(), "loaders.") {
		// 	continue
		// }
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

func (p *parser) loadInputs(inputs map[string][]toml.Primitive) (map[string][]*telegraf.InputPlugin, error) {
	configs := make(map[string][]*telegraf.InputPlugin)

	// Function on Registry?
	// Don't call this loader anymore
	loaders, err := models.Check(p.registry.Inputs)
	if err != nil {
		return nil, err
	}

	for name, primitive := range inputs {
		loader, ok := loaders[name]
		if !ok {
			return nil, fmt.Errorf("unknown plugin: %s", name)
		}

		for _, prim := range primitive {
			// Parse common Input level configuration
			inputConfig := &telegraf.InputConfig{}
			if err := p.md.PrimitiveDecode(prim, inputConfig); err != nil {
				return nil, err
			}

			// Parse specific Input configuration
			config := p.loadConfig(prim, loader)

			cur, ok := configs[name]
			if !ok {
				cur = make([]*telegraf.InputPlugin, 0)
			}
			plugin := &telegraf.InputPlugin{
				InputConfig: inputConfig,
				Config:      config,
			}
			configs[name] = append(cur, plugin)
		}
	}
	return configs, nil
}

func (p *parser) loadOutputs(outputs map[string][]toml.Primitive) (map[string][]*telegraf.OutputPlugin, error) {
	configs := make(map[string][]*telegraf.OutputPlugin)

	loaders, err := models.Check(p.registry.Outputs)
	if err != nil {
		return nil, err
	}

	for name, primitive := range outputs {
		loader, ok := loaders[name]
		if !ok {
			return nil, fmt.Errorf("unknown plugin: %s", name)
		}

		for _, prim := range primitive {
			// Parse Output level configuration
			outputConfig := &telegraf.OutputConfig{}
			if err := p.md.PrimitiveDecode(prim, outputConfig); err != nil {
				return nil, err
			}

			// Parse specific configuration
			config := p.loadConfig(prim, loader)

			cur, ok := configs[name]
			if !ok {
				cur = make([]*telegraf.OutputPlugin, 0)
			}

			plugin := &telegraf.OutputPlugin{
				OutputConfig: outputConfig,
				Config:       config,
			}
			configs[name] = append(cur, plugin)
		}
	}
	return configs, nil
}

func (p *parser) loadLoaders(loaders map[string][]toml.Primitive) (map[string][]*telegraf.LoaderPlugin, error) {
	configs := make(map[string][]*telegraf.LoaderPlugin, 0)

	loaderConfigs, err := models.Check(p.registry.Loaders)
	if err != nil {
		return nil, err
	}

	for name, primitive := range loaders {
		loader, ok := loaderConfigs[name]
		if !ok {
			return nil, fmt.Errorf("unknown plugin: %s", name)
		}

		for _, prim := range primitive {
			loaderConfig := &telegraf.LoaderConfig{}
			if err := p.md.PrimitiveDecode(prim, loaderConfig); err != nil {
				return nil, err
			}

			// Parse specific configuration
			config := p.loadConfig(prim, loader)

			cur, ok := configs[name]
			if !ok {
				cur = make([]*telegraf.LoaderPlugin, 0)
			}

			plugin := &telegraf.LoaderPlugin{
				LoaderConfig: loaderConfig,
				Config:       config,
			}
			configs[name] = append(cur, plugin)
		}
	}
	return configs, nil
}

func (p *parser) loadConfig(prim toml.Primitive, factory interface{}) interface{} {
	vfactory := reflect.ValueOf(factory)

	// Get the Type of the first and only argument
	configType := vfactory.Type().In(0)

	// Create a new config struct
	config := reflect.New(configType.Elem()).Interface()

	// Parse TOML into config struct
	if err := p.md.PrimitiveDecode(prim, config); err != nil {
		log.Fatal(err)
	}
	return config
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
