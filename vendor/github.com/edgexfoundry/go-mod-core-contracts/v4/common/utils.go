//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
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
	ValueTypeObject, ValueTypeObjectArray,
}

// NormalizeValueType normalizes the valueType to upper camel case
func NormalizeValueType(valueType string) (string, error) {
	for _, v := range valueTypes {
		if strings.EqualFold(valueType, v) {
			return v, nil
		}
	}
	return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unable to normalize the unknown value type %s", valueType), nil)
}

// BuildTopic is a helper function to build MessageBus topic from multiple parts
func BuildTopic(parts ...string) string {
	return strings.Join(parts, "/")
}

// URLEncode encodes the input string with additional common character support
func URLEncode(s string) string {
	res := url.PathEscape(s)
	res = strings.Replace(res, "+", "%2B", -1) // MQTT topic reserved char
	res = strings.Replace(res, "-", "%2D", -1)
	res = strings.Replace(res, ".", "%2E", -1) // RegexCmd and Redis topic reserved char
	res = strings.Replace(res, "_", "%5F", -1)
	res = strings.Replace(res, "~", "%7E", -1)

	return res
}

type pathBuilder struct {
	sb                    strings.Builder
	enableNameFieldEscape bool
}

func NewPathBuilder() *pathBuilder {
	return &pathBuilder{}
}

func (b *pathBuilder) EnableNameFieldEscape(enableNameFieldEscape bool) *pathBuilder {
	b.enableNameFieldEscape = enableNameFieldEscape
	return b
}

func (b *pathBuilder) SetPath(path string) *pathBuilder {
	b.sb.WriteString(path + "/")
	return b
}

// SetNameFieldPath set name path, such as device name, profile name, interval name
func (b *pathBuilder) SetNameFieldPath(namePath string) *pathBuilder {
	if b.enableNameFieldEscape {
		namePath = URLEncode(namePath)
	}
	b.sb.WriteString(namePath + "/")
	return b
}

func (b *pathBuilder) BuildPath() string {
	return strings.TrimSuffix(b.sb.String(), "/")
}

// StrValueFromProperties retrieves the string value associated with the given key from the provided properties map.
func StrValueFromProperties(key string, properties map[string]any) (val string) {
	if propVal, ok := properties[key]; ok {
		val = fmt.Sprintf("%v", propVal)
	}
	return val
}

// StrSliceFromProperties retrieves a slice of strings associated with the given key from the provided properties map.
func StrSliceFromProperties(key string, properties map[string]any) []string {
	if propVal, ok := properties[key]; ok {
		switch slice := propVal.(type) {
		case []string:
			return slice
		case []any:
			strSlice := make([]string, len(slice))
			for i, v := range slice {
				strSlice[i] = fmt.Sprintf("%v", v)
			}
			return strSlice
		}
	}
	return []string{}
}
