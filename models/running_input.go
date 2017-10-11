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
	Config *telegraf.CommonInputConfig
	Input  telegraf.Input
}

func NewRunningInput(
	config *telegraf.InputConfig,
	factory telegraf.PluginFactory,
) (*RunningInput, error) {
	input, err := CreateInput(config.PluginConfig, factory)
	if err != nil {
		return nil, err
	}

	// switch input := input.(type) {
	// case telegraf.ParserInput:
	// 	// create parser

	// 	parser, err := CreateParser(config.ParserConfig, factory)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	input.SetParser(parser)
	// }

	return &RunningInput{config.Config, input}, nil
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
