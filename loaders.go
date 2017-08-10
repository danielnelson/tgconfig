package telegraf

import (
	"context"
	"errors"
)

var ReloadConfig = errors.New("reload config")

// Loader is a config plugin, corresponds to the Input or Output struct
type Loader interface {
	Name() string

	Load(context.Context, *Plugins) (*Config, error)

	Monitor(context.Context) error

	// Should there be a new return type, enum or interface or error
	// results: cancelled, timeout, reload, error(ie: connection failed)
	//
	// Do not return until listening?  What if can't establish connection?
	//
	// Is it okay to miss signals?
	MonitorC(context.Context) <-chan error
}
