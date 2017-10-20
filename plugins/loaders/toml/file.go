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

func New(config *Config) ([]telegraf.Loader, error) {
	return []telegraf.Loader{&Toml{Config: *config}}, nil
}

func (c *Toml) Load(ctx context.Context, registry telegraf.ConfigRegistry) (*telegraf.Config, error) {
	reader, err := os.Open(c.Config.Path)
	if err != nil {
		return nil, err
	}

	parser := NewParser(registry)
	return parser.Parse(reader)
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
