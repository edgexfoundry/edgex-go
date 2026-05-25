//
// Copyright (C) 2020-2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	edgexErrors "github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type BaseReading struct {
	Id             string         `json:"id,omitempty"`
	Origin         int64          `json:"origin" validate:"required"`
	DeviceName     string         `json:"deviceName" validate:"required,edgex-dto-none-empty-string"`
	ResourceName   string         `json:"resourceName" validate:"required,edgex-dto-none-empty-string"`
	ProfileName    string         `json:"profileName" validate:"required,edgex-dto-none-empty-string"`
	ValueType      string         `json:"valueType" validate:"required,edgex-dto-value-type"`
	Units          string         `json:"units,omitempty"`
	Tags           Tags           `json:"tags,omitempty"`
	Extensions     map[string]any `json:"-" xml:"-"`
	BinaryReading  `json:",inline" validate:"-"`
	SimpleReading  `json:",inline" validate:"-"`
	ObjectReading  `json:",inline" validate:"-"`
	NullReading    `json:",inline" validate:"-"`
	NumericReading `json:",inline" validate:"-"`
}

type SimpleReading struct {
	Value string `json:"value"`
}

type BinaryReading struct {
	BinaryValue []byte `json:"binaryValue,omitempty" validate:"omitempty"`
	MediaType   string `json:"mediaType,omitempty" validate:"required_with=BinaryValue"`
}

type ObjectReading struct {
	ObjectValue any `json:"objectValue,omitempty"`
}

type NullReading struct {
	isNull bool // indicate the reading value should be null in the JSON payload
}

func (b BaseReading) IsNull() bool {
	return b.isNull
}

// NumericReading is used for querying event readings from core-data, requiring DB layer support for numeric data types (e.g., Postgres), and we still use SimpleReading for adding event readings.
type NumericReading struct {
	NumericValue any
}

func newBaseReading(profileName string, deviceName string, resourceName string, valueType string) BaseReading {
	return BaseReading{
		Id:           uuid.NewString(),
		Origin:       time.Now().UnixNano(),
		DeviceName:   deviceName,
		ResourceName: resourceName,
		ProfileName:  profileName,
		ValueType:    valueType,
		Extensions:   make(map[string]any),
		Tags:         make(Tags),
	}
}

