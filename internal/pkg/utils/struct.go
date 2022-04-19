//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"reflect"
)

func OnlyOneFieldUpdated(fieldName string, model interface{}) bool {
	res := true

	t := reflect.TypeOf(model)
	v := reflect.ValueOf(model)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name != fieldName && f.Name != "Name" && f.Name != "Id" {
			if !v.FieldByName(f.Name).IsNil() {
				res = false
				break
			}
		}
	}

	return res
}
