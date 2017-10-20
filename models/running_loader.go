package models

import (
	"context"

	telegraf "github.com/influxdata/tgconfig"
)

// RunningLoader exists for symmetry with the other Running classes.
type RunningLoader struct {
	Config *telegraf.CommonLoaderConfig
	Loader telegraf.Loader
	Name   string
}

func NewRunningLoaders(
	name string,
	config *telegraf.LoaderConfig,
	registry telegraf.Registry,
) ([]*RunningLoader, error) {
	loaders, err := registry.CreateLoaders(name, config.PluginConfig)
	if err != nil {
		return nil, err
	}

	r := make([]*RunningLoader, len(loaders))
	for i, loader := range loaders {
		r[i] = &RunningLoader{
			Config: config.Config,
			Loader: loader,
			Name:   name,
		}
	}
	return r, nil
}

func (rc *RunningLoader) Watch(ctx context.Context) (telegraf.Waiter, error) {
	return rc.Loader.Watch(ctx)
}

func (rc *RunningLoader) Load(
	ctx context.Context,
	registry telegraf.ConfigRegistry,
) (*telegraf.Config, error) {
	return rc.Loader.Load(ctx, registry)
}
