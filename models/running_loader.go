package models

import (
	"context"
	"fmt"

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
	factory, ok := registry.GetFactory(telegraf.LoaderType, name)
	if !ok {
		return nil, fmt.Errorf("unknown plugin: %s", name)
	}
	loader, err := CreateLoader(config.PluginConfig, factory)
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
