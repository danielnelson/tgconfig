package parsers

import (
	"github.com/influxdata/tgconfig/plugins/parsers/collectd"
	"github.com/influxdata/tgconfig/plugins/parsers/influx"
)

var Parsers = map[string]interface{}{
	influx.Name:   influx.New,
	collectd.Name: collectd.New,
}
