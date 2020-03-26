// Code generated by go-swagger; DO NOT EDIT.

// Panther is a scalable, powerful, cloud-native SIEM written in Golang/React.
// Copyright (C) 2020 Panther Labs Inc
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// StatusCount status count
//
// swagger:model StatusCount
type StatusCount struct {

	// error
	// Minimum: 0
	Error *int64 `json:"error,omitempty"`

	// fail
	// Minimum: 0
	Fail *int64 `json:"fail,omitempty"`

	// pass
	// Minimum: 0
	Pass *int64 `json:"pass,omitempty"`
}

// Validate validates this status count
func (m *StatusCount) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateError(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateFail(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validatePass(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *StatusCount) validateError(formats strfmt.Registry) error {

	if swag.IsZero(m.Error) { // not required
		return nil
	}

	if err := validate.MinimumInt("error", "body", int64(*m.Error), 0, false); err != nil {
		return err
	}

	return nil
}

func (m *StatusCount) validateFail(formats strfmt.Registry) error {

	if swag.IsZero(m.Fail) { // not required
		return nil
	}

	if err := validate.MinimumInt("fail", "body", int64(*m.Fail), 0, false); err != nil {
		return err
	}

	return nil
}

func (m *StatusCount) validatePass(formats strfmt.Registry) error {

	if swag.IsZero(m.Pass) { // not required
		return nil
	}

	if err := validate.MinimumInt("pass", "body", int64(*m.Pass), 0, false); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *StatusCount) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *StatusCount) UnmarshalBinary(b []byte) error {
	var res StatusCount
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
