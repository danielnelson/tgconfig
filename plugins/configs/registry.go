package configs

import (
	"github.com/influxdata/tgconfig/plugins/configs/null"
	"github.com/influxdata/tgconfig/plugins/configs/toml"
)

var Configs = map[string]interface{}{
	null.Name: null.New,
	toml.Name: toml.New,
}
