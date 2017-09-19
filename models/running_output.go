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
	*telegraf.OutputPlugin
}

func NewRunningOutput(
	plugin *telegraf.OutputPlugin,
) *RunningOutput {
	return &RunningOutput{plugin}
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
