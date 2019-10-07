/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package errorconcept

import "github.com/edgexfoundry/go-mod-core-contracts/clients/types"

// NewServiceClientHttpError represents the accessor for the service-client-specific error concepts
func NewServiceClientHttpError(err error) *serviceClientHttpError {
	return &serviceClientHttpError{Err: err}
}

type serviceClientHttpError struct {
	Err error
}

func (r serviceClientHttpError) httpErrorCode() int {
	return r.Err.(types.ErrServiceClient).StatusCode
}

func (r serviceClientHttpError) isA(err error) bool {
	_, ok := err.(types.ErrServiceClient)
	return ok
}

func (r serviceClientHttpError) message(err error) string {
	return err.Error()
}
