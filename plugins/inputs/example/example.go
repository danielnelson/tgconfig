package example

import (
	telegraf "github.com/influxdata/tgconfig"
	"github.com/influxdata/tgconfig/plugins/inputs"
)

// Config contains configuration for ExampleInput.  It's structure
// must match the data in the configuration file or source.
type Config struct {
	Value string `toml:"value"`
}

// Example is an example input plugin.
type Example struct {
	Config Config

	parser telegraf.Parser
}

// New creates an Example from a Config.
func New(config *Config) ([]telegraf.Input, error) {
	return []telegraf.Input{&Example{Config: *config}}, nil
}

func (p *Example) Gather() error {
	return nil
}

func (p *Example) SetParser(parser telegraf.Parser) {
	p.parser = parser
}

func init() {
	inputs.Add("example", New)
}
