// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright NetFoundry Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// __          __              _
// \ \        / /             (_)
//  \ \  /\  / /_ _ _ __ _ __  _ _ __   __ _
//   \ \/  \/ / _` | '__| '_ \| | '_ \ / _` |
//    \  /\  / (_| | |  | | | | | | | | (_| | : This file is generated, do not edit it.
//     \/  \/ \__,_|_|  |_| |_|_|_| |_|\__, |
//                                      __/ |
//                                     |___/

package rest_model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// AuthenticatorUpdateWithCurrent All of the fields on an authenticator that will be updated
//
// swagger:model authenticatorUpdateWithCurrent
type AuthenticatorUpdateWithCurrent struct {
	AuthenticatorUpdate

	// current password
	// Required: true
	CurrentPassword *Password `json:"currentPassword"`
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (m *AuthenticatorUpdateWithCurrent) UnmarshalJSON(raw []byte) error {
	// AO0
	var aO0 AuthenticatorUpdate
	if err := swag.ReadJSON(raw, &aO0); err != nil {
		return err
	}
	m.AuthenticatorUpdate = aO0

	// AO1
	var dataAO1 struct {
		CurrentPassword *Password `json:"currentPassword"`
	}
	if err := swag.ReadJSON(raw, &dataAO1); err != nil {
		return err
	}

	m.CurrentPassword = dataAO1.CurrentPassword

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (m AuthenticatorUpdateWithCurrent) MarshalJSON() ([]byte, error) {
	_parts := make([][]byte, 0, 2)

	aO0, err := swag.WriteJSON(m.AuthenticatorUpdate)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO0)
	var dataAO1 struct {
		CurrentPassword *Password `json:"currentPassword"`
	}

	dataAO1.CurrentPassword = m.CurrentPassword

	jsonDataAO1, errAO1 := swag.WriteJSON(dataAO1)
	if errAO1 != nil {
		return nil, errAO1
	}
	_parts = append(_parts, jsonDataAO1)
	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this authenticator update with current
func (m *AuthenticatorUpdateWithCurrent) Validate(formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with AuthenticatorUpdate
	if err := m.AuthenticatorUpdate.Validate(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCurrentPassword(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *AuthenticatorUpdateWithCurrent) validateCurrentPassword(formats strfmt.Registry) error {

	if err := validate.Required("currentPassword", "body", m.CurrentPassword); err != nil {
		return err
	}

	if err := validate.Required("currentPassword", "body", m.CurrentPassword); err != nil {
		return err
	}

	if m.CurrentPassword != nil {
		if err := m.CurrentPassword.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("currentPassword")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("currentPassword")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this authenticator update with current based on the context it is used
func (m *AuthenticatorUpdateWithCurrent) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with AuthenticatorUpdate
	if err := m.AuthenticatorUpdate.ContextValidate(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateCurrentPassword(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *AuthenticatorUpdateWithCurrent) contextValidateCurrentPassword(ctx context.Context, formats strfmt.Registry) error {

	if m.CurrentPassword != nil {
		if err := m.CurrentPassword.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("currentPassword")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("currentPassword")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *AuthenticatorUpdateWithCurrent) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *AuthenticatorUpdateWithCurrent) UnmarshalBinary(b []byte) error {
	var res AuthenticatorUpdateWithCurrent
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
