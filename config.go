package telegraf

import (
	"fmt"
)

type PluginType int

const (
	InputType PluginType = iota
	OutputType
	LoaderType
	ParserType
)

// AgentConfig contains the Agent configuration
type AgentConfig struct {
	Interval int `toml:"interval"`
}

// FilterConfig contains the standard filtering configuration.  We may need
// one of these for each of inputs, processors, aggregators, outputs.
type FilterConfig struct {
	NameOverride string `toml:"name_override"`
}

// ParserConfig is the shared configuration for Parsers.
type ParserConfig struct {
	DataFormat string `toml:"data_format"`
}

// CommonInputConfig is the common config for all Inputs.
type CommonInputConfig struct {
	FilterConfig
	ParserConfig
}

// func (c *CommonInputConfig) String() string {
// 	var b bytes.Buffer
// 	enc := toml.NewEncoder(&b)
// 	enc.Encode(c)
// 	return b.String()
// 	// return fmt.Sprintf("  input:name_override: %s", c.NameOverride)
// }

// OutputConfig is the required config for all outputs.
//
// This configurations is generally used by the RunningOutput.
type CommonOutputConfig struct {
	FilterConfig
}

func (c *CommonOutputConfig) String() string {
	return fmt.Sprintf("  output:name_override: %s", c.NameOverride)
}

// LoaderConfig is the config for any loaders.
//
// Just here for symmetry for now
type CommonLoaderConfig struct {
}

// Config struct for plugin
type PluginConfig = interface{}

// Factory function for plugin: func (*PluginConfig) (Plugin, error)
type PluginFactory = interface{}

type InputConfig struct {
	Config       *CommonInputConfig
	PluginConfig PluginConfig
	ParserConfig PluginConfig
}

// OutputPlugin packages the global settings with the Output instance.
type OutputConfig struct {
	Config       *CommonOutputConfig
	PluginConfig PluginConfig
}

// LoaderPlugin packages the global Loader config with the Loader config.
//
// LoaderPlugin exists for symmetry with InputPlugin/OutputPlugin.  If a
// LoaderConfig was introduced it would be stored here.
type LoaderConfig struct {
	Config       *CommonLoaderConfig
	PluginConfig PluginConfig
}

// Config is the top level configuration struct.
//
// Loader plugins build this struct.
type Config struct {
	Agent   AgentConfig
	Inputs  map[string][]*InputConfig
	Outputs map[string][]*OutputConfig
	Loaders map[string][]*LoaderConfig
}

type FactoryRegistry interface {
	GetFactory(pluginType PluginType, name string) (PluginFactory, bool)
	GetConfigRegistry() ConfigRegistry
}

type ConfigRegistry interface {
	GetPluginConfig(pluginType PluginType, name string) (PluginConfig, bool)
}
