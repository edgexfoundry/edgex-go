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

package delegate

import (
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"
)

// testHandler contains references to dependencies required by the Handler implementation.
type testHandler struct {
	before string
	after  string
}

// newTestHandler is a factory function that returns a testHandler.
func newTestHandler(before, after string) *testHandler {
	return &testHandler{
		before: before,
		after:  after,
	}
}

// Handler implements middleware.Handler contract; ignores request, assumes response is a string, and executes wrapped
// application.Executable's response with supplied before and after text.
func (h *testHandler) Handler(
	request interface{},
	behavior *application.Behavior,
	execute application.Executable) (response interface{}, status infrastructure.Status) {

	r, _ := execute(request)
	responseString, _ := r.(string)
	return h.before + concat(behavior.Version, behavior.Kind, behavior.Action) + responseString + h.after,
		infrastructure.StatusSuccess
}

// concat concatenates version, kind, and action into a single string.
func concat(version, kind, action string) string {
	return fmt.Sprintf("(%s,%s,%s)", version, kind, action)
}
