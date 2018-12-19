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
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

type ErrEventNotFound struct {
	id string
}

func (e ErrEventNotFound) Error() string {
	return fmt.Sprintf("no event found for id %s", e.id)
}

func NewErrEventNotFound(id string) error {
	return &ErrEventNotFound{id: id}
}

type ErrValueDescriptorInvalid struct {
	name string
	err  error
}

func (b ErrValueDescriptorInvalid) Error() string {
	return fmt.Sprintf("invalid value descriptor '%s': %v", b.name, b.err)
}

func NewErrValueDescriptorInvalid(name string, err error) error {
	return &ErrValueDescriptorInvalid{name: name, err: err}
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

type ErrLimitExceeded struct {
	limit int
}

func (e ErrLimitExceeded) Error() string {
	return fmt.Sprintf("limit %d exceeds configured max", e.limit)
}

func NewErrLimitExceeded(limit int) error {
	return &ErrLimitExceeded{limit: limit}
}

type ErrJsonDecoding struct {
	name string
}

func (e ErrJsonDecoding) Error() string {
	return fmt.Sprintf("error decoding the reading: %s", e.name)
}

func NewErrJsonDecoding(name string) error {
	return &ErrJsonDecoding{name: name}
}

type ErrDbNotFound struct {
}

func (e ErrDbNotFound) Error() string {
	return db.ErrNotFound.Error()
}

func NewErrDbNotFound() error {
	return &ErrDbNotFound{}
}

type ErrInvalidId struct {
	id string
}

func (e ErrInvalidId) Error() string {
	return fmt.Sprintf("invalid ID: %s", e.id)
}

func NewErrInvalidId(id string) error {
	return &ErrInvalidId{id: id}
}
