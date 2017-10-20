package collectd

import (
	telegraf "github.com/influxdata/tgconfig"
)

const (
	Name = "collectd"
)

type Config struct {
	AuthFile string `toml:"collectd_auth_file"`
}

type Collectd struct {
	AuthFile string
}

func (i *Collectd) Parse(buf []byte) ([]telegraf.Metric, error) {
	return []telegraf.Metric{}, nil
}

func New(config *Config) (telegraf.Parser, error) {
	return &Collectd{}, nil
}
