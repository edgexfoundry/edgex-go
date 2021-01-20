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

package container

import (
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

// QueueName contains the name of scheduler's SchedulerQueueClient implementation in the DIC.
var QueueName = di.TypeInstanceToName((*interfaces.SchedulerQueueClient)(nil))

// QueueFrom helper function queries the DIC and returns scheduler's SchedulerQueueClient implementation.
func QueueFrom(get di.Get) interfaces.SchedulerQueueClient {
	return get(QueueName).(interfaces.SchedulerQueueClient)
}
