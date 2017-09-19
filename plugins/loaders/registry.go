package loaders

import (
	"github.com/influxdata/tgconfig/plugins/loaders/null"
	"github.com/influxdata/tgconfig/plugins/loaders/toml"
)

var Loaders = map[string]interface{}{
	null.Name:     null.New,
	toml.Name:     toml.New,
	toml.HTTPName: toml.NewHTTP,
}
