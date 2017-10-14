package models

import (
	"fmt"
	"reflect"

	"github.com/influxdata/tgconfig"
)

type Plugins map[string]interface{}

type Factories struct {
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
) (*Factories, error) {
	_, err := Check(loaders)
	if err != nil {
		return nil, err
	}
	_, err = Check(inputs)
	if err != nil {
		return nil, err
	}
	_, err = Check(outputs)
	if err != nil {
		return nil, err
	}
	_, err = Check(parsers)
	if err != nil {
		return nil, err
	}

	return &Factories{
		Loaders: loaders,
		Inputs:  inputs,
		Outputs: outputs,
		Parsers: parsers,
	}, nil
}

func (c *Factories) GetFactory(
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

	if !ok {
		return nil, false
	}

	return factory, true
}

func (c *Factories) GetConfigRegistry() telegraf.ConfigRegistry {
	configs := Configs(*c)
	return &configs
}

// ConfigRegistry holds the set of available plugins.  This provides a layer of
// indirection so that you can define a custom set of plugins.
//
// plugin_name -> plugin_factory
// i.e.: "cpu" -> cpu.New(*cpu.Config) (Input, error)
type Configs struct {
	Loaders map[string]telegraf.PluginFactory
	Inputs  map[string]telegraf.PluginFactory
	Outputs map[string]telegraf.PluginFactory
	Parsers map[string]telegraf.PluginFactory
}

func (c *Configs) GetPluginConfig(
	pluginType telegraf.PluginType,
	name string,
) (telegraf.PluginConfig, bool) {
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
		return nil, false
	}

	vfactory := reflect.ValueOf(factory)

	// Get the Type of the first and only argument
	configType := vfactory.Type().In(0)

	// Create a new config struct
	return reflect.New(configType.Elem()).Interface(), true
}

func Check(plugins Plugins) (Plugins, error) {
	for name, factory := range plugins {
		f := reflect.ValueOf(factory)
		if f.Kind() != reflect.Func {
			return nil, fmt.Errorf("invalid factory %s", name)
		}

		if f.Type().NumIn() != 1 {
			return nil, fmt.Errorf("invalid arguments %s", name)
		}

		c := f.Type().In(0)
		if c.Kind() != reflect.Ptr {
			return nil, fmt.Errorf("invalid argument type %s", name)
		}
	}
	return plugins, nil
}
