package null

import (
	"context"
	"sync"

	telegraf "github.com/influxdata/tgconfig"
)

const (
	Name = "null"
)

type Config struct {
}

type Null struct {
}

func (l *Null) Load(ctx context.Context, registry *telegraf.ConfigRegistry) (*telegraf.Config, error) {
	return &telegraf.Config{}, nil
}

func (l *Null) Name() string {
	return Name
}

func (l *Null) Watch(ctx context.Context) (telegraf.Waiter, error) {
	return NewNullWaiter(ctx)
}

type NullWaiter struct {
	ctx context.Context
	wg  sync.WaitGroup
}

func NewNullWaiter(ctx context.Context) (*NullWaiter, error) {
	w := &NullWaiter{ctx: ctx}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		select {
		case <-w.ctx.Done():
			break
		}
	}()
	return w, nil
}

func (w *NullWaiter) Wait() error {
	w.wg.Wait()
	return w.ctx.Err()
}

// Debugging
func (l *Null) String() string {
	return "Config: null"
}

func New(config *Config) (telegraf.Loader, error) {
	return &Null{}, nil
}
