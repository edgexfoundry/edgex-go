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

package config

// Notification Info provides properties related to the assembly of notification content
// Deprecated: this struct should be moved to go-mod-core-contracts
type NotificationInfo struct {
	Content           string
	Description       string
	Label             string
	PostDeviceChanges bool
	Sender            string
	Slug              string
}
