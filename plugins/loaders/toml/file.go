package toml

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	telegraf "github.com/influxdata/tgconfig"
)

const (
	Name = "toml"
)

type Toml struct {
	Config Config
}

type Config struct {
	// Path is the main config file
	Path string
	// Directory is an directory containing config snippets
	Directory string
}

func New(config *Config) (telegraf.Loader, error) {
	return &Toml{Config: *config}, nil
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

func (c *Toml) Watch(ctx context.Context) (telegraf.Waiter, error) {
	return NewSignalWaiter(ctx)
}

type SignalWaiter struct {
	ctx context.Context
	wg  sync.WaitGroup
}

func NewSignalWaiter(ctx context.Context) (*SignalWaiter, error) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGHUP)

	w := &SignalWaiter{ctx: ctx}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		select {
		case <-sigC:
			break
		case <-ctx.Done():
			break
		}
		signal.Stop(sigC)
	}()
	return w, nil
}

func (w *SignalWaiter) Wait() error {
	w.wg.Wait()
	return w.ctx.Err()
}

// Debugging
func (c *Toml) String() string {
	return "Config: toml"
}
