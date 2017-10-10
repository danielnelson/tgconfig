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
type InputConfig struct {
	FilterConfig
	ParserConfig
}

func (c *InputConfig) String() string {
	return fmt.Sprintf("  input:name_override: %s", c.NameOverride)
}

// OutputConfig is the required config for all outputs.
//
// This configurations is generally used by the RunningOutput.
type OutputConfig struct {
	FilterConfig
}

func (c *OutputConfig) String() string {
	return fmt.Sprintf("  output:name_override: %s", c.NameOverride)
}

// LoaderConfig is the shared config for all loaders.
//
// Just here for symmetry for now
type LoaderConfig struct {
}

// InputPlugin packages the global settings with the Input instance.
type InputPlugin struct {
	Config interface{}
	*InputConfig
}

// OutputPlugin packages the global settings with the Output instance.
type OutputPlugin struct {
	Config interface{}
	*OutputConfig
}

// LoaderPlugin packages the global Loader config with the Loader config.
//
// LoaderPlugin exists for symmetry with InputPlugin/OutputPlugin.  If a
// LoaderConfig was introduced it would be stored here.
type LoaderPlugin struct {
	Config interface{}
	*LoaderConfig
}

// Config is the top level configuration struct.
//
// Loader plugins build this struct.
type Config struct {
	Agent AgentConfig
	// input-name -> ConfigType
	Inputs  map[string][]*InputPlugin
	Outputs map[string][]*OutputPlugin
	Loaders map[string][]*LoaderPlugin
}

// PluginRegistry holds the set of available plugins.  This provides a layer of
// indirection so that you can define a custom set of plugins.
type PluginRegistry struct {
	Loaders map[string]interface{}
	Inputs  map[string]interface{}
	Outputs map[string]interface{}
}

type ConfigRegistry struct {
	Loaders map[string]interface{}
	Inputs  map[string]interface{}
	Outputs map[string]interface{}
}
