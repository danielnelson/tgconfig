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

func NewRunningInputs(
	name string,
	config *telegraf.InputConfig,
	registry telegraf.Registry,
) ([]*RunningInput, error) {
	inputs, err := registry.CreateInputs(name, config.PluginConfig)
	if err != nil {
		return nil, err
	}

	for _, input := range inputs {
		switch input := input.(type) {
		case telegraf.ParserInput:
			parserName := "influx"
			if config.Config.DataFormat != "" {
				parserName = config.Config.DataFormat
			}
			parser, err := registry.CreateParser(parserName, config.ParserConfig)
			if err != nil {
				return nil, err
			}
			input.SetParser(parser)
		}
	}

	r := make([]*RunningInput, len(inputs))
	for i, input := range inputs {
		r[i] = &RunningInput{config.Config, input}
	}
	return r, nil
}
