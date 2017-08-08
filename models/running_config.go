package models

import (
	"fmt"

	"github.com/influxdata/tgconfig"
)

type RunningConfig struct {
	*telegraf.ConfigLoaderPlugin
}

func (rc *RunningConfig) String() string {
	switch s := rc.ConfigLoader.(type) {
	case fmt.Stringer:
		return s.String()
	default:
		return ""
	}
}
