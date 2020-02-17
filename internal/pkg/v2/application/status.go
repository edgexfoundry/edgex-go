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

package application

import "github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"

const (
	StatusBatchNotRoutableRequestFailure infrastructure.Status = 20001
	StatusBatchUnmarshalFailure          infrastructure.Status = 20002
	StatusUseCaseContentErrorFailure     infrastructure.Status = 20003
	StatusUseCaseUnmarshalFailure        infrastructure.Status = 20004
	StatusTypeAssertionFailure           infrastructure.Status = 20005
	StatusRequestIdEmptyFailure          infrastructure.Status = 20006
	StatusBehaviorNotValidFailure        infrastructure.Status = 20007

	// core/metadata
	StatusAddressableMissingID       infrastructure.Status = 21000
	StatusAddressableMissingName     infrastructure.Status = 21001
	StatusAddressableMissingProtocol infrastructure.Status = 21002
	StatusAddressableMissingAddress  infrastructure.Status = 21003
)
