package models

import (
	"fmt"

	telegraf "github.com/influxdata/tgconfig"
)

// RunningInput ensures measurement filtering is applied correctly to all
// Inputs and ensures Inputs are used correctly.
//
// Existing: internal/models/running_input.RunningInput
type RunningInput struct {
	Config *telegraf.CommonInputConfig
	Input  telegraf.Input
}

func NewRunningInput(
	name string,
	config *telegraf.InputConfig,
	registry telegraf.FactoryRegistry,
) (*RunningInput, error) {
	factory, ok := registry.GetFactory(telegraf.InputType, name)
	if !ok {
		return nil, fmt.Errorf("unknown plugin: %s", name)
	}
	input, err := CreateInput(config.PluginConfig, factory)
	if err != nil {
		return nil, err
	}

	switch input := input.(type) {
	case telegraf.ParserInput:
		parserName := "influx"
		if config.Config.DataFormat != "" {
			parserName = config.Config.DataFormat
		}
		// how do we know what parser to load?  based on data_format?
		factory, ok := registry.GetFactory(telegraf.ParserType, parserName)
		if !ok {
			return nil, fmt.Errorf("unknown plugin: %s", name)
		}
		parser, err := CreateParser(config.ParserConfig, factory)
		if err != nil {
			return nil, err
		}
		input.SetParser(parser)
	}

	return &RunningInput{config.Config, input}, nil
}
