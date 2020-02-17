/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package infrastructure

import "github.com/google/uuid"

type Identity string

// NewIdentity returns a new identity (currently uuid).
func NewIdentity() Identity {
	return StringToIdentity(NewIdentityString())
}

// NewIdentityString returns a stringified new identity.
func NewIdentityString() string {
	return uuid.New().String()
}

// StringToIdentity converts string to Identity.
func StringToIdentity(s string) Identity {
	return Identity(s)
}

// IdentityToString converts Identity to string.
func IdentityToString(identity Identity) string {
	return string(identity)
}
