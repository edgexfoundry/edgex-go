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
 *******************************************************************************/

package errors

import "fmt"

type ErrNotificationNotFound struct {
	slug string
}

func (e ErrNotificationNotFound) Error() string {
	return fmt.Sprintf("Notification '%s' not found", e.slug)
}

func NewErrNotificationNotFound(slug string) error {
	return ErrNotificationNotFound{slug: slug}
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

type ErrInvalidEmailAddresses struct {
	description string
}

func (e ErrInvalidEmailAddresses) Error() string {
	return fmt.Sprintf("%s", e.description)
}

func NewErrInvalidEmailAddresses(description string) error {
	return ErrInvalidEmailAddresses{description: description}
}
