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

import "fmt"

type ErrValueDescriptorInvalid struct {
	name string
}

func (b ErrValueDescriptorInvalid) Error() string {
	return fmt.Sprintf("value descriptor '%s' is either not unique or not formatted properly", b.name)
}

func NewErrValueDescriptorInvalid(name string) error {
	return &ErrValueDescriptorInvalid{name: name}
}

type ErrValueDescriptorNotFound struct {
	id string
}

func (e ErrValueDescriptorNotFound) Error() string {
	return fmt.Sprintf("no value descriptor for reading '%s'", e.id)
}

func NewErrValueDescriptorNotFound(id string) error {
	return &ErrValueDescriptorNotFound{id: id}
}

type ErrUnsupportedDatabase struct {
	dbType string
}

func (e ErrUnsupportedDatabase) Error() string {
	return fmt.Sprintf("database type '%s' not supported", e.dbType)
}

func NewErrUnsupportedDatabase(dbType string) error {
	return &ErrUnsupportedDatabase{dbType: dbType}
}

type ErrUnsupportedPublisher struct {
	pubType string
}

func (e ErrUnsupportedPublisher) Error() string {
	return fmt.Sprintf("publisher type '%s' not supported", e.pubType)
}

func NewErrUnsupportedPublisher(pubType string) error {
	return &ErrUnsupportedPublisher{pubType: pubType}
}

type ErrValueDescriptorInUse struct {
	name string
}

func (e ErrValueDescriptorInUse) Error() string {
	return fmt.Sprintf("value descriptor '%s' still referenced by readings", e.name)
}

func NewErrValueDescriptorInUse(name string) error {
	return &ErrValueDescriptorInUse{name: name}
}
