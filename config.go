package telegraf

import (
	"fmt"
)

// FilterConfig contains the standard filtering configuration.  We may need
// one of these for each of inputs, processors, aggregators, outputs.
type FilterConfig struct {
	NameOverride string `toml:"name_override"`
}

// AgentConfig contains the Agent configuration
type AgentConfig struct {
	Interval int `toml:"interval"`
}

// InputConfig is the required config for all inputs.
//
// This configuration is generally used by the Running output.  This has
// upsides and downsides:
// - pro:
//   - This configuration is hidden from the plugin; reducing clutter and
//     temptation to misuse.
//   - Plugin author does not need to remember to add standard options.
// - con:
//   - Breaks the product that a Config plugin must produce into two parts.
//   - Could be hard to split item with some parsers?
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

//
type InputPlugin struct {
	Input
	Config *InputConfig
}

type OutputPlugin struct {
	Output
	Config *OutputConfig
}

// here now for symmetry...
type ConfigLoaderPlugin struct {
	ConfigLoader
}

// Config is the top level configuration struct.
type Config struct {
	Agent   AgentConfig
	Inputs  []*InputPlugin
	Outputs []*OutputPlugin
	Configs []*ConfigLoaderPlugin
}

type Plugins struct {
	Configs map[string]interface{}
	Inputs  map[string]interface{}
	Outputs map[string]interface{}
}
