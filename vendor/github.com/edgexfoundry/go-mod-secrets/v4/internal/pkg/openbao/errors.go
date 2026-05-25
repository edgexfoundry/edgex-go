/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Corp.
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

package openbao

import (
	"fmt"
)

// ErrCaRootCert error when the provided CA Root certificate is invalid.
type ErrCaRootCert struct {
	path        string
	description string
}

func (e ErrCaRootCert) Error() string {
	return fmt.Sprintf("Unable to use the certificate '%s': %s", e.path, e.description)
}

type ErrHTTPResponse struct {
	StatusCode int
	ErrMsg     string
}

func (err ErrHTTPResponse) Error() string {
	return fmt.Sprintf("HTTP response with status code %d, message: %s", err.StatusCode, err.ErrMsg)
}
