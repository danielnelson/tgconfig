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

func NewRunningOutputs(
	name string,
	config *telegraf.OutputConfig,
	registry telegraf.Registry,
) ([]*RunningOutput, error) {
	outputs, err := registry.CreateOutputs(name, config.PluginConfig)
	if err != nil {
		return nil, err
	}

	r := make([]*RunningOutput, len(outputs))
	for i, output := range outputs {
		r[i] = &RunningOutput{config.Config, output}
	}
	return r, nil
}
