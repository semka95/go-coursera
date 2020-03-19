package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	e := reflect.Indirect(reflect.Indirect(reflect.ValueOf(&out)).Elem())
	f := data.(map[string]interface{})

	for i := 0; i < e.NumField(); i++ {
		name := e.Type().Field(i).Name
		switch b := e.Type().Field(i).Type.Kind(); b {
		case reflect.Int:
			val, ok := f[name].(float64)
			if !ok {
				return fmt.Errorf("int err")
			}
			e.Field(i).SetInt(int64(val))
		case reflect.Float64:
			val, ok := f[name].(float64)
			if !ok {
				return fmt.Errorf("float err")
			}
			e.Field(i).SetFloat(val)
		case reflect.String:
			val, ok := f[name].(string)
			if !ok {
				return fmt.Errorf("string err")
			}
			e.Field(i).SetString(val)
		case reflect.Bool:
			val, ok := f[name].(bool)
			if !ok {
				return fmt.Errorf("bool err")
			}
			e.Field(i).SetBool(val)
		case reflect.Struct:
			err := i2s(f[name], e.Field(i).Addr().Interface())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
