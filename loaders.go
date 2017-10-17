package telegraf

import (
	"context"
)

// Loader is the interface for a plugin that loads a Config.
type Loader interface {
	// Watch establishes watching for updates.
	//
	// Does not return until the watch is established or an error occurs.
	Watch(context.Context) (Waiter, error)

	// Load loads the Config.
	Load(context.Context, ConfigRegistry) (*Config, error)
}

// Should this be WatchWaiter?
//
// Waiter allows you to wait for a watch to complete.
type Waiter interface {
	// Wait blocks until the watch has completed.
	Wait() error
}
