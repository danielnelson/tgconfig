package models

import (
	"context"

	telegraf "github.com/influxdata/tgconfig"
)

// RunningLoader exists for symmetry with the other Running classes.
type RunningLoader struct {
	Config *telegraf.CommonLoaderConfig
	Loader telegraf.Loader
}

func NewRunningLoader(
	name string,
	config *telegraf.LoaderConfig,
	registry telegraf.FactoryRegistry,
) (*RunningLoader, error) {
	loader, err := registry.CreateLoader(
		telegraf.LoaderType, name, config.PluginConfig)
	if err != nil {
		return nil, err
	}

	return &RunningLoader{config.Config, loader}, nil
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
