package outputs

import (
	"github.com/influxdata/tgconfig/plugins/outputs/example"
)

var Outputs = map[string]interface{}{
	example.Name: example.New,
}
