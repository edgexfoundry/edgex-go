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

package v2dot0

import (
	dtoBaseV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/common/base"
	dtoV2dot0 "github.com/edgexfoundry/edgex-go/internal/pkg/v2/application/dto/v2dot0/core/metadata/addressable/create"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
)

// FactoryValidCreateRequest returns a valid addressable add request.
func FactoryValidCreateRequest(requestID string) *dtoV2dot0.Request {
	return dtoV2dot0.NewRequest(
		dtoBaseV2dot0.NewRequest(requestID),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
		test.FactoryRandomString(),
	)
}
