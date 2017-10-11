package example

import (
	"fmt"
	"strings"

	telegraf "github.com/influxdata/tgconfig"
)

const (
	Name = "example"
)

// ExampleOutputConfig contains the configuration for ExampleOutput.
type Config struct {
	Value string `toml:"value"`
}

// ExampleOutput is an example output plugin.
type Example struct {
	Config Config
}

// Connect connects the output.
func (p *Example) Connect() error {
	return nil
}

func (p *Example) String() string {
	return strings.Join([]string{
		"Output: " + Name,
		fmt.Sprintf("  value:%s", p.Config.Value),
	}, "\n")
}

// NewExampleOutput creates an ExampleOutput from an ExampleOutputConfig.
func New(config *Config) (telegraf.Output, error) {
	return &Example{*config}, nil
}
