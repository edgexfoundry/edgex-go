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

package waitfor

import "fmt"

type uriFlagsVar []string

// String overrides the Flag.Value interface, method String() string
func (uris *uriFlagsVar) String() string {
	return fmt.Sprint(*uris)
}

// Set overrides the Flag.Value interface, method Set(string) error
// uriFlagsVar is a slice of string flags and thus we aggregate it over each flag call
func (uris *uriFlagsVar) Set(value string) error {
	*uris = append(*uris, value)
	return nil
}
