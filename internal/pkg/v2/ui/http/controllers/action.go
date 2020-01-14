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

package controllers

import "net/http"

const (
	// ActionCreate defines the http transport's method for a create action.
	ActionCreate = http.MethodPut

	// ActionRead defines the http transport's method for a read action.
	ActionRead = http.MethodGet

	// ActionUpdate defines the http transport's method for a update action.
	ActionUpdate = http.MethodPatch

	// ActionDelete defines the http transport's method for a delete action.
	ActionDelete = http.MethodDelete

	// ActionCommand defines the http transport's method for a command (e.g. "do this") action.
	ActionCommand = http.MethodPost
)
