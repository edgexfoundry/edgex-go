//
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import "encoding/xml"

type Tags map[string]any

// MarshalXML fulfills the Marshaler interface for Tags field, which is being ignored
// from XML Marshaling since maps are not supported. We have to provide our own
// marshaling of the Tags field if it is non-empty.
func (t Tags) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(t) == 0 {
		return nil
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range t {
		xmlMapEntry := struct {
			XMLName xml.Name
			Value   any `xml:",chardata"`
		}{
			XMLName: xml.Name{Local: k},
			Value:   v,
		}

		err = e.Encode(xmlMapEntry)
		if err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}
