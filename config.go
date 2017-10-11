package telegraf

import (
	"fmt"
)

// FilterConfig contains the standard filtering configuration.  We may need
// one of these for each of inputs, processors, aggregators, outputs.
type FilterConfig struct {
	NameOverride string `toml:"name_override"`
}

// ParserConfig is a new method for defining parsers.  Still needs some work.
type ParserConfig struct {
	Parser string `toml:"parser"`
}

// AgentConfig contains the Agent configuration
type AgentConfig struct {
	Interval int `toml:"interval"`
}

// InputConfig is the required config for all inputs.
//
// This configuration is generally used by the Running input.  This has
// upsides and downsides:
// - pro:
//   - This configuration is hidden from the plugin; reducing clutter and
//     temptation to misuse.
//   - Plugin author does not need to remember to add standard options.
// - con:
//   - Breaks the product that a Config plugin must produce into two parts.
type CommonInputConfig struct {
	FilterConfig
	ParserConfig

	Config interface{}
}

func (c *CommonInputConfig) String() string {
	return fmt.Sprintf("  input:name_override: %s", c.NameOverride)
}

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
type CommonLoaderConfig struct{}

func (c *CommonLoaderConfig) String() string {
	return ""
}

type PluginConfig = interface{}
type PluginFactory = interface{}

type InputConfig struct {
	Config       *CommonInputConfig
	PluginConfig PluginConfig
}

func (c *InputConfig) String() string {
	return ""
}

// OutputPlugin packages the global settings with the Output instance.
type OutputConfig struct {
	Config       *CommonOutputConfig
	PluginConfig PluginConfig
}

func (c *OutputConfig) String() string {
	return ""
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

// type ConfigRegistry interface {
// 	Input(name string) interface{}

// }

// ConfigRegistry holds the set of available plugins.  This provides a layer of
// indirection so that you can define a custom set of plugins.
//
// plugin_name -> plugin_factory
// i.e.: "cpu" -> cpu.New(*cpu.Config) (Input, error)
type ConfigRegistry struct {
	Loaders map[string]PluginFactory
	Inputs  map[string]PluginFactory
	Outputs map[string]PluginFactory
}


