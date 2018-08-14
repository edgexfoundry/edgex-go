/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
package messaging

import "github.com/edgexfoundry/edgex-go/pkg/models"

// Mock implementation of the event publisher for testing purposes
type mockEventPublisher struct {
	pubType int
}

func newMockEventPublisher(config PubSubConfiguration) EventPublisher {
	return &mockEventPublisher{
		pubType:   MOCK,
	}
}

func (zep *mockEventPublisher) SendEventMessage(e models.Event) error {
	return nil
}