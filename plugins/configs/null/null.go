package null

import (
	"context"

	telegraf "github.com/influxdata/tgconfig"
)

const (
	Name = "null"
)

type Config struct {
}

type Null struct {
}

func New(config *Config) (telegraf.ConfigLoader, error) {
	return &Null{}, nil
}

func (c *Null) Load(plugins *telegraf.Plugins) (*telegraf.Config, error) {
	return &telegraf.Config{}, nil
}

func (c *Null) Monitor(ctx context.Context) error {
	return nil
}

// Debugging
func (c *Null) String() string {
	return "Config: null"
}
