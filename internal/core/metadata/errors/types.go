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
	return fmt.Sprintf("limit %d exceeds configured max", e.limit)
}

func NewErrLimitExceeded(limit int) error {
	return &ErrLimitExceeded{limit: limit}
}

type ErrDuplicateAddressableName struct {
	name string
}

func (e ErrDuplicateAddressableName) Error() string {
	return fmt.Sprintf("duplicate name for addressable: %s", e.name)
}

func NewErrDuplicateAddressableName(name string) error {
	return &ErrDuplicateAddressableName{name: name}
}

type ErrEmptyAddressableName struct {
}

func (e ErrEmptyAddressableName) Error() string {
	return "name is required for addressable"
}

func NewErrEmptyAddressableName() error {
	return &ErrEmptyAddressableName{}
}

type ErrAddressableNotFound struct {
	id string
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
	return &ErrAddressableNotFound{id: id}
}

type ErrAddressableInUse struct {
	name string
}

func (e ErrAddressableInUse) Error() string {
	return fmt.Sprintf("addressable '%s' still in use", e.name)
}

func NewErrAddressableInUse(name string) error {
	return &ErrAddressableInUse{name: name}
}