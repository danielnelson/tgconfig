package models

import (
	"fmt"
	"reflect"

	"github.com/influxdata/tgconfig"
)

type factories struct {
	Loaders map[string]telegraf.PluginFactory
	Inputs  map[string]telegraf.PluginFactory
	Outputs map[string]telegraf.PluginFactory
	Parsers map[string]telegraf.PluginFactory
}

func NewFactories(
	loaders map[string]telegraf.PluginFactory,
	inputs map[string]telegraf.PluginFactory,
	outputs map[string]telegraf.PluginFactory,
	parsers map[string]telegraf.PluginFactory,
) (*factories, error) {
	err := check(loaders)
	if err != nil {
		return nil, err
	}
	err = check(inputs)
	if err != nil {
		return nil, err
	}
	err = check(outputs)
	if err != nil {
		return nil, err
	}
	err = check(parsers)
	if err != nil {
		return nil, err
	}

	return &factories{
		Loaders: loaders,
		Inputs:  inputs,
		Outputs: outputs,
		Parsers: parsers,
	}, nil
}

func (c *factories) GetFactory(
	pluginType telegraf.PluginType,
	name string,
) (telegraf.PluginFactory, bool) {
	var factory telegraf.PluginFactory
	var ok bool

	switch pluginType {
	case telegraf.LoaderType:
		factory, ok = c.Loaders[name]
	case telegraf.InputType:
		factory, ok = c.Inputs[name]
	case telegraf.OutputType:
		factory, ok = c.Outputs[name]
	case telegraf.ParserType:
		factory, ok = c.Parsers[name]
	}

	return factory, ok
}

func (c *factories) GetConfigRegistry() telegraf.ConfigRegistry {
	return &configs{*c}
}

// Configs provides access to plugins config structure by type and name.
type configs struct {
	factories factories
}

func (c *configs) GetPluginConfig(
	pluginType telegraf.PluginType,
	name string,
) (telegraf.PluginConfig, bool) {
	factory, ok := c.factories.GetFactory(pluginType, name)
	if !ok {
		return nil, false
	}

	vfactory := reflect.ValueOf(factory)

	// Get the Type of the first and only argument
	configType := vfactory.Type().In(0)

	// Create a new config struct of this type
	return reflect.New(configType.Elem()).Interface(), true
}

// check validates the Plugin factories
func check(plugins map[string]telegraf.PluginFactory) error {
	for name, factory := range plugins {
		f := reflect.ValueOf(factory)
		if f.Kind() != reflect.Func {
			return fmt.Errorf("invalid factory %s", name)
		}

		if f.Type().NumIn() != 1 {
			return fmt.Errorf("invalid arguments %s", name)
		}

		c := f.Type().In(0)
		if c.Kind() != reflect.Ptr {
			return fmt.Errorf("invalid argument type %s", name)
		}

		// TODO: Check return values: (pluginType, error)
	}
	return nil
}

func (c *factories) CreateInput(
	pluginType telegraf.PluginType,
	name string,
	config telegraf.PluginConfig,
) (telegraf.Input, error) {
	plugin, err := c.createPlugin(pluginType, name, config)
	if err != nil {
		return nil, err
	}

	switch plugin := plugin.(type) {
	case telegraf.Input:
		return plugin, nil
	default:
		panic("input not created")
	}
}

func (c *factories) CreateParser(
	pluginType telegraf.PluginType,
	name string,
	config telegraf.PluginConfig,
) (telegraf.Parser, error) {
	plugin, err := c.createPlugin(pluginType, name, config)
	if err != nil {
		return nil, err
	}

	switch plugin := plugin.(type) {
	case telegraf.Parser:
		return plugin, nil
	default:
		panic("parser not created")
	}
}

func (c *factories) CreateOutput(
	pluginType telegraf.PluginType,
	name string,
	config telegraf.PluginConfig,
) (telegraf.Output, error) {
	plugin, err := c.createPlugin(pluginType, name, config)
	if err != nil {
		return nil, err
	}

	switch plugin := plugin.(type) {
	case telegraf.Output:
		return plugin, nil
	default:
		panic("output not created")
	}
}

func (c *factories) CreateLoader(
	pluginType telegraf.PluginType,
	name string,
	config telegraf.PluginConfig,
) (telegraf.Loader, error) {
	plugin, err := c.createPlugin(pluginType, name, config)
	if err != nil {
		return nil, err
	}

	switch plugin := plugin.(type) {
	case telegraf.Loader:
		return plugin, nil
	default:
		panic("loader not created")
	}
}

func (c *factories) createPlugin(
	pluginType telegraf.PluginType,
	name string,
	config telegraf.PluginConfig,
) (interface{}, error) {
	var factory telegraf.PluginFactory
	var ok bool

	switch pluginType {
	case telegraf.LoaderType:
		factory, ok = c.Loaders[name]
	case telegraf.InputType:
		factory, ok = c.Inputs[name]
	case telegraf.OutputType:
		factory, ok = c.Outputs[name]
	case telegraf.ParserType:
		factory, ok = c.Parsers[name]
	}
	if !ok {
		return nil, fmt.Errorf("unknown plugin %s", name)
	}

	vfactory := reflect.ValueOf(factory)

	// Call factory with the config struct
	args := make([]reflect.Value, 1)
	args[0] = reflect.ValueOf(config)
	result := vfactory.Call(args)
	if len(result) != 2 {
		panic("incorrect number of return values")
	}

	plugin := result[0].Interface()
	switch err := result[1].Interface().(type) {
	case error:
		return nil, err
	default:
		return plugin, nil
	}
}
