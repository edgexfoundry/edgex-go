//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	edgexErrors "github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

type BaseReading struct {
	Id            string `json:"id,omitempty"`
	Origin        int64  `json:"origin" validate:"required"`
	DeviceName    string `json:"deviceName" validate:"required,edgex-dto-none-empty-string"`
	ResourceName  string `json:"resourceName" validate:"required"`
	ProfileName   string `json:"profileName" validate:"required,edgex-dto-none-empty-string"`
	ValueType     string `json:"valueType" validate:"required,edgex-dto-value-type"`
	Units         string `json:"units,omitempty"`
	Tags          Tags   `json:"tags,omitempty"`
	BinaryReading `json:",inline" validate:"-"`
	SimpleReading `json:",inline" validate:"-"`
	ObjectReading `json:",inline" validate:"-"`
}

type SimpleReading struct {
	Value string `json:"value"`
}

type BinaryReading struct {
	BinaryValue []byte `json:"binaryValue,omitempty" validate:"gt=0,required"`
	MediaType   string `json:"mediaType,omitempty" validate:"required"`
}

type ObjectReading struct {
	ObjectValue interface{} `json:"objectValue,omitempty" validate:"required"`
}

func newBaseReading(profileName string, deviceName string, resourceName string, valueType string) BaseReading {
	return BaseReading{
		Id:           uuid.NewString(),
		Origin:       time.Now().UnixNano(),
		DeviceName:   deviceName,
		ResourceName: resourceName,
		ProfileName:  profileName,
		ValueType:    valueType,
	}
}

// NewSimpleReading creates and returns a new initialized BaseReading with its SimpleReading initialized
func NewSimpleReading(profileName string, deviceName string, resourceName string, valueType string, value interface{}) (BaseReading, error) {
	stringValue, err := convertInterfaceValue(valueType, value)
	if err != nil {
		return BaseReading{}, err
	}

	reading := newBaseReading(profileName, deviceName, resourceName, valueType)
	reading.SimpleReading = SimpleReading{
		Value: stringValue,
	}
	return reading, nil
}

// NewBinaryReading creates and returns a new initialized BaseReading with its BinaryReading initialized
func NewBinaryReading(profileName string, deviceName string, resourceName string, binaryValue []byte, mediaType string) BaseReading {
	reading := newBaseReading(profileName, deviceName, resourceName, common.ValueTypeBinary)
	reading.BinaryReading = BinaryReading{
		BinaryValue: binaryValue,
		MediaType:   mediaType,
	}
	return reading
}

// NewObjectReading creates and returns a new initialized BaseReading with its ObjectReading initialized
func NewObjectReading(profileName string, deviceName string, resourceName string, objectValue interface{}) BaseReading {
	reading := newBaseReading(profileName, deviceName, resourceName, common.ValueTypeObject)
	reading.ObjectReading = ObjectReading{
		ObjectValue: objectValue,
	}
	return reading
}

