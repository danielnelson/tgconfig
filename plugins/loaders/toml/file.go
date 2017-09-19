package toml

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	telegraf "github.com/influxdata/tgconfig"
)

const (
	Name = "toml"
)

type SignalWatcher struct {
	sigC chan os.Signal
}

type Toml struct {
	Config Config
	SignalWatcher
}

type Config struct {
	// Path is the main config file
	Path string
	// Directory is an directory containing config snippets
	Directory string
}

func New(config *Config) (telegraf.Loader, error) {
	return &Toml{Config: *config,
		SignalWatcher: SignalWatcher{}}, nil
}

func (c *Toml) Name() string {
	return Name
}

func (c *Toml) Load(ctx context.Context, registry *telegraf.PluginRegistry) (*telegraf.Config, error) {
	reader, err := os.Open(c.Config.Path)
	if err != nil {
		return nil, err
	}

	parser := NewParser(registry)
	return parser.Parse(reader)
}

func (c *Toml) Monitor(ctx context.Context) error {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP)
	defer signal.Stop(signals)

	select {
	case <-signals:
		return telegraf.ReloadConfig
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Toml) MonitorC(ctx context.Context) (<-chan error, error) {
	out := make(chan error)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP)

	go func() {
		select {
		case sig := <-signals:
			if sig == syscall.SIGHUP {
				out <- telegraf.ReloadConfig
				break
			}
		case <-ctx.Done():
			out <- ctx.Err()
			break
		}

		signal.Stop(signals)
		close(out)
	}()

	return out, nil
}

// StartWatch establishes the watch
func (w *SignalWatcher) StartWatch(ctx context.Context) error {
	w.sigC = make(chan os.Signal, 1)
	signal.Notify(w.sigC, syscall.SIGHUP)
	return nil
}

// WaitWatch blocks until the Loader should be reloaded
func (w *SignalWatcher) WaitWatch(ctx context.Context) error {
	select {
	case signal := <-w.sigC:
		if signal == syscall.SIGHUP {
			break
		}
	case <-ctx.Done():
		break
	}
	signal.Stop(w.sigC)
	return ctx.Err()
}

// Debugging
func (c *Toml) String() string {
	return "Config: toml"
}
