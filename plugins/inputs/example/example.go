package example

import (
	"fmt"
	"strings"

	telegraf "github.com/influxdata/tgconfig"
)

const (
	Name = "example"
)

// Config contains configuration for ExampleInput.  It's structure
// must match the data in the configuration file or source.
type Config struct {
	telegraf.ParserConfig

	Value string `toml:"value"`
}

// Example is an example input plugin.
type Example struct {
	Config Config
}

// New creates an Example from a Config.
func New(config *Config) (telegraf.Input, error) {
	return &Example{*config}, nil
}

func (p *Example) Gather() error {
	return nil
}

// Debugging
func (p *Example) String() string {
	return strings.Join(
		[]string{
			"Input: " + Name,
			fmt.Sprintf("  value:%s", p.Config.Value),
			p.Config.ParserConfig.String(),
		}, "\n")
}
