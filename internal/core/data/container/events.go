/********************************************************************************
 *  Copyright 2019 Dell Inc.
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
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
)

// EventsChannelName contains the name of the Events channel instance in the DIC.
var EventsChannelName = "CoreDataEventsChannel"

// PublisherEventsChannelFrom helper function queries the DIC and returns the Events channel instance used for
// publishing over the channel.
//
// NOTE If there is a need to obtain a consuming version of the channel create a new helper function which will get the
// channel from the container and cast it to a consuming channel. The type casting will aid in avoiding errors by
// restricting functionality.
func PublisherEventsChannelFrom(get di.Get) chan<- interface{} {
	return get(EventsChannelName).(chan interface{})
}
