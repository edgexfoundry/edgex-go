//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

import "maps"

type ResourceProperties struct {
	ValueType    string
	ReadWrite    string
	Units        string
	Minimum      *float64
	Maximum      *float64
	DefaultValue string
	Mask         *uint64
	Shift        *int64
	Scale        *float64
	Offset       *float64
	Base         *float64
	Assertion    string
	MediaType    string
	Optional     map[string]any
}

func (rp ResourceProperties) Clone() ResourceProperties {
	var minimum *float64
	if rp.Minimum != nil {
		val := *rp.Minimum
		minimum = &val
	}
	var maximum *float64
	if rp.Maximum != nil {
		val := *rp.Maximum
		maximum = &val
	}
	var mask *uint64
	if rp.Mask != nil {
		val := *rp.Mask
		mask = &val
	}
	var shift *int64
	if rp.Shift != nil {
		val := *rp.Shift
		shift = &val
	}
	var scale *float64
	if rp.Scale != nil {
		val := *rp.Scale
		scale = &val
	}
	var offset *float64
	if rp.Offset != nil {
		val := *rp.Offset
		offset = &val
	}
	var base *float64
	if rp.Base != nil {
		val := *rp.Base
		base = &val
	}
	cloned := ResourceProperties{
		ValueType:    rp.ValueType,
		ReadWrite:    rp.ReadWrite,
		Units:        rp.Units,
		Minimum:      minimum,
		Maximum:      maximum,
		DefaultValue: rp.DefaultValue,
		Mask:         mask,
		Shift:        shift,
		Scale:        scale,
		Offset:       offset,
		Base:         base,
		Assertion:    rp.Assertion,
		MediaType:    rp.MediaType,
	}
	if len(rp.Optional) > 0 {
		rp.Optional = make(map[string]any)
		maps.Copy(rp.Optional, rp.Optional)
	}
	return cloned
}
