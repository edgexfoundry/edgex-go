//
// Copyright (c) 2022 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
// in compliance with the License. You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software distributed under the License
// is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
// or implied. See the License for the specific language governing permissions and limitations under
// the License.
//
// SPDX-License-Identifier: Apache-2.0'
//

package runtimetokenprovider

// RuntimeTokenProvider returns service scope authorization token for secret store during service's run time
type RuntimeTokenProvider interface {
	// GetRawToken generates service scope secretstore token from the runtime service like spiffe token provider
	// and returns authorization token for delayed-start services
	// also returns any error it might have during the whole process
	GetRawToken(serviceKey string) (string, error)
}
