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

// InputPlugin packages the global settings with the Input instance.
// internal/models/running_input.InputPlugin
type InputPlugin struct {
	Input
	Config *InputConfig
}

// InputPlugin packages the global settings with the Output instance.
// internal/models/running_output.OutputPlugin
type OutputPlugin struct {
	Output
	Config *OutputConfig
}

// LoaderPlugin exists for symmetry with InputPlugin/OutputPlugin.  If a
// shared option was introduced for LoaderPlugin it could be stored here.
type LoaderPlugin struct {
	Loader
}

// Config is the top level configuration struct.
//
// Config plugins return this.
type Config struct {
	Agent   AgentConfig
	Inputs  []*InputPlugin
	Outputs []*OutputPlugin
	Loaders []*LoaderPlugin
}

// Plugins holds a set of available plugins.  This provides a layer of
// indirection so that you can define a custom set of plugins.
type Plugins struct {
	Loaders map[string]interface{}
	Inputs  map[string]interface{}
	Outputs map[string]interface{}
}
