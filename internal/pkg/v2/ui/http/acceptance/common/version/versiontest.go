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

package version

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/acceptance/common/version/v2dot0"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common/batch"
	controller "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common/version"

	"github.com/gorilla/mux"
)

// UseCaseTest verifies version endpoint returns expected result; common implementation intended to be executed by
// each service that includes version support.
func UseCaseTest(t *testing.T, router *mux.Router) {
	test.UseCaseAcceptance(
		t,
		router,
		controller.Method,
		map[string][]*test.Case{
			application.Version2dot0: v2dot0.UseCaseTestCases(t),
		},
	)
}

// BatchTest verifies version requests sent to batch endpoint return expected results; common implementation
// intended to be executed by each service that includes version support.
func BatchTest(t *testing.T, router *mux.Router) {
	test.BatchAcceptance(
		t,
		router,
		batch.Method,
		map[string][]*test.Case{
			application.Version2dot0: v2dot0.BatchTestCases(t, controller.Kind, controller.Action),
		},
	)
}