func convertInterfaceValue(valueType string, value interface{}) (string, error) {
	switch valueType {
	case common.ValueTypeBool:
		return convertSimpleValue(valueType, reflect.Bool, value)
	case common.ValueTypeString:
		return convertSimpleValue(valueType, reflect.String, value)

	case common.ValueTypeUint8:
		return convertSimpleValue(valueType, reflect.Uint8, value)
	case common.ValueTypeUint16:
		return convertSimpleValue(valueType, reflect.Uint16, value)
	case common.ValueTypeUint32:
		return convertSimpleValue(valueType, reflect.Uint32, value)
	case common.ValueTypeUint64:
		return convertSimpleValue(valueType, reflect.Uint64, value)

	case common.ValueTypeInt8:
		return convertSimpleValue(valueType, reflect.Int8, value)
	case common.ValueTypeInt16:
		return convertSimpleValue(valueType, reflect.Int16, value)
	case common.ValueTypeInt32:
		return convertSimpleValue(valueType, reflect.Int32, value)
	case common.ValueTypeInt64:
		return convertSimpleValue(valueType, reflect.Int64, value)

	case common.ValueTypeFloat32:
		return convertFloatValue(valueType, reflect.Float32, value)
	case common.ValueTypeFloat64:
		return convertFloatValue(valueType, reflect.Float64, value)

	case common.ValueTypeBoolArray:
		return convertSimpleArrayValue(valueType, reflect.Bool, value)
	case common.ValueTypeStringArray:
		return convertSimpleArrayValue(valueType, reflect.String, value)

	case common.ValueTypeUint8Array:
		return convertSimpleArrayValue(valueType, reflect.Uint8, value)
	case common.ValueTypeUint16Array:
		return convertSimpleArrayValue(valueType, reflect.Uint16, value)
	case common.ValueTypeUint32Array:
		return convertSimpleArrayValue(valueType, reflect.Uint32, value)
	case common.ValueTypeUint64Array:
		return convertSimpleArrayValue(valueType, reflect.Uint64, value)

	case common.ValueTypeInt8Array:
		return convertSimpleArrayValue(valueType, reflect.Int8, value)
	case common.ValueTypeInt16Array:
		return convertSimpleArrayValue(valueType, reflect.Int16, value)
	case common.ValueTypeInt32Array:
		return convertSimpleArrayValue(valueType, reflect.Int32, value)
	case common.ValueTypeInt64Array:
		return convertSimpleArrayValue(valueType, reflect.Int64, value)

	case common.ValueTypeFloat32Array:
		arrayValue, ok := value.([]float32)
		if !ok {
			return "", fmt.Errorf("unable to cast value to []float32 for %s", valueType)
		}

		return convertFloat32ArrayValue(arrayValue)
	case common.ValueTypeFloat64Array:
		arrayValue, ok := value.([]float64)
		if !ok {
			return "", fmt.Errorf("unable to cast value to []float64 for %s", valueType)
		}

		return convertFloat64ArrayValue(arrayValue)

	default:
		return "", fmt.Errorf("invalid simple reading type of %s", valueType)
	}
}

func convertSimpleValue(valueType string, kind reflect.Kind, value interface{}) (string, error) {
	if err := validateType(valueType, kind, value); err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", value), nil
}

func convertFloatValue(valueType string, kind reflect.Kind, value interface{}) (string, error) {
	if err := validateType(valueType, kind, value); err != nil {
		return "", err
	}

	return fmt.Sprintf("%e", value), nil
}

func convertSimpleArrayValue(valueType string, kind reflect.Kind, value interface{}) (string, error) {
	if err := validateType(valueType, kind, value); err != nil {
		return "", err
	}

	result := fmt.Sprintf("%v", value)
	result = strings.ReplaceAll(result, " ", ", ")
	return result, nil
}

func convertFloat32ArrayValue(values []float32) (string, error) {
	var result strings.Builder
	result.WriteString("[")
	first := true
	for _, value := range values {
		if first {
			floatValue, err := convertFloatValue(common.ValueTypeFloat32, reflect.Float32, value)
			if err != nil {
				return "", err
			}
			result.WriteString(floatValue)
			first = false
			continue
		}

		floatValue, err := convertFloatValue(common.ValueTypeFloat32, reflect.Float32, value)
		if err != nil {
			return "", err
		}
		result.WriteString(", " + floatValue)
	}

	result.WriteString("]")
	return result.String(), nil
}

func convertFloat64ArrayValue(values []float64) (string, error) {
	var result strings.Builder
	result.WriteString("[")
	first := true
	for _, value := range values {
		if first {
			floatValue, err := convertFloatValue(common.ValueTypeFloat64, reflect.Float64, value)
			if err != nil {
				return "", err
			}
			result.WriteString(floatValue)
			first = false
			continue
		}

		floatValue, err := convertFloatValue(common.ValueTypeFloat64, reflect.Float64, value)
		if err != nil {
			return "", err
		}
		result.WriteString(", " + floatValue)
	}

	result.WriteString("]")
	return result.String(), nil
}

