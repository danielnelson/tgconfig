package influx

import (
	telegraf "github.com/influxdata/tgconfig"
)

const (
	Name = "influx"
)

type Config struct{}

type Influx struct{}

func (i *Influx) Parse(buf []byte) ([]telegraf.Metric, error) {
	return []telegraf.Metric{}, nil
}

func New(config *Config) (telegraf.Parser, error) {
	return &Influx{}, nil
}
