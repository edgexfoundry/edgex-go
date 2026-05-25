//
// Copyright (C) 2020-2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	dtoCommon "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type Event struct {
	dtoCommon.Versionable `json:",inline"`
	Id                    string         `json:"id" validate:"required,uuid"`
	DeviceName            string         `json:"deviceName" validate:"required,edgex-dto-none-empty-string"`
	ProfileName           string         `json:"profileName" validate:"required,edgex-dto-none-empty-string"`
	SourceName            string         `json:"sourceName" validate:"required,edgex-dto-none-empty-string"`
	Origin                int64          `json:"origin" validate:"required"`
	Readings              []BaseReading  `json:"readings" validate:"gt=0,dive,required"`
	Tags                  Tags           `json:"tags,omitempty"`
	Extensions            map[string]any `json:"-" xml:"-"`
}

// NewEvent creates and returns an initialized Event with no Readings
func NewEvent(profileName, deviceName, sourceName string) Event {
	return Event{
		Versionable: dtoCommon.NewVersionable(),
		Id:          uuid.NewString(),
		DeviceName:  deviceName,
		ProfileName: profileName,
		SourceName:  sourceName,
		Origin:      time.Now().UnixNano(),
		Readings:    make([]BaseReading, 0),
		Extensions:  make(map[string]any),
		Tags:        make(Tags),
	}
}

// FromEventModelToDTO transforms the Event Model to the Event DTO
func FromEventModelToDTO(event models.Event) Event {
	var readings []BaseReading
	for _, reading := range event.Readings {
		readings = append(readings, FromReadingModelToDTO(reading))
	}

	tags := make(map[string]interface{})
	for tag, value := range event.Tags {
		tags[tag] = value
	}

	return Event{
		Versionable: dtoCommon.NewVersionable(),
		Id:          event.Id,
		DeviceName:  event.DeviceName,
		ProfileName: event.ProfileName,
		SourceName:  event.SourceName,
		Origin:      event.Origin,
		Readings:    readings,
		Tags:        tags,
	}
}

// AddSimpleReading adds a simple reading to the Event
func (e *Event) AddSimpleReading(resourceName string, valueType string, value interface{}) error {
	reading, err := NewSimpleReading(e.ProfileName, e.DeviceName, resourceName, valueType, value)
	if err != nil {
		return err
	}
	e.Readings = append(e.Readings, reading)
	return nil
}

// AddBinaryReading adds a binary reading to the Event
func (e *Event) AddBinaryReading(resourceName string, binaryValue []byte, mediaType string) {
	e.Readings = append(e.Readings, NewBinaryReading(e.ProfileName, e.DeviceName, resourceName, binaryValue, mediaType))
}

// AddObjectReading adds a object reading to the Event
func (e *Event) AddObjectReading(resourceName string, objectValue interface{}) {
	e.Readings = append(e.Readings, NewObjectReading(e.ProfileName, e.DeviceName, resourceName, objectValue))
}

// AddNullReading adds a simple reading with null value to the Event
func (e *Event) AddNullReading(resourceName string, valueType string) {
	e.Readings = append(e.Readings, NewNullReading(e.ProfileName, e.DeviceName, resourceName, valueType))
}

// ToXML provides a XML representation of the Event as a string
func (e *Event) ToXML() (string, error) {
	eventXml, err := xml.Marshal(e)
	if err != nil {
		return "", err
	}

	return string(eventXml), nil
}

func (e Event) MarshalJSON() ([]byte, error) {
	data, err := e.marshal(json.Marshal)
	if err != nil || len(e.Extensions) == 0 {
		return data, err
	}
	return mergeExtensions(data, e.Extensions, json.Unmarshal, json.Marshal)
}

func (e Event) MarshalCBOR() ([]byte, error) {
	data, err := e.marshal(cbor.Marshal)
	if err != nil || len(e.Extensions) == 0 {
		return data, err
	}
	return mergeExtensions(data, e.Extensions, cbor.Unmarshal, cbor.Marshal)
}

func (e Event) marshal(marshal func(any) ([]byte, error)) ([]byte, error) {
	var aux struct {
		dtoCommon.Versionable `json:",inline"`
		Id                    string        `json:"id" validate:"required,uuid"`
		DeviceName            string        `json:"deviceName"`
		ProfileName           string        `json:"profileName"`
		SourceName            string        `json:"sourceName"`
		Origin                int64         `json:"origin"`
		Readings              []BaseReading `json:"readings"`
		Tags                  Tags          `json:"tags,omitempty"`
	}
	aux.Versionable = e.Versionable
	aux.Id = e.Id
	aux.DeviceName = e.DeviceName
	aux.ProfileName = e.ProfileName
	aux.SourceName = e.SourceName
	aux.Origin = e.Origin
	aux.Tags = e.Tags
	if len(e.Readings) > 0 {
		aux.Readings = make([]BaseReading, len(e.Readings))
	}

	if os.Getenv(common.EnvOptimizeEventPayload) == common.ValueTrue {
		for i, reading := range e.Readings {
			reading.Id = ""
			reading.DeviceName = ""
			reading.ProfileName = ""
			if e.Origin == reading.Origin {
				reading.Origin = 0
			}
			if len(e.Readings) == 1 && e.SourceName == reading.ResourceName {
				reading.ResourceName = ""
			}
			aux.Readings[i] = reading
		}
	} else {
		copy(aux.Readings, e.Readings)
	}

	return marshal(aux)
}

