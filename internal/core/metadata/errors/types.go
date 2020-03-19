/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

package errors

import (
	"fmt"
)

type ErrLimitExceeded struct {
	limit int
}

func (e ErrLimitExceeded) Error() string {
	return fmt.Sprintf("result count exceeds configured max %d", e.limit)
}

func NewErrLimitExceeded(limit int) error {
	return ErrLimitExceeded{limit: limit}
}

type ErrDuplicateName struct {
	msg string
}

func (e ErrDuplicateName) Error() string {
	return e.msg
}

func NewErrDuplicateName(message string) error {
	return ErrDuplicateName{msg: message}
}

type ErrEmptyAddressableName struct {
}

func (e ErrEmptyAddressableName) Error() string {
	return "name is required for addressable"
}

func NewErrEmptyAddressableName() error {
	return ErrEmptyAddressableName{}
}

type ErrAddressableNotFound struct {
	id   string
	name string
}

func (e ErrAddressableNotFound) Error() string {
	if e.id != "" {
		return fmt.Sprintf("addressable with id '%s' not found", e.id)
	} else if e.name != "" {
		return fmt.Sprintf("addressable with name '%s' not found", e.name)
	}
	return "addressable not found"
}

func NewErrAddressableNotFound(id string, name string) error {
	return ErrAddressableNotFound{id: id}
}

type ErrAddressableInUse struct {
	name string
}

func (e ErrAddressableInUse) Error() string {
	return fmt.Sprintf("addressable '%s' still in use", e.name)
}

func NewErrAddressableInUse(name string) error {
	return ErrAddressableInUse{name: name}
}

type ErrBadRequest struct {
	value string
}

func (e ErrBadRequest) Error() string {
	return fmt.Sprintf("received value %v is invalid", e.value)
}

func NewErrBadRequest(invalid string) error {
	return ErrBadRequest{value: invalid}
}

type ErrItemNotFound struct {
	key string
}

func (e ErrItemNotFound) Error() string {
	return fmt.Sprintf("no item found for supplied key: %s", e.key)
}

func NewErrItemNotFound(key string) error {
	return ErrItemNotFound{key: key}
}

type ErrDeviceProfileNotFound struct {
	name string
	id   string
}

func (e ErrDeviceProfileNotFound) Error() string {
	return fmt.Sprintf("device profile not found -- id: '%s' name: '%s'", e.id, e.name)
}

func NewErrDeviceProfileNotFound(id string, name string) error {
	return ErrDeviceProfileNotFound{
		id:   id,
		name: name,
	}
}

type ErrDeviceProfileInvalidState struct {
	id          string
	name        string
	description string
}

func (e ErrDeviceProfileInvalidState) Error() string {
	return fmt.Sprintf("device profile invalid state -- id: '%s' name: '%s' description: '%s'", e.id, e.name, e.description)
}

func NewErrDeviceProfileInvalidState(id string, name string, description string) error {
	return ErrDeviceProfileInvalidState{
		id:          id,
		name:        name,
		description: description,
	}
}

type ErrEmptyDeviceProfileName struct {
}

func (e ErrEmptyDeviceProfileName) Error() string {
	return fmt.Sprint("Device profile name cannot be empty")
}

func NewErrEmptyDeviceProfileName() error {
	return ErrEmptyDeviceProfileName{}
}

type ErrEmptyFile struct {
	fileType string
}

func (e ErrEmptyFile) Error() string {
	return fmt.Sprintf("%s file is empty", e.fileType)
}

func NewErrEmptyFile(fileType string) ErrEmptyFile {
	return ErrEmptyFile{
		fileType: fileType,
	}
}

type ErrNameCollision struct {
	name   string
	fromID string
	toID   string
}

func (e ErrNameCollision) Error() string {
	return fmt.Sprintf("%s is used by two provision watchers: %s and %s", e.name, e.fromID, e.toID)
}

func NewErrNameCollision(name, fromID, toID string) ErrNameCollision {
	return ErrNameCollision{
		name:   name,
		fromID: fromID,
		toID:   toID,
	}
}
