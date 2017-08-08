package telegraf

import "context"

// ConfigLoader is a config plugin
type ConfigLoader interface {
	Load(*Plugins) (*Config, error)

	// todo: new return type, enum or interface or error
	// how to wait on more than one of these? n-goroutines?
	// start/stop?
	Monitor(context.Context) error
}