func validateType(valueType string, kind reflect.Kind, value interface{}) error {
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		if kind != reflect.TypeOf(value).Elem().Kind() {
			return fmt.Errorf("slice of type of value `%s` not a match for specified ValueType '%s", kind.String(), valueType)
		}
		return nil
	}

	if kind != reflect.TypeOf(value).Kind() {
		return fmt.Errorf("type of value `%s` not a match for specified ValueType '%s", kind.String(), valueType)
	}

	return nil
}

// Validate satisfies the Validator interface
func (b BaseReading) Validate() error {
	if b.ValueType == common.ValueTypeBinary {
		// validate the inner BinaryReading struct
		binaryReading := b.BinaryReading
		if err := common.Validate(binaryReading); err != nil {
			return err
		}
	} else if b.ValueType == common.ValueTypeObject {
		// validate the inner ObjectReading struct
		objectReading := b.ObjectReading
		if err := common.Validate(objectReading); err != nil {
			return err
		}
	} else {
		// validate the inner SimpleReading struct
		simpleReading := b.SimpleReading
		if err := common.Validate(simpleReading); err != nil {
			return err
		}
		if err := ValidateValue(b.ValueType, simpleReading.Value); err != nil {
			return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, fmt.Sprintf("The value does not match the %v valueType", b.ValueType), nil)
		}
	}

	return nil
}

// ToReadingModel converts Reading DTO to Reading Model
func ToReadingModel(r BaseReading) models.Reading {
	var readingModel models.Reading
	br := models.BaseReading{
		Id:           r.Id,
		Origin:       r.Origin,
		DeviceName:   r.DeviceName,
		ResourceName: r.ResourceName,
		ProfileName:  r.ProfileName,
		ValueType:    r.ValueType,
		Units:        r.Units,
		Tags:         r.Tags,
	}
	if r.ValueType == common.ValueTypeBinary {
		readingModel = models.BinaryReading{
			BaseReading: br,
			BinaryValue: r.BinaryValue,
			MediaType:   r.MediaType,
		}
	} else if r.ValueType == common.ValueTypeObject {
		readingModel = models.ObjectReading{
			BaseReading: br,
			ObjectValue: r.ObjectValue,
		}
	} else {
		readingModel = models.SimpleReading{
			BaseReading: br,
			Value:       r.Value,
		}
	}
	return readingModel
}

func FromReadingModelToDTO(reading models.Reading) BaseReading {
	var baseReading BaseReading
	switch r := reading.(type) {
	case models.BinaryReading:
		baseReading = BaseReading{
			Id:            r.Id,
			Origin:        r.Origin,
			DeviceName:    r.DeviceName,
			ResourceName:  r.ResourceName,
			ProfileName:   r.ProfileName,
			ValueType:     r.ValueType,
			Units:         r.Units,
			Tags:          r.Tags,
			BinaryReading: BinaryReading{BinaryValue: r.BinaryValue, MediaType: r.MediaType},
		}
	case models.ObjectReading:
		baseReading = BaseReading{
			Id:            r.Id,
			Origin:        r.Origin,
			DeviceName:    r.DeviceName,
			ResourceName:  r.ResourceName,
			ProfileName:   r.ProfileName,
			ValueType:     r.ValueType,
			Units:         r.Units,
			Tags:          r.Tags,
			ObjectReading: ObjectReading{ObjectValue: r.ObjectValue},
		}
	case models.SimpleReading:
		baseReading = BaseReading{
			Id:            r.Id,
			Origin:        r.Origin,
			DeviceName:    r.DeviceName,
			ResourceName:  r.ResourceName,
			ProfileName:   r.ProfileName,
			ValueType:     r.ValueType,
			Units:         r.Units,
			Tags:          r.Tags,
			SimpleReading: SimpleReading{Value: r.Value},
		}
	}

	return baseReading
}

