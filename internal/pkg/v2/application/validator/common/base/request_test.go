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

package base

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dtoError "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"

	"github.com/stretchr/testify/assert"
)

// TestValidate tests the Validator for BaseRequest properties.
func TestValidate(t *testing.T) {
	requestID := test.FactoryRandomString()
	variations := []struct {
		name             string
		validate         func() (interface{}, infrastructure.Status)
		expectedResponse interface{}
		expectedStatus   infrastructure.Status
	}{
		{
			name: "valid",
			validate: func() (interface{}, infrastructure.Status) {
				return Validate(requestID, test.FactoryRandomString())
			},
			expectedResponse: nil,
			expectedStatus:   infrastructure.StatusSuccess,
		},
		{
			name: "empty requestID",
			validate: func() (interface{}, infrastructure.Status) {
				return Validate(requestID, "")
			},
			expectedResponse: dtoError.NewResponse(
				dtoBase.NewResponse("", requestID, application.StatusRequestIdEmptyFailure),
			),
			expectedStatus: application.StatusRequestIdEmptyFailure,
		},
	}
	for variationIndex := range variations {
		t.Run(
			variations[variationIndex].name,
			func(t *testing.T) {
				result, status := variations[variationIndex].validate()

				assert.Equal(t, variations[variationIndex].expectedResponse, result)
				assert.Equal(t, variations[variationIndex].expectedStatus, status)
			},
		)
	}
}
