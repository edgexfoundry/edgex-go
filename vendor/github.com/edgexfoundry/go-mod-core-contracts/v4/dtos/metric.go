/*******************************************************************************
 * Copyright 2022 Intel Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package dtos

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
)

// Metric defines the metric data for a specific named metric
type Metric struct {
	common.Versionable `json:",inline"`
	// Name is the identifier of the collected Metric
	Name string `json:"name" validate:"edgex-dto-none-empty-string"`
	// Fields are the Key/Value measurements associated with the metric
	Fields []MetricField `json:"fields,omitempty" validate:"required"`
	// Tags and the Key/Value tags associated with the metric
	Tags []MetricTag `json:"tags,omitempty"`
	// Timestamp is the time and date the metric was collected.
	Timestamp int64 `json:"timestamp" validate:"required"`
}

// MetricField defines a metric field associated with a metric
type MetricField struct {
	// Name is the identifier of the metric field
	Name string `json:"name" validate:"edgex-dto-none-empty-string"`
	// Value is measurement for the metric field
	Value interface{} `json:"value" validate:"required"`
}

// MetricTag defines a metric tag associated with a metric
type MetricTag struct {
	// Name is the identifier of the metric tag
	Name string `json:"name" validate:"edgex-dto-none-empty-string"`
	// Value is tag vale for the metric tag
	Value string `json:"value" validate:"required"`
}

// NewMetric creates a new metric for the specified data
func NewMetric(name string, fields []MetricField, tags []MetricTag) (Metric, error) {
	if err := ValidateMetricName(name, "metric"); err != nil {
		return Metric{}, err
	}

	if len(fields) == 0 {
		return Metric{}, errors.New("one or more metric fields are required")
	}

	for _, field := range fields {
		if err := ValidateMetricName(field.Name, "field"); err != nil {
			return Metric{}, err
		}
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			if err := ValidateMetricName(tag.Name, "tag"); err != nil {
				return Metric{}, err
			}
		}
	}

	metric := Metric{
		Versionable: common.NewVersionable(),
		Name:        name,
		Fields:      fields,
		Timestamp:   time.Now().UnixNano(),
		Tags:        tags,
	}

	return metric, nil
}

func ValidateMetricName(name string, nameType string) error {
	if len(strings.TrimSpace(name)) == 0 {
		return fmt.Errorf("%s name can not be empty or blank", nameType)
	}

	// TODO: Use regex to validate Name characters
	return nil
}

// ToLineProtocol transforms the Metric to Line Protocol syntax which is most commonly used with InfluxDB
// For more information on Line Protocol see: https://docs.influxdata.com/influxdb/v2.0/reference/syntax/line-protocol/
// Line Protocol Syntax:
//
//	<measurement>[,<tag_key>=<tag_value>[,<tag_key>=<tag_value>]] <field_key>=<field_value>[,<field_key>=<field_value>] [<timestamp>]
//
// Examples:
//
//	measurementName fieldKey="field string value" 1556813561098000000
//	myMeasurement,tag1=value1,tag2=value2 fieldKey="fieldValue" 1556813561098000000
//
// Note that this is a simple helper function for those receiving this DTO that are pushing metrics to an endpoint
// that receives LineProtocol such as InfluxDb or Telegraf
func (m *Metric) ToLineProtocol() string {
	var fields strings.Builder
	isFirst := true
	for _, field := range m.Fields {
		// Fields section doesn't have a leading comma per syntax above so need to skip the comma on the first field
		if isFirst {
			isFirst = false
		} else {
			fields.WriteString(",")
		}
		fields.WriteString(field.Name + "=" + formatLineProtocolValue(field.Value))
	}

	// Tags section does have a leading comma per syntax above
	var tags strings.Builder
	for _, tag := range m.Tags {
		tags.WriteString("," + tag.Name + "=" + tag.Value)
	}

	result := fmt.Sprintf("%s%s %s %d", m.Name, tags.String(), fields.String(), m.Timestamp)

	return result
}

func formatLineProtocolValue(value interface{}) string {
	switch value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", value)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%di", value)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%du", value)
	default:
		return fmt.Sprintf("%v", value)
	}
}
