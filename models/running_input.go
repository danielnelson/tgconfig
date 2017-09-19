package models

import (
	"fmt"
	"strings"

	"github.com/influxdata/tgconfig"
)

// RunningInput ensures measurement filtering is applied correctly to all
// Inputs and ensures Inputs are used correctly.
//
// Existing: internal/models/running_input.RunningInput
type RunningInput struct {
	*telegraf.InputPlugin
}

func NewRunningInput(
	plugin *telegraf.InputPlugin,
) *RunningInput {
	return &RunningInput{plugin}
}

func (ri *RunningInput) String() string {
	lines := []string{}

	switch s := ri.Input.(type) {
	case fmt.Stringer:
		lines = append(lines, s.String())
	}

	lines = append(lines, ri.Config.String())
	return strings.Join(lines, "\n")
}
