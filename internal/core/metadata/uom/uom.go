//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package uom

type UnitsOfMeasureImpl struct {
	Source string          `json:"source,omitempty" yaml:"Source,omitempty"`
	Units  map[string]Unit `json:"units,omitempty" yaml:"Units,omitempty"`
}

type Unit struct {
	Source string   `json:"source,omitempty" yaml:"Source,omitempty"`
	Values []string `json:"values,omitempty" yaml:"Values,omitempty"`
}

func (u *UnitsOfMeasureImpl) Validate(unit string) bool {
	if unit == "" || len(u.Units) == 0 {
		return true
	}

	for _, units := range u.Units {
		for _, v := range units.Values {
			if unit == v {
				return true
			}
		}
	}

	return false
}
