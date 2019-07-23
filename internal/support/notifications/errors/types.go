/*******************************************************************************
 * Copyright 2019 VMware Inc.
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
 *
 *******************************************************************************/

package errors

import (
	"fmt"
)

type ErrNotificationInUse struct {
	name string
}

func (e ErrNotificationInUse) Error() string {
	return fmt.Sprintf("Notification '%s' is still in use", e.name)
}

func NewErrNotificationInUse(name string) error {
	return ErrNotificationInUse{name: name}
}

type ErrDistribution struct {
	slug string
}

func (e ErrDistribution) Error() string {
	return fmt.Sprintf("Could not distribute notification '%s'", e.slug)
}

func NewErrDistribution(slug string) error {
	return ErrDistribution{slug: slug}
}

type ErrUpdate struct {
	slug string
}

func (e ErrUpdate) Error() string {
	return fmt.Sprintf("Could not update notification '%s'", e.slug)
}

func NewErrUpdate(slug string) error {
	return ErrUpdate{slug: slug}
}

type ErrNotificationNotFound struct {
	slug string
}

func (e ErrNotificationNotFound) Error() string {
	return fmt.Sprintf("Notification '%s' not found", e.slug)
}

func NewErrNotificationNotFound(slug string) error {
	return ErrNotificationNotFound{slug: slug}
}

type ErrNotificationsNotFound struct {
}

func (e ErrNotificationsNotFound) Error() string {
	return fmt.Sprintf("No notifications found")
}

func NewErrNotificationsNotFound() error {
	return ErrNotificationsNotFound{}
}

type ErrSubscriptionNotFound struct {
	slug string
}

func (e ErrSubscriptionNotFound) Error() string {
	return fmt.Sprintf("Subscription '%s' not found", e.slug)
}

func NewErrSubscriptionNotFound(slug string) error {
	return ErrSubscriptionNotFound{slug: slug}
}

type ErrSubscriptionInUse struct {
	name string
}

func (e ErrSubscriptionInUse) Error() string {
	return fmt.Sprintf("Notification '%s' is still in use", e.name)
}

func NewErrSubscriptionInUse(name string) error {
	return ErrSubscriptionInUse{name: name}
}
