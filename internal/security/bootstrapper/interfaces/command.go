/*******************************************************************************
 * Copyright 2021 Intel Corporation
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

package interfaces

const (
	// StatusCodeExitNormal exit code for normal case
	StatusCodeExitNormal = 0
	// StatusCodeNoOptionSelected exit code for missing options case
	StatusCodeNoOptionSelected = 1
	// StatusCodeExitWithError is exit code for error case
	StatusCodeExitWithError = 2
	// JSONContentType is the content type for JSON based body/payload
	JSONContentType = "application/json"
)

// Command implement the Command pattern
type Command interface {
	Execute() (statusCode int, err error)
	GetCommandName() string
}
