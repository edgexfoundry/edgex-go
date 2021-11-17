//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

var valueTypes = []string{
	ValueTypeBool, ValueTypeString,
	ValueTypeUint8, ValueTypeUint16, ValueTypeUint32, ValueTypeUint64,
	ValueTypeInt8, ValueTypeInt16, ValueTypeInt32, ValueTypeInt64,
	ValueTypeFloat32, ValueTypeFloat64,
	ValueTypeBinary,
	ValueTypeBoolArray, ValueTypeStringArray,
	ValueTypeUint8Array, ValueTypeUint16Array, ValueTypeUint32Array, ValueTypeUint64Array,
	ValueTypeInt8Array, ValueTypeInt16Array, ValueTypeInt32Array, ValueTypeInt64Array,
	ValueTypeFloat32Array, ValueTypeFloat64Array,
	ValueTypeObject,
}

// // NormalizeValueType normalizes the valueType to upper camel case
func NormalizeValueType(valueType string) (string, error) {
	for _, v := range valueTypes {
		if strings.ToLower(valueType) == strings.ToLower(v) {
			return v, nil
		}
	}
	return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unable to normalize the unknown value type %s", valueType), nil)
}
