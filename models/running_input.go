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
	Config *telegraf.InputConfig
	Input  telegraf.Input
}

func NewRunningInput(config *telegraf.InputConfig, input telegraf.Input) *RunningInput {
	return &RunningInput{config, input}
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
