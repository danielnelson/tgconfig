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

func (c *Null) Load(ctx context.Context, plugins *telegraf.Plugins) (*telegraf.Config, error) {
	return &telegraf.Config{}, nil
}

func (c *Null) Name() string {
	return Name
}

// Monitor is the minimum implementation of ConfigPlugin.Monitor
func (c *Null) Monitor(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	}
}

// MonitorC is the minimum implementation of ConfigPlugin.MonitorC
func (c *Null) MonitorC(ctx context.Context) <-chan error {
	out := make(chan error)

	go func() {
		select {
		case <-ctx.Done():
			out <- ctx.Err()
			break
		}
		close(out)
	}()

	return out
}

// Debugging
func (c *Null) String() string {
	return "Config: null"
}

func New(config *Config) (telegraf.Loader, error) {
	return &Null{}, nil
}
