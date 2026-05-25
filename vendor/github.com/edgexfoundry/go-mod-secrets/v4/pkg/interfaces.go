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

package pkg

import "net/http"

// Caller interface used to abstract the implementation details for issuing an HTTP request. This allows for easier testing by the way of mocks.
type Caller interface {
	Do(req *http.Request) (*http.Response, error)
}

// TokenExpiredCallback is the callback function to handle the case when the secret store token has already expired
type TokenExpiredCallback func(expiredToken string) (replacementToken string, retry bool)
