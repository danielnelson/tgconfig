package models

import (
	"fmt"
	"reflect"

	"github.com/influxdata/tgconfig"
)

func CreateInput(
	config telegraf.PluginConfig,
	factory telegraf.PluginFactory,
) (telegraf.Input, error) {
	plugin, err := createPlugin(config, factory)

	switch plugin := plugin.(type) {
	case telegraf.Input:
		return plugin, nil
	}

	switch err := err.(type) {
	case error:
		return nil, err
	}

	panic("Error loading plugin")
}

func CreateOutput(
	config telegraf.PluginConfig,
	factory telegraf.PluginFactory,
) (telegraf.Output, error) {
	plugin, err := createPlugin(config, factory)

	switch plugin := plugin.(type) {
	case telegraf.Output:
		return plugin, nil
	}

	switch err := err.(type) {
	case error:
		return nil, err
	}

	panic("Error loading plugin")
}

func CreateLoader(
	config telegraf.PluginConfig,
	factory telegraf.PluginFactory,
) (telegraf.Loader, error) {
	plugin, err := createPlugin(config, factory)

	switch plugin := plugin.(type) {
	case telegraf.Loader:
		return plugin, nil
	}

	switch err := err.(type) {
	case error:
		return nil, err
	}

	panic("Error loading plugin")
}

func CreateParser(
	config telegraf.PluginConfig,
	factory telegraf.PluginFactory,
) (telegraf.Parser, error) {
	plugin, err := createPlugin(config, factory)

	switch plugin := plugin.(type) {
	case telegraf.Parser:
		return plugin, nil
	}

	switch err := err.(type) {
	case error:
		return nil, err
	}

	panic("Error loading plugin")
}

func createPlugin(
	config telegraf.PluginConfig,
	factory telegraf.PluginFactory,
) (interface{}, interface{}) {
	vfactory := reflect.ValueOf(factory)

	// Call factory with the config struct
	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(config)
	result := vfactory.Call(in)
	if len(result) != 2 {
		panic(fmt.Sprintf("plugin factory does not return correct values: %T", factory))
	}
	plugin := result[0].Interface()
	err := result[1].Interface()
	return plugin, err
}
