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

func (l *Null) Load(ctx context.Context, registry *telegraf.PluginRegistry) (*telegraf.Config, error) {
	return &telegraf.Config{}, nil
}

func (l *Null) Name() string {
	return Name
}

// Monitor is the minimum implementation of ConfigPlugin.Monitor
func (l *Null) Monitor(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	}
}

// MonitorC is the minimum implementation of ConfigPlugin.MonitorC
func (l *Null) MonitorC(ctx context.Context) (<-chan error, error) {
	out := make(chan error)

	go func() {
		select {
		case <-ctx.Done():
			out <- ctx.Err()
			break
		}
		close(out)
	}()

	return out, nil
}

// StartWatch establishes the watch
func (l *Null) StartWatch(ctx context.Context) error {
	return nil
}

// WaitWatch blocks until the Loader should be reloaded
func (l *Null) WaitWatch(ctx context.Context) error {
	select {
	case <-ctx.Done():
		break
	}
	return ctx.Err()
}

// Debugging
func (l *Null) String() string {
	return "Config: null"
}

func New(config *Config) (telegraf.Loader, error) {
	return &Null{}, nil
}
