package models

import (
	"fmt"

	telegraf "github.com/influxdata/tgconfig"
)

// RunningLoader exists for symmetry with the other Running classes.
type RunningLoader struct {
	*telegraf.LoaderPlugin
}

func NewRunningLoader(
	plugin *telegraf.LoaderPlugin,
) *RunningLoader {
	return &RunningLoader{plugin}
}

func (rc *RunningLoader) String() string {
	switch s := rc.Loader.(type) {
	case fmt.Stringer:
		return s.String()
	default:
		return ""
	}
}
