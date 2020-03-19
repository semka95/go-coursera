package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	e := reflect.Indirect(reflect.Indirect(reflect.ValueOf(&out)).Elem())
	if !e.CanSet() {
		return fmt.Errorf("not settable")
	}
	if e.Kind() == reflect.Slice {
		s := reflect.New(e.Type().Elem()).Interface()
		f, ok := data.([]interface{})
		if !ok {
			return fmt.Errorf("data is not slice")
		}

		for _, v := range f {
			err := i2s(v, s)
			if err != nil {
				return err
			}
			e.Set(reflect.Append(e, reflect.Indirect(reflect.ValueOf(s))))
		}
		return nil
	}

	f, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("not map")
	}

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
			_, ok := f[name].(map[string]interface{})
			if !ok {
				return fmt.Errorf("should be map")
			}
			err := i2s(f[name], e.Field(i).Addr().Interface())
			if err != nil {
				return err
			}
		case reflect.Slice:
			s := reflect.New(e.Type().Field(i).Type.Elem()).Interface()
			a, ok := f[name].([]interface{})
			if !ok {
				return fmt.Errorf("need array")
			}

			for _, v := range a {
				err := i2s(v, s)
				if err != nil {
					return err
				}
				e.Field(i).Set(reflect.Append(e.Field(i), reflect.Indirect(reflect.ValueOf(s))))
			}
		}
	}

	return nil
}
