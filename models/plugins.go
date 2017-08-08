package models

import (
	"fmt"
	"reflect"
)

type Plugins map[string]interface{}

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
