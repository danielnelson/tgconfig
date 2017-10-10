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
	Config *telegraf.OutputConfig
	Output telegraf.Output
}

func NewRunningOutput(config *telegraf.OutputConfig, output telegraf.Output) *RunningOutput {
	return &RunningOutput{config, output}
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
