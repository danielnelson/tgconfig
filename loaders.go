package telegraf

import (
	"context"
	"errors"
)

var ReloadConfig = errors.New("reload config")

// Loader is a config plugin, corresponds to the Input or Output struct
type Loader interface {
	Name() string

	// Watch begins watching for updates, once this function returns the watch
	// is established.
	Watch(context.Context) (Waiter, error)

	// Should we remove registry?  Would need to move it to the New function.
	Load(context.Context, ConfigRegistry) (*Config, error)
}

// Should this be WatchWaiter?
type Waiter interface {
	// Wait blocks until the watch has completed
	Wait() error
}