// ValidateValue used to check whether the value and valueType are matched
func ValidateValue(valueType string, value string) error {
	if strings.Contains(valueType, "Array") {
		return parseArrayValue(valueType, value)
	} else {
		return parseSimpleValue(valueType, value)
	}
}

func parseSimpleValue(valueType string, value string) (err error) {
	switch valueType {
	case common.ValueTypeBool:
		_, err = strconv.ParseBool(value)

	case common.ValueTypeUint8:
		_, err = strconv.ParseUint(value, 10, 8)
	case common.ValueTypeUint16:
		_, err = strconv.ParseUint(value, 10, 16)
	case common.ValueTypeUint32:
		_, err = strconv.ParseUint(value, 10, 32)
	case common.ValueTypeUint64:
		_, err = strconv.ParseUint(value, 10, 64)

	case common.ValueTypeInt8:
		_, err = strconv.ParseInt(value, 10, 8)
	case common.ValueTypeInt16:
		_, err = strconv.ParseInt(value, 10, 16)
	case common.ValueTypeInt32:
		_, err = strconv.ParseInt(value, 10, 32)
	case common.ValueTypeInt64:
		_, err = strconv.ParseInt(value, 10, 64)

	case common.ValueTypeFloat32:
		_, err = strconv.ParseFloat(value, 32)
	case common.ValueTypeFloat64:
		_, err = strconv.ParseFloat(value, 64)
	}

	if err != nil {
		return err
	}
	return nil
}

func parseArrayValue(valueType string, value string) (err error) {
	arrayValue := strings.Split(value[1:len(value)-1], ",") // trim "[" and "]"

	for _, v := range arrayValue {
		v = strings.TrimSpace(v)
		switch valueType {
		case common.ValueTypeBoolArray:
			err = parseSimpleValue(common.ValueTypeBool, v)

		case common.ValueTypeUint8Array:
			err = parseSimpleValue(common.ValueTypeUint8, v)
		case common.ValueTypeUint16Array:
			err = parseSimpleValue(common.ValueTypeUint16, v)
		case common.ValueTypeUint32Array:
			err = parseSimpleValue(common.ValueTypeUint32, v)
		case common.ValueTypeUint64Array:
			err = parseSimpleValue(common.ValueTypeUint64, v)

		case common.ValueTypeInt8Array:
			err = parseSimpleValue(common.ValueTypeInt8, v)
		case common.ValueTypeInt16Array:
			err = parseSimpleValue(common.ValueTypeInt16, v)
		case common.ValueTypeInt32Array:
			err = parseSimpleValue(common.ValueTypeInt32, v)
		case common.ValueTypeInt64Array:
			err = parseSimpleValue(common.ValueTypeInt64, v)

		case common.ValueTypeFloat32Array:
			err = parseSimpleValue(common.ValueTypeFloat32, v)
		case common.ValueTypeFloat64Array:
			err = parseSimpleValue(common.ValueTypeFloat64, v)

		}
		if err != nil {
			return err
		}
	}
	return nil
}

// UnmarshalObjectValue is a helper function used to unmarshal the ObjectValue of a reading to the passed in target type.
// Note that this function will only work on readings with 'Object' valueType.  An error will be returned when invoking
// this function on a reading with valueType other than 'Object'.
func (b BaseReading) UnmarshalObjectValue(target any) error {
	if b.ValueType == common.ValueTypeObject {
		// marshal the current reading ObjectValue to JSON
		jsonEncodedData, err := json.Marshal(b.ObjectValue)
		if err != nil {
			return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, "failed to encode the object value of reading to JSON", err)
		}
		// unmarshal the JSON into the passed in target
		err = json.Unmarshal(jsonEncodedData, target)
		if err != nil {
			return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, fmt.Sprintf("failed to unmarshall the object value of reading into type %v", reflect.TypeOf(target).String()), err)
		}
	} else {
		return edgexErrors.NewCommonEdgeX(edgexErrors.KindContractInvalid, fmt.Sprintf("invalid usage of UnmarshalObjectValue function invocation on reading with %v valueType", b.ValueType), nil)
	}

	return nil
}
