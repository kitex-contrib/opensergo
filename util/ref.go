package util

import "reflect"

// ParamTypeName get request/response param type name
func ParamTypeName(inf interface{}) string {
	reflectValue := reflect.ValueOf(inf)
	if reflectValue.Kind() == reflect.Ptr {
		if reflectValue.IsNil() && reflectValue.CanAddr() {
			reflectValue.Set(reflect.New(reflectValue.Type().Elem()))
		}
		reflectValue = reflectValue.Elem()
	}

	if reflectValue.NumField() == 0 {
		return ""
	}

	field := reflectValue.Field(0)
	if field.Kind() == reflect.Ptr {
		if field.IsNil() && field.CanAddr() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}
	return field.Type().String()
}
