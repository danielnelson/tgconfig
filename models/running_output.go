package models

import (
	"fmt"
	"strings"

	telegraf "github.com/influxdata/tgconfig"
)

type RunningOutput struct {
	*telegraf.OutputPlugin
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
