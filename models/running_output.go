package models

import (
	telegraf "github.com/influxdata/tgconfig"
)

// RunningOutput ensures measurement filtering is applied correctly to all
// Output, handles buffering, and ensures the Output is used correctly with
// respect to concurrency.
//
// Existing: models/running_output.RunningOutput
type RunningOutput struct {
	Config *telegraf.CommonOutputConfig
	Output telegraf.Output
}

func NewRunningOutput(
	name string,
	config *telegraf.OutputConfig,
	registry telegraf.FactoryRegistry,
) (*RunningOutput, error) {
	output, err := registry.CreateOutput(
		telegraf.OutputType, name, config.PluginConfig)
	if err != nil {
		return nil, err
	}

	return &RunningOutput{config.Config, output}, nil
}
