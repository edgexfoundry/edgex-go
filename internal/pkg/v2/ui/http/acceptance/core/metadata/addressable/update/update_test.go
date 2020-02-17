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

package add

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure/test"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/acceptance/core/metadata"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/acceptance/core/metadata/addressable/update/v2dot0"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/common/batch"
	controller "github.com/edgexfoundry/edgex-go/internal/pkg/v2/ui/http/controllers/core/metadata/addressable/update"
)

// TestUseCaseAcceptanceUpdate verifies addressable update use-case endpoint returns expected result.
func TestUseCaseAcceptanceUpdate(t *testing.T) {
	cancel, wg, router := metadata.NewSUT(t, []string{test.EnvNoSecurity}, []string{})
	test.UseCaseAcceptance(
		t,
		router,
		controller.Method,
		map[string][]*test.Case{
			application.Version2dot0: v2dot0.UseCaseTestCases(t),
		},
	)
	cancel()
	wg.Wait()
}

// TestBatchAcceptanceUpdate verifies addressable update via batch endpoint returns expected result.
func TestBatchAcceptanceUpdate(t *testing.T) {
	cancel, wg, router := metadata.NewSUT(t, []string{test.EnvNoSecurity}, []string{})
	test.BatchAcceptance(
		t,
		router,
		batch.Method,
		map[string][]*test.Case{
			application.Version2dot0: v2dot0.BatchTestCases(t, controller.Kind, controller.Action),
		},
	)
	cancel()
	wg.Wait()
}
