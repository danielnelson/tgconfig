package models

import (
	"fmt"
	"strings"

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
	factory, ok := registry.GetFactory(telegraf.OutputType, name)
	if !ok {
		return nil, fmt.Errorf("unknown plugin: %s", name)
	}
	output, err := CreateOutput(config.PluginConfig, factory)
	if err != nil {
		return nil, err
	}

	return &RunningOutput{config.Config, output}, nil
}

func (ro *RunningOutput) String() string {
	lines := []string{}

	switch s := ro.Output.(type) {
	case fmt.Stringer:
		lines = append(lines, s.String())
	}

	lines = append(lines, ro.Config.String())
	return strings.Join(lines, "\n")
}
