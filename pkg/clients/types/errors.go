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
package types

import "fmt"

type ErrNotFound struct{}

func (e ErrNotFound) Error() string {
	return "Item not found"
}

type ErrResponseNil struct{}

func (e ErrResponseNil) Error() string {
	return "Response was nil"
}

type ErrServiceClient struct {
	StatusCode int
	bodyBytes  []byte
	errMsg     string
}

func NewErrServiceClient(statusCode int, body []byte) error {
	e := &ErrServiceClient{StatusCode: statusCode, bodyBytes: body}
	return e
}

func (e ErrServiceClient) Error() string {
	return fmt.Sprintf("%d - %s", e.StatusCode, e.bodyBytes)
}
