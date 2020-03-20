package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	resStruct := reflect.Indirect(reflect.Indirect(reflect.ValueOf(&out)).Elem())
	if !resStruct.CanSet() {
		return fmt.Errorf("object is not settable, pass reference")
	}

	if resStruct.Kind() == reflect.Slice {
		newStruct := reflect.New(resStruct.Type().Elem()).Interface()
		inputData, ok := data.([]interface{})
		if !ok {
			return fmt.Errorf("expected input data to be a slice")
		}

		for _, v := range inputData {
			err := i2s(v, newStruct)
			if err != nil {
				return err
			}
			structValue := reflect.Indirect(reflect.ValueOf(newStruct))
			resStruct.Set(reflect.Append(resStruct, structValue))
		}
		return nil
	}

	inputData, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected input data to be a map")
	}

	for i := 0; i < resStruct.NumField(); i++ {
		fieldName := resStruct.Type().Field(i).Name
		switch resStruct.Type().Field(i).Type.Kind() {
		case reflect.Int:
			val, ok := inputData[fieldName].(float64)
			if !ok {
				return fmt.Errorf("field %s expected to be int, but got %v", fieldName, inputData[fieldName])
			}
			resStruct.Field(i).SetInt(int64(val))
		case reflect.Float64:
			val, ok := inputData[fieldName].(float64)
			if !ok {
				return fmt.Errorf("field %s expected to be float, but got %v", fieldName, inputData[fieldName])
			}
			resStruct.Field(i).SetFloat(val)
		case reflect.String:
			val, ok := inputData[fieldName].(string)
			if !ok {
				return fmt.Errorf("field %s expected to be string, but got %v", fieldName, inputData[fieldName])
			}
			resStruct.Field(i).SetString(val)
		case reflect.Bool:
			val, ok := inputData[fieldName].(bool)
			if !ok {
				return fmt.Errorf("field %s expected to be bool, but got %v", fieldName, inputData[fieldName])
			}
			resStruct.Field(i).SetBool(val)
		case reflect.Struct:
			_, ok := inputData[fieldName].(map[string]interface{})
			if !ok {
				return fmt.Errorf("expected input data to be a map")
			}
			err := i2s(inputData[fieldName], resStruct.Field(i).Addr().Interface())
			if err != nil {
				return err
			}
		case reflect.Slice:
			newStruct := reflect.New(resStruct.Type().Field(i).Type.Elem()).Interface()
			dataSlice, ok := inputData[fieldName].([]interface{})
			if !ok {
				return fmt.Errorf("expected input data to be a slice")
			}

			for _, v := range dataSlice {
				err := i2s(v, newStruct)
				if err != nil {
					return err
				}
				structValue := reflect.Indirect(reflect.ValueOf(newStruct))
				resStruct.Field(i).Set(reflect.Append(resStruct.Field(i), structValue))
			}
		}
	}

	return nil
}
