package telegraf

import (
	"context"
	"errors"
)

var ReloadConfig = errors.New("reload config")

// Loader is a config plugin, corresponds to the Input or Output struct
type Loader interface {
	Name() string

	// Should we remove registry?  Would need to move it to the New function.
	Load(context.Context, *PluginRegistry) (*Config, error)

	Monitor(context.Context) error

	// Should there be a new return type, enum or interface or error
	// results: cancelled, timeout, reload, error(ie: connection failed)
	//
	// Do not return until listening?  What if can't establish connection?
	//
	// Is it okay to miss signals?
	//
	// Delete, just an experiement but I prefer the plain Monitor functin.

	MonitorC(context.Context) (<-chan error, error)

	// Could make this a different interface that can optionally be
	// implemented.

	Watch(context.Context) (Waiter, error)
}

type Waiter interface {
	// Wait blocks until the watch has completed
	Wait() error
}
