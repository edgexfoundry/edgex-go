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

// Interval
type ErrIntervalNotFound struct{
	id string
}

func (e ErrIntervalNotFound) Error() string{
	return fmt.Sprintf("no interval found for id: %s",e.id)
}

func NewErrIntervalNotFound(id string) error {
	return &ErrIntervalNotFound{id: id}
}

type ErrIntervalNameInUse struct {
	name string
}

func (e ErrIntervalNameInUse) Error() string {
	return fmt.Sprintf("interval name: %s in use",e.name)
}

func NewErrIntervalNameInUse(name string) error{
	return &ErrIntervalNameInUse{name: name}
}

type ErrIntervalStillUsedByIntervalActions struct {
	name string
}
func (e ErrIntervalStillUsedByIntervalActions) Error() string {
	return fmt.Sprintf("interval still in use by intervalAction(s) name:  %s",e.name)
}

func NewErrIntervalStillInUse(name string) error{
	return &ErrIntervalStillUsedByIntervalActions{name: name}
}

// IntervalAction
type ErrIntervalActionNotFound struct {
	id string
}

func (e ErrIntervalActionNotFound) Error() string{
	return fmt.Sprintf("no intervalAction found with id: %s",e.id)
}

func NewErrIntervalActionNotFound(id string) error {
	return &ErrIntervalActionNotFound{id: id}
}

type ErrIntervalActionTargetNameRequired struct {
	id string
}

func (e ErrIntervalActionTargetNameRequired) Error() string{
	return fmt.Sprintf("intervalAction [ %s ] requires a target none provided. ",e.id)
}

func NewErrIntervalActionTargetNameRequired(id string) error {
	return &ErrIntervalActionTargetNameRequired{id: id}
}


type ErrIntervalActionNameInUse struct {
	name string
}

func (e ErrIntervalActionNameInUse) Error() string{
	return fmt.Sprintf("intervalAction name: %s in use",e.name)
}

func NewErrIntervalActionNameInUse(name string) error {
	return &ErrIntervalActionNameInUse{name: name}
}


type ErrInvalidTimeFormat struct {
	 value string
}

func (e ErrInvalidTimeFormat) Error() string {
	return fmt.Sprintf("invalid time format for value: %s",e.value)
}

func NewErrInvalidTimeFormat(value string) error {
	return &ErrInvalidTimeFormat{value: value}
}

type ErrInvalidFrequencyFormat struct {
	 frequency string
}

func (e ErrInvalidFrequencyFormat) Error() string {
	return fmt.Sprintf("invalid frequency format for value: %s", e.frequency)
}

func NewErrInvalidFrequencyFormat(frequency string) error{
	return &ErrInvalidFrequencyFormat{frequency: frequency}
}

type ErrInvalidCronFormat struct {
	cron string
}

func (e ErrInvalidCronFormat) Error() string {
	return fmt.Sprintf("invalid cron format for value: %s", e.cron)
}

func NewErrInvalidCronFormat(cron string) error {
	return &ErrInvalidCronFormat{cron: cron}
}

type ErrDbNotFound struct {
}

func (e ErrDbNotFound) Error() string {
	return db.ErrNotFound.Error()
}

func NewErrDbNotFound() error {
	return &ErrDbNotFound{}
}