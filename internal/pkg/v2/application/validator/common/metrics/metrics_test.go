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

package metrics

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/delegate"
	dtoBase "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dtoError "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/error"
	dto "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/metrics"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"

	"github.com/stretchr/testify/assert"
)

// TestValidate tests the Validator for Version request.
func TestValidate(t *testing.T) {
	readingID := test.FactoryRandomString()
	invalidRequestType := "string is not the request type we're expecting."

	variations := []struct {
		name             string
		request          interface{}
		expectedResponse interface{}
		expectedStatus   infrastructure.Status
	}{
		{
			name:             "valid",
			request:          dto.NewRequest(dtoBase.NewRequest(readingID)),
			expectedResponse: dto.NewRequest(dtoBase.NewRequest(readingID)),
			expectedStatus:   infrastructure.StatusSuccess,
		},
		{
			name:    "invalid type",
			request: invalidRequestType,
			expectedResponse: dtoError.NewResponse(
				dtoBase.NewResponse("", invalidRequestType, application.StatusTypeAssertionFailure),
			),
			expectedStatus: application.StatusTypeAssertionFailure,
		},
		{
			name:    "empty requestID",
			request: dto.NewRequest(dtoBase.NewRequest("")),
			expectedResponse: dtoError.NewResponse(
				dtoBase.NewResponse(
					"",
					dto.NewRequest(dtoBase.NewRequest("")),
					application.StatusRequestIdEmptyFailure,
				),
			),
			expectedStatus: application.StatusRequestIdEmptyFailure,
		},
	}
	for variationIndex := range variations {
		t.Run(
			variations[variationIndex].name,
			func(t *testing.T) {
				result, status := Validate(
					variations[variationIndex].request,
					application.NewBehavior(application.Version2, application.KindMetrics, application.ActionCommand),
					delegate.TestExecutable,
				)

				assert.Equal(t, variations[variationIndex].expectedResponse, result)
				assert.Equal(t, variations[variationIndex].expectedStatus, status)
			},
		)
	}
}
