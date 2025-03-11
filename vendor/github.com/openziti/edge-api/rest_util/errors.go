/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package rest_util

import (
	"fmt"
	"github.com/openziti/edge-api/rest_model"
)

// WrapErr return an error that has been wrapped so that err.Error() prints useful API error information if possible.
// If the error does not support deep API error information, the error is returned as is.
func WrapErr(err error) error {
	if errWithPayload, ok := err.(ApiErrorPayload); ok {
		apiErrEnv := errWithPayload.GetPayload()

		if apiErrEnv != nil && apiErrEnv.Error != nil {
			return &APIFormattedError{
				source:   err,
				APIError: apiErrEnv.Error,
			}
		}
	}

	return err
}

// APIFormattedError takes a rest_model.APIError and wraps it so that it can output
// helpful information rather than pointer addresses for `Data` and `Meta`
type APIFormattedError struct {
	source error
	*rest_model.APIError
}

func (e *APIFormattedError) Unwrap() error {
	return e.source
}

func (e *APIFormattedError) Error() string {
	causeStr := ""

	if e.Cause != nil {
		if e.Cause.APIError.Code != "" {
			cause := APIFormattedError{APIError: &e.Cause.APIError}
			causeStr = cause.Error()
		} else if e.Cause.APIFieldError.Field != "" {
			causeStr = fmt.Sprintf("error in field %s with value %v: %s", e.Cause.APIFieldError.Field, e.Cause.APIFieldError.Value, e.Cause.APIFieldError.Reason)
		}
	}
	result := fmt.Sprintf("error for request %s: %s: %s", e.RequestID, e.Code, e.Message)

	if causeStr != "" {
		result = result + ", caused by: " + causeStr
	}

	return result
}
