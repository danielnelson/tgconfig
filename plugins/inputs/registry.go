package inputs

import (
	"github.com/influxdata/tgconfig/plugins/inputs/example"
	"github.com/influxdata/tgconfig/plugins/inputs/example2"
)

var Inputs = map[string]interface{}{
	example.Name:  example.New,
	example2.Name: example2.New,
}