func (e *Event) UnmarshalJSON(b []byte) error {
	return e.unmarshal(b, jsonUnmarshalUseNumber)
}

func (e *Event) UnmarshalCBOR(b []byte) error {
	return e.unmarshal(b, cbor.Unmarshal)
}

func (e *Event) unmarshal(data []byte, unmarshal func([]byte, any) error) error {
	*e = Event{}

	var (
		rawMap map[string]any
		err    error
	)
	if err = unmarshal(data, &rawMap); err != nil {
		return err
	}
	// When cbor.Unmarshal decodes into map[string]any, the top-level keys are strings,
	// but nested map values (e.g. readings, tags) are decoded as map[any]any because
	// their target type is any. normalizeMap recursively converts these to map[string]any
	// so that subsequent type assertions work correctly for both JSON and CBOR paths.
	normalizeMap(rawMap)

	if e.ApiVersion, err = popStringValueFromKey(rawMap, keyApiVersion); err != nil {
		return err
	}
	if e.Id, err = popStringValueFromKey(rawMap, keyId); err != nil {
		return err
	}
	if e.DeviceName, err = popStringValueFromKey(rawMap, keyDeviceName); err != nil {
		return err
	}
	if e.ProfileName, err = popStringValueFromKey(rawMap, keyProfileName); err != nil {
		return err
	}
	if e.SourceName, err = popStringValueFromKey(rawMap, keySourceName); err != nil {
		return err
	}

	switch v := popKey(rawMap, keyOrigin).(type) {
	case json.Number:
		if e.Origin, err = v.Int64(); err != nil {
			return fmt.Errorf("failed to decode origin: %w", err)
		}
	case uint64: // CBOR, positive integers decode as uint64
		if v > math.MaxInt64 {
			return fmt.Errorf("origin value %d overflows int64", v)
		}
		e.Origin = int64(v)
	case int64: // CBOR, negative integers decode as int64
		e.Origin = v
	case nil:
		// key absent — leave Origin as zero
	default:
		return fmt.Errorf("origin must be a numeric type, got %T", v)
	}

	// Pop readings before convertJSONNumbers so that json.Number values inside each reading
	// (e.g. origin) are preserved for precise int64 conversion in populateFromMap.
	var rawReadings []any
	if v := popKey(rawMap, keyReadings); v != nil {
		var ok bool
		if rawReadings, ok = v.([]any); !ok {
			return fmt.Errorf("failed to decode readings: expected []any, got %T", v)
		}
	}

	// convert json.Number in rawMap to native numeric types before assigning Tags/Extensions
	convertJSONNumbers(rawMap)

	if rawTags := popKey(rawMap, keyTags); rawTags != nil {
		if tags, ok := rawTags.(map[string]any); ok {
			e.Tags = tags
		} else {
			return fmt.Errorf("failed to decode tags: expected map[string]any, got %T", rawTags)
		}
	} else {
		e.Tags = make(map[string]any)
	}

	if len(rawReadings) > 0 {
		e.Readings = make([]BaseReading, len(rawReadings))
		for i, raw := range rawReadings {
			readingMap, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("failed to decode readings[%d]: expected map[string]any, got %T", i, raw)
			}
			if err := e.Readings[i].populateFromMap(readingMap); err != nil {
				return err
			}
		}
	} else {
		e.Readings = make([]BaseReading, 0)
	}

	if os.Getenv(common.EnvOptimizeEventPayload) == common.ValueTrue {
		// recover the reduced fields
		for i, reading := range e.Readings {
			e.Readings[i].DeviceName = e.DeviceName
			e.Readings[i].ProfileName = e.ProfileName
			if reading.Origin == 0 {
				e.Readings[i].Origin = e.Origin
			}
			if len(e.Readings) == 1 && len(reading.ResourceName) == 0 {
				e.Readings[i].ResourceName = e.SourceName
			}
		}
	}

	// remaining keys are extensions
	if len(rawMap) > 0 {
		e.Extensions = rawMap
	} else {
		e.Extensions = make(map[string]any)
	}
	return nil
}
