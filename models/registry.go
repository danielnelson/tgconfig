package models

import (
	"fmt"
	"reflect"

	"github.com/influxdata/tgconfig"
)

func NewRegistry(
	loaders map[string]telegraf.PluginFactory,
	inputs map[string]telegraf.PluginFactory,
	outputs map[string]telegraf.PluginFactory,
	parsers map[string]telegraf.PluginFactory,
) (*registry, error) {
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

	registry := &registry{
		loaders: loaders,
		inputs:  inputs,
		outputs: outputs,
		parsers: parsers,
	}

	return registry, nil
}

func (c *registry) GetFactory(
	pluginType telegraf.PluginType,
	name string,
) (telegraf.PluginFactory, bool) {
	var factory telegraf.PluginFactory
	var ok bool

	switch pluginType {
	case telegraf.LoaderType:
		factory, ok = c.loaders[name]
	case telegraf.InputType:
		factory, ok = c.inputs[name]
	case telegraf.OutputType:
		factory, ok = c.outputs[name]
	case telegraf.ParserType:
		factory, ok = c.parsers[name]
	}

	return factory, ok
}

func (c *registry) GetConfigRegistry() telegraf.ConfigRegistry {
	// cycle...
	return &configs{*c}
}

func (c *configs) GetPluginConfig(
	pluginType telegraf.PluginType,
	name string,
) (telegraf.PluginConfig, bool) {
	factory, ok := c.registry.GetFactory(pluginType, name)
	if !ok {
		return nil, false
	}

	vfactory := reflect.ValueOf(factory)

	// Get the Type of the first and only argument
	configType := vfactory.Type().In(0)

	// Create a new config struct of this type
	return reflect.New(configType.Elem()).Interface(), true
}

func (c *registry) CreateInputs(
	name string,
	config telegraf.PluginConfig,
) ([]telegraf.Input, error) {
	plugins, err := c.createPlugins(telegraf.InputType, name, config)
	if err != nil {
		return nil, err
	}

	inputs := plugins.([]telegraf.Input)
	return inputs, nil
}

func (c *registry) CreateParser(
	name string,
	config telegraf.PluginConfig,
) (telegraf.Parser, error) {
	plugins, err := c.createPlugins(telegraf.ParserType, name, config)
	if err != nil {
		return nil, err
	}

	parser := plugins.(telegraf.Parser)
	return parser, nil
}

func (c *registry) CreateOutputs(
	name string,
	config telegraf.PluginConfig,
) ([]telegraf.Output, error) {
	plugins, err := c.createPlugins(telegraf.OutputType, name, config)
	if err != nil {
		return nil, err
	}

	outputs := plugins.([]telegraf.Output)
	return outputs, nil
}

func (c *registry) CreateLoaders(
	name string,
	config telegraf.PluginConfig,
) ([]telegraf.Loader, error) {
	plugins, err := c.createPlugins(telegraf.LoaderType, name, config)
	if err != nil {
		return nil, err
	}

	loaders := plugins.([]telegraf.Loader)
	return loaders, nil
}

type registry struct {
	loaders map[string]telegraf.PluginFactory
	inputs  map[string]telegraf.PluginFactory
	outputs map[string]telegraf.PluginFactory
	parsers map[string]telegraf.PluginFactory
}

// configs provides access to plugins config structure by type and name.
type configs struct {
	registry registry
}

func (c *registry) createPlugins(
	pluginType telegraf.PluginType,
	name string,
	config telegraf.PluginConfig,
) (interface{}, error) {
	var factory telegraf.PluginFactory
	var ok bool

	switch pluginType {
	case telegraf.LoaderType:
		factory, ok = c.loaders[name]
	case telegraf.InputType:
		factory, ok = c.inputs[name]
	case telegraf.OutputType:
		factory, ok = c.outputs[name]
	case telegraf.ParserType:
		factory, ok = c.parsers[name]
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

	plugins := result[0].Interface()
	switch err := result[1].Interface().(type) {
	case error:
		return nil, err
	default:
		return plugins, nil
	}
}

// check validates the Plugin factories
func check(plugins map[string]telegraf.PluginFactory) error {
	for name, factory := range plugins {
		f := reflect.ValueOf(factory)

		// Check factory is a function
		if f.Kind() != reflect.Func {
			return fmt.Errorf("invalid factory %s", name)
		}

		// Check factory has one argument
		if f.Type().NumIn() != 1 {
			return fmt.Errorf("invalid arguments %s", name)
		}

		// Check argument is a ptr
		c := f.Type().In(0)
		if c.Kind() != reflect.Ptr {
			return fmt.Errorf("invalid argument type %s", name)
		}

		// TODO: Check return values: (pluginType, error)?
	}
	return nil
}
