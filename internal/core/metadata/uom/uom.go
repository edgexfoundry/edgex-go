//
// Copyright (C) 2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package uom

type UnitsOfMeasureImpl struct {
	Uom ConfigurationStruct
}

type ConfigurationStruct struct {
	Source string
	Units  map[string]Unit
}

type Unit struct {
	Source string
	Values []string
}

func (u *UnitsOfMeasureImpl) Validate(unit string) bool {
	if unit == "" || len(u.Uom.Units) == 0 {
		return true
	}

	for _, units := range u.Uom.Units {
		for _, v := range units.Values {
			if unit == v {
				return true
			}
		}
	}

	return false
}
