package toml

import (
	"fmt"
	"io"

	"github.com/BurntSushi/toml"

	telegraf "github.com/influxdata/tgconfig"
)

type parser struct {
	md       toml.MetaData
	registry telegraf.ConfigRegistry
}

func NewParser(registry telegraf.ConfigRegistry) *parser {
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

func (p *parser) loadInputs(inputs map[string][]toml.Primitive) (map[string][]*telegraf.InputConfig, error) {
	inputConfigs := make(map[string][]*telegraf.InputConfig)

	// Function on Registry?
	// Don't call this loader anymore
	// factories, err := models.Check(p.registry.Inputs)
	// if err != nil {
	// 	return nil, err
	// }

	for name, primitives := range inputs {
		configs := make([]*telegraf.InputConfig, 0)
		for _, primitive := range primitives {
			pluginConfig, ok := p.registry.GetPluginConfig(telegraf.InputType, name)
			if !ok {
				return nil, fmt.Errorf("unknown input plugin: %s", name)
			}

			// Parse specific configuration
			if err := p.md.PrimitiveDecode(primitive, pluginConfig); err != nil {
				return nil, err
			}

			// Parse common configuration
			commonConfig := &telegraf.CommonInputConfig{}
			if err := p.md.PrimitiveDecode(primitive, commonConfig); err != nil {
				return nil, err
			}

			// We don't know if this plugin will have a parser until we new
			// it, so we will just always try to load a parser just in case.
			dataFormat := commonConfig.DataFormat
			if dataFormat == "" {
				dataFormat = "influx"
			}

			// Parse parser configuration
			parserConfig, ok := p.registry.GetPluginConfig(telegraf.ParserType, dataFormat)
			if err := p.md.PrimitiveDecode(primitive, parserConfig); err != nil {
				return nil, err
			}

			plugin := &telegraf.InputConfig{
				Config:       commonConfig,
				PluginConfig: pluginConfig,
				ParserConfig: parserConfig,
			}
			configs = append(configs, plugin)
		}
		inputConfigs[name] = configs
	}
	return inputConfigs, nil
}

func (p *parser) loadOutputs(outputs map[string][]toml.Primitive) (map[string][]*telegraf.OutputConfig, error) {
	outputConfigs := make(map[string][]*telegraf.OutputConfig)

	for name, primitives := range outputs {
		configs := make([]*telegraf.OutputConfig, 0)
		for _, primitive := range primitives {
			pluginConfig, ok := p.registry.GetPluginConfig(telegraf.OutputType, name)
			if !ok {
				return nil, fmt.Errorf("unknown output plugin: %s", name)
			}

			// Parse specific configuration
			if err := p.md.PrimitiveDecode(primitive, pluginConfig); err != nil {
				return nil, err
			}

			// Parse common configuration
			commonConfig := &telegraf.CommonOutputConfig{}
			if err := p.md.PrimitiveDecode(primitive, commonConfig); err != nil {
				return nil, err
			}

			plugin := &telegraf.OutputConfig{
				Config:       commonConfig,
				PluginConfig: pluginConfig,
			}
			configs = append(configs, plugin)
		}
		outputConfigs[name] = configs
	}
	return outputConfigs, nil
}

func (p *parser) loadLoaders(loaders map[string][]toml.Primitive) (map[string][]*telegraf.LoaderConfig, error) {
	loaderConfigs := make(map[string][]*telegraf.LoaderConfig, 0)

	for name, primitives := range loaders {
		configs := make([]*telegraf.LoaderConfig, 0)
		for _, primitive := range primitives {
			pluginConfig, ok := p.registry.GetPluginConfig(telegraf.LoaderType, name)
			if !ok {
				return nil, fmt.Errorf("unknown loader plugin: %s", name)
			}

			// Parse Loader specific configuration
			if err := p.md.PrimitiveDecode(primitive, pluginConfig); err != nil {
				return nil, err
			}

			// Parse common Loader configuration
			commonConfig := &telegraf.CommonLoaderConfig{}
			if err := p.md.PrimitiveDecode(primitive, commonConfig); err != nil {
				return nil, err
			}

			plugin := &telegraf.LoaderConfig{
				Config:       commonConfig,
				PluginConfig: pluginConfig,
			}
			configs = append(configs, plugin)
		}
		loaderConfigs[name] = configs
	}
	return loaderConfigs, nil
}
