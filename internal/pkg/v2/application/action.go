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

package application

const (
	// ActionCreate defines the transport-agnostic representation for a create action.
	ActionCreate = "create"

	// ActionRead defines the transport-agnostic representation for a read action.
	ActionRead = "read"

	// ActionUpdate defines the transport-agnostic representation for a update action.
	ActionUpdate = "update"

	// ActionDelete defines the transport-agnostic representation for a delete action.
	ActionDelete = "delete"

	// ActionCommand defines the transport-agnostic representation for a command (e.g. "do this") action.
	ActionCommand = "command"
)
