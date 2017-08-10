package models

import (
	"fmt"
	"strings"

	"github.com/influxdata/tgconfig"
)

// Existing: models/running_input.RunningInput
type RunningInput struct {
	*telegraf.InputPlugin
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
