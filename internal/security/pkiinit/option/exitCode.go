//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//
package option

// ExitCode is the code used for exit status
type exitCode int

const (
	// normal exit code
	normal exitCode = 0
	// noOptionSelected exit code
	noOptionSelected exitCode = 1
	// exitWithError is exit code for error
	exitWithError exitCode = 2
)

func (code exitCode) intValue() int {
	return int(code)
}
