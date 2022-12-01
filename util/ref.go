// Copyright 2021 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
