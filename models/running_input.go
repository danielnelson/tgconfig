package models

import (
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
	input, err := registry.CreateInput(
		telegraf.InputType, name, config.PluginConfig)
	if err != nil {
		return nil, err
	}

	switch input := input.(type) {
	case telegraf.ParserInput:
		parserName := "influx"
		if config.Config.DataFormat != "" {
			parserName = config.Config.DataFormat
		}
		parser, err := registry.CreateParser(
			telegraf.ParserType, parserName, config.ParserConfig)
		if err != nil {
			return nil, err
		}
		input.SetParser(parser)
	}

	return &RunningInput{config.Config, input}, nil
}
