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
	config *telegraf.LoaderConfig,
	factory telegraf.PluginFactory,
) (*RunningLoader, error) {
	loader, err := CreateLoader(config.PluginConfig, factory)
	if err != nil {
		return nil, err
	}

	return &RunningLoader{config.Config, loader}, nil
}

func (rc *RunningLoader) String() string {
	switch s := rc.Loader.(type) {
	case fmt.Stringer:
		return s.String()
	default:
		return ""
	}
}

func (rc *RunningLoader) Name() string {
	return rc.Loader.Name()
}

func (rc *RunningLoader) Load(
	ctx context.Context,
	registry *telegraf.ConfigRegistry,
) (*telegraf.Config, error) {
	return rc.Loader.Load(ctx, registry)
}

func (rc *RunningLoader) Watch(ctx context.Context) (telegraf.Waiter, error) {
	return rc.Loader.Watch(ctx)
}
