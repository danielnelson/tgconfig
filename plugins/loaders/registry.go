package loaders

import (
	"github.com/influxdata/tgconfig/plugins/loaders/http"
	"github.com/influxdata/tgconfig/plugins/loaders/null"
	"github.com/influxdata/tgconfig/plugins/loaders/toml"
)

var Loaders = map[string]interface{}{
	http.Name: http.New,
	null.Name: null.New,
	toml.Name: toml.New,
}