// NewSimpleReading creates and returns a new initialized BaseReading with its SimpleReading initialized
func NewSimpleReading(profileName string, deviceName string, resourceName string, valueType string, value any) (BaseReading, error) {
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
func NewObjectReading(profileName string, deviceName string, resourceName string, objectValue any) BaseReading {
	reading := newBaseReading(profileName, deviceName, resourceName, common.ValueTypeObject)
	reading.ObjectReading = ObjectReading{
		ObjectValue: objectValue,
	}
	return reading
}

// NewObjectReadingWithArray creates and returns a new initialized BaseReading with its ObjectReading initialized with ObjectArray valueType
func NewObjectReadingWithArray(profileName string, deviceName string, resourceName string, objectValue any) BaseReading {
	reading := newBaseReading(profileName, deviceName, resourceName, common.ValueTypeObjectArray)
	reading.ObjectReading = ObjectReading{
		ObjectValue: objectValue,
	}
	return reading
}

// NewNullReading creates and returns a new initialized BaseReading with null
func NewNullReading(profileName string, deviceName string, resourceName string, valueType string) BaseReading {
	reading := newBaseReading(profileName, deviceName, resourceName, valueType)
	reading.isNull = true
	return reading
}

// NewNumericReading creates and returns a new initialized BaseReading with its NumericReading initialized
func NewNumericReading(profileName string, deviceName string, resourceName string, valueType string, value any) BaseReading {
	reading := newBaseReading(profileName, deviceName, resourceName, valueType)
	reading.NumericReading = NumericReading{
		NumericValue: value,
	}
	return reading
}

func convertInterfaceValue(valueType string, value any) (string, error) {
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
		arrayValue, ok := value.([]string)
		if !ok {
			return "", fmt.Errorf("unable to cast value to []string for %s", valueType)
		}

		return convertStringArrayValue(arrayValue)
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

func convertSimpleValue(valueType string, kind reflect.Kind, value any) (string, error) {
	if err := validateType(valueType, kind, value); err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", value), nil
}

func convertFloatValue(valueType string, kind reflect.Kind, value any) (string, error) {
	if err := validateType(valueType, kind, value); err != nil {
		return "", err
	}

	switch kind {
	case reflect.Float32:
		// as above has validated the value type/kind/value, it is safe to cast the value to float32 here
		float32Val, ok := value.(float32)
		if !ok {
			return "", fmt.Errorf("unable to cast value to float32 for %s", valueType)
		}
		return strconv.FormatFloat(float64(float32Val), 'e', -1, 32), nil
	case reflect.Float64:
		// as above has validated the value type/kind/value, it is safe to cast the value to float64 here
		float64Val, ok := value.(float64)
		if !ok {
			return "", fmt.Errorf("unable to cast value to float64 for %s", valueType)
		}
		return strconv.FormatFloat(float64Val, 'e', -1, 64), nil
	default:
		return "", fmt.Errorf("invalid kind %s to convert float value to string", kind.String())
	}
}

func convertStringArrayValue(values []string) (string, error) {
	result, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("failed to marshal StringArray value: %v", err)
	}
	return string(result), nil
}

func convertSimpleArrayValue(valueType string, kind reflect.Kind, value any) (string, error) {
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

func validateType(valueType string, kind reflect.Kind, value any) error {
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
	if b.isNull {
		return nil
	}
	switch b.ValueType {
	case common.ValueTypeBinary:
		// validate the inner BinaryReading struct
		binaryReading := b.BinaryReading
		if err := common.Validate(binaryReading); err != nil {
			return err
		}
	case common.ValueTypeObject, common.ValueTypeObjectArray:
		// validate the inner ObjectReading struct
		objectReading := b.ObjectReading
		if err := common.Validate(objectReading); err != nil {
			return err
		}
	default:
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
	if r.isNull {
		return models.NewNullReading(br)
	}
	switch r.ValueType {
	case common.ValueTypeBinary:
		readingModel = models.BinaryReading{
			BaseReading: br,
			BinaryValue: r.BinaryValue,
			MediaType:   r.MediaType,
		}
	case common.ValueTypeObject, common.ValueTypeObjectArray:
		readingModel = models.ObjectReading{
			BaseReading: br,
			ObjectValue: r.ObjectValue,
		}
	default:
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
	case models.NullReading:
		baseReading = BaseReading{
			Id:           r.Id,
			Origin:       r.Origin,
			DeviceName:   r.DeviceName,
			ResourceName: r.ResourceName,
			ProfileName:  r.ProfileName,
			ValueType:    r.ValueType,
			Units:        r.Units,
			Tags:         r.Tags,
			NullReading:  NullReading{isNull: true},
		}
	case models.NumericReading:
		baseReading = BaseReading{
			Id:             r.Id,
			Origin:         r.Origin,
			DeviceName:     r.DeviceName,
			ResourceName:   r.ResourceName,
			ProfileName:    r.ProfileName,
			ValueType:      r.ValueType,
			Units:          r.Units,
			Tags:           r.Tags,
			NumericReading: NumericReading{NumericValue: r.NumericValue},
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
// Note that this function will only work on readings with 'Object' or 'ObjectArray' valueType.  An error will be returned when invoking
// this function on a reading with valueType other than 'Object' or 'ObjectArray'.
func (b BaseReading) UnmarshalObjectValue(target any) error {
	if b.ValueType == common.ValueTypeObject || b.ValueType == common.ValueTypeObjectArray {
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

func (b BaseReading) MarshalJSON() ([]byte, error) {
	data, err := b.marshal(json.Marshal)
	if err != nil || len(b.Extensions) == 0 {
		return data, err
	}
	return mergeExtensions(data, b.Extensions, json.Unmarshal, json.Marshal)
}

func (b BaseReading) MarshalCBOR() ([]byte, error) {
	data, err := b.marshal(cbor.Marshal)
	if err != nil || len(b.Extensions) == 0 {
		return data, err
	}
	return mergeExtensions(data, b.Extensions, cbor.Unmarshal, cbor.Marshal)
}

func (b BaseReading) marshal(marshal func(any) ([]byte, error)) ([]byte, error) {
	type reading struct {
		Id           string `json:"id,omitempty" cbor:"id,omitempty"`
		Origin       int64  `json:"origin,omitempty" cbor:"origin,omitempty"`
		DeviceName   string `json:"deviceName,omitempty" cbor:"deviceName,omitempty"`
		ResourceName string `json:"resourceName,omitempty" cbor:"resourceName,omitempty"`
		ProfileName  string `json:"profileName,omitempty" cbor:"profileName,omitempty"`
		ValueType    string `json:"valueType" cbor:"valueType"`
		Units        string `json:"units,omitempty" cbor:"units,omitempty"`
		Tags         Tags   `json:"tags,omitempty" cbor:"tags,omitempty"`
	}
	if b.isNull {
		return marshal(&struct {
			reading     `json:",inline"`
			Value       any `json:"value"`
			BinaryValue any `json:"binaryValue"`
			ObjectValue any `json:"objectValue"`
		}{
			reading: reading{
				Id:           b.Id,
				Origin:       b.Origin,
				DeviceName:   b.DeviceName,
				ResourceName: b.ResourceName,
				ProfileName:  b.ProfileName,
				ValueType:    b.ValueType,
				Units:        b.Units,
				Tags:         b.Tags,
			},
			Value:       nil,
			BinaryValue: nil,
			ObjectValue: nil,
		})
	}
	r := reading{
		Id:           b.Id,
		Origin:       b.Origin,
		DeviceName:   b.DeviceName,
		ResourceName: b.ResourceName,
		ProfileName:  b.ProfileName,
		ValueType:    b.ValueType,
		Units:        b.Units,
		Tags:         b.Tags,
	}
	switch b.ValueType {
	case common.ValueTypeObject, common.ValueTypeObjectArray:
		return marshal(&struct {
			reading       `json:",inline"`
			ObjectReading `json:",inline" validate:"-"`
		}{
			reading:       r,
			ObjectReading: b.ObjectReading,
		})
	case common.ValueTypeBinary:
		return marshal(&struct {
			reading       `json:",inline"`
			BinaryReading `json:",inline" validate:"-"`
		}{
			reading:       r,
			BinaryReading: b.BinaryReading,
		})
	case common.ValueTypeUint8, common.ValueTypeUint16, common.ValueTypeUint32, common.ValueTypeUint64,
		common.ValueTypeInt8, common.ValueTypeInt16, common.ValueTypeInt32, common.ValueTypeInt64,
		common.ValueTypeFloat32, common.ValueTypeFloat64,
		common.ValueTypeInt8Array, common.ValueTypeInt16Array, common.ValueTypeInt32Array, common.ValueTypeInt64Array,
		common.ValueTypeUint8Array, common.ValueTypeUint16Array, common.ValueTypeUint32Array, common.ValueTypeUint64Array,
		common.ValueTypeFloat32Array, common.ValueTypeFloat64Array:
		// the reading value should be stored in SimpleReading or NumericReading
		if b.NumericValue == nil {
			return marshal(&struct {
				reading       `json:",inline"`
				SimpleReading `json:",inline" validate:"-"`
			}{
				reading:       r,
				SimpleReading: b.SimpleReading,
			})
		} else {
			type numericReading struct { // define another struct to output the numeric reading value
				Value any `json:"value"`
			}
			return marshal(&struct {
				reading        `json:",inline"`
				numericReading `json:",inline" validate:"-"`
			}{
				reading:        r,
				numericReading: numericReading{Value: b.NumericValue},
			})
		}
	default:
		return marshal(&struct {
			reading       `json:",inline"`
			SimpleReading `json:",inline" validate:"-"`
		}{
			reading:       r,
			SimpleReading: b.SimpleReading,
		})
	}
}

func (b *BaseReading) UnmarshalJSON(data []byte) error {
	return b.Unmarshal(data, jsonUnmarshalUseNumber)
}

func (b *BaseReading) UnmarshalCBOR(data []byte) error {
	return b.Unmarshal(data, cbor.Unmarshal)
}

func (b *BaseReading) Unmarshal(data []byte, unmarshal func([]byte, any) error) error {
	var rawMap map[string]any
	if err := unmarshal(data, &rawMap); err != nil {
		return err
	}
	// When cbor.Unmarshal decodes into map[string]any, the top-level keys are strings,
	// but nested map values (e.g. tags, objectValue) are decoded as map[any]any because
	// their target type is any. normalizeMap recursively converts these to map[string]any
	// so that subsequent type assertions work correctly for both JSON and CBOR paths.
	normalizeMap(rawMap)
	return b.populateFromMap(rawMap)
}

func (b *BaseReading) populateFromMap(rawMap map[string]any) error {
	*b = BaseReading{}

	var err error
	if b.Id, err = popStringValueFromKey(rawMap, keyId); err != nil {
		return err
	}
	if b.DeviceName, err = popStringValueFromKey(rawMap, keyDeviceName); err != nil {
		return err
	}
	if b.ResourceName, err = popStringValueFromKey(rawMap, keyResourceName); err != nil {
		return err
	}
	if b.ProfileName, err = popStringValueFromKey(rawMap, keyProfileName); err != nil {
		return err
	}
	if b.ValueType, err = popStringValueFromKey(rawMap, keyValueType); err != nil {
		return err
	}
	if b.Units, err = popStringValueFromKey(rawMap, keyUnits); err != nil {
		return err
	}

	switch v := popKey(rawMap, keyOrigin).(type) {
	case json.Number:
		if b.Origin, err = v.Int64(); err != nil {
			return fmt.Errorf("failed to decode origin: %w", err)
		}
	case uint64: // CBOR, positive integers decode as uint64
		if v > math.MaxInt64 {
			return fmt.Errorf("origin value %d overflows int64", v)
		}
		b.Origin = int64(v)
	case int64: // CBOR, negative integers decode as int64
		b.Origin = v
	case nil:
		// key absent — leave Origin as zero
	default:
		return fmt.Errorf("failed to decode origin: unsupported type %T", v)
	}

	// convert json.Number in rawMap to native numeric types before assigning Tags/Extensions
	convertJSONNumbers(rawMap)

	if rawTags := popKey(rawMap, keyTags); rawTags != nil {
		if tags, ok := rawTags.(map[string]any); ok {
			b.Tags = tags
		} else {
			return fmt.Errorf("failed to decode tags: expected map[string]any, got %T", rawTags)
		}
	} else {
		b.Tags = make(map[string]any)
	}

	// BinaryReading: JSON gives base64 string, CBOR gives []byte
	switch v := popKey(rawMap, keyBinaryValue).(type) {
	case string:
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return fmt.Errorf("failed to decode binaryValue: %w", err)
		}
		b.BinaryValue = decoded
	case []byte:
		b.BinaryValue = v
	}
	if b.MediaType, err = popStringValueFromKey(rawMap, keyMediaType); err != nil {
		return err
	}

	// ObjectReading
	objectValue := popKey(rawMap, keyObjectValue)
	b.ObjectValue = objectValue

	// Value -> SimpleReading / NumericReading
	value := popKey(rawMap, keyValue)
	if value != nil {
		b.SimpleReading = SimpleReading{Value: fmt.Sprintf("%v", value)}
	}
	// if the value is not string, let the NumericReading contain the equivalent value
	if _, ok := value.(string); !ok {
		b.NumericReading = NumericReading{NumericValue: value}
	}

	switch b.ValueType {
	case common.ValueTypeObject, common.ValueTypeObjectArray:
		if objectValue == nil {
			b.isNull = true
		}
	case common.ValueTypeBinary:
		if b.BinaryValue == nil {
			b.isNull = true
		}
	default:
		if value == nil {
			b.isNull = true
		}
	}

	// remaining keys are extensions
	if len(rawMap) > 0 {
		b.Extensions = rawMap
	} else {
		b.Extensions = make(map[string]any)
	}
	return nil
}
