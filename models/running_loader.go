package models

import (
	"fmt"

	telegraf "github.com/influxdata/tgconfig"
)

type RunningLoader struct {
	*telegraf.LoaderPlugin
}

func (rc *RunningLoader) String() string {
	switch s := rc.Loader.(type) {
	case fmt.Stringer:
		return s.String()
	default:
		return ""
	}
}
