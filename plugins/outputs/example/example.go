package example

import (
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

// NewExampleOutput creates an ExampleOutput from an ExampleOutputConfig.
func New(config *Config) ([]telegraf.Output, error) {
	return []telegraf.Output{&Example{*config}}, nil
}
