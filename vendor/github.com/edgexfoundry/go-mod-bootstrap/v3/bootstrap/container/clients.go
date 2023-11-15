//
// Copyright (c) 2022 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package container

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
)

// EventClientName contains the name of the EventClient's implementation in the DIC.
var EventClientName = di.TypeInstanceToName((*interfaces.EventClient)(nil))

// EventClientFrom helper function queries the DIC and returns the EventClient's implementation.
func EventClientFrom(get di.Get) interfaces.EventClient {
	if get(EventClientName) == nil {
		return nil
	}

	return get(EventClientName).(interfaces.EventClient)
}

// ReadingClientName contains the name of the ReadingClient instance in the DIC.
var ReadingClientName = di.TypeInstanceToName((*interfaces.ReadingClient)(nil))

// ReadingClientFrom helper function queries the DIC and returns the ReadingClient instance.
func ReadingClientFrom(get di.Get) interfaces.ReadingClient {
	client, ok := get(ReadingClientName).(interfaces.ReadingClient)
	if !ok {
		return nil
	}

	return client
}

// CommandClientName contains the name of the CommandClient's implementation in the DIC.
var CommandClientName = di.TypeInstanceToName((*interfaces.CommandClient)(nil))

// CommandClientFrom helper function queries the DIC and returns the CommandClient's implementation.
func CommandClientFrom(get di.Get) interfaces.CommandClient {
	if get(CommandClientName) == nil {
		return nil
	}

	return get(CommandClientName).(interfaces.CommandClient)
}

// NotificationClientName contains the name of the NotificationClient's implementation in the DIC.
var NotificationClientName = di.TypeInstanceToName((*interfaces.NotificationClient)(nil))

// NotificationClientFrom helper function queries the DIC and returns the NotificationClient's implementation.
func NotificationClientFrom(get di.Get) interfaces.NotificationClient {
	if get(NotificationClientName) == nil {
		return nil
	}

	return get(NotificationClientName).(interfaces.NotificationClient)
}

// SubscriptionClientName contains the name of the SubscriptionClient's implementation in the DIC.
var SubscriptionClientName = di.TypeInstanceToName((*interfaces.SubscriptionClient)(nil))

// SubscriptionClientFrom helper function queries the DIC and returns the SubscriptionClient's implementation.
func SubscriptionClientFrom(get di.Get) interfaces.SubscriptionClient {
	if get(SubscriptionClientName) == nil {
		return nil
	}

	return get(SubscriptionClientName).(interfaces.SubscriptionClient)
}

// DeviceServiceClientName contains the name of the DeviceServiceClient's implementation in the DIC.
var DeviceServiceClientName = di.TypeInstanceToName((*interfaces.DeviceServiceClient)(nil))

// DeviceServiceClientFrom helper function queries the DIC and returns the DeviceServiceClient's implementation.
func DeviceServiceClientFrom(get di.Get) interfaces.DeviceServiceClient {
	if get(DeviceServiceClientName) == nil {
		return nil
	}

	return get(DeviceServiceClientName).(interfaces.DeviceServiceClient)
}

// DeviceProfileClientName contains the name of the DeviceProfileClient's implementation in the DIC.
var DeviceProfileClientName = di.TypeInstanceToName((*interfaces.DeviceProfileClient)(nil))

// DeviceProfileClientFrom helper function queries the DIC and returns the DeviceProfileClient's implementation.
func DeviceProfileClientFrom(get di.Get) interfaces.DeviceProfileClient {
	if get(DeviceProfileClientName) == nil {
		return nil
	}

	return get(DeviceProfileClientName).(interfaces.DeviceProfileClient)
}

// DeviceClientName contains the name of the DeviceClient's implementation in the DIC.
var DeviceClientName = di.TypeInstanceToName((*interfaces.DeviceClient)(nil))

// DeviceClientFrom helper function queries the DIC and returns the DeviceClient's implementation.
func DeviceClientFrom(get di.Get) interfaces.DeviceClient {
	if get(DeviceClientName) == nil {
		return nil
	}

	return get(DeviceClientName).(interfaces.DeviceClient)
}

// ProvisionWatcherClientName contains the name of the ProvisionWatcherClient's implementation in the DIC.
var ProvisionWatcherClientName = di.TypeInstanceToName((*interfaces.ProvisionWatcherClient)(nil))

// ProvisionWatcherClientFrom helper function queries the DIC and returns the ProvisionWatcherClient's implementation.
func ProvisionWatcherClientFrom(get di.Get) interfaces.ProvisionWatcherClient {
	if get(ProvisionWatcherClientName) == nil {
		return nil
	}

	return get(ProvisionWatcherClientName).(interfaces.ProvisionWatcherClient)
}

// IntervalClientName contains the name of the IntervalClient's implementation in the DIC.
var IntervalClientName = di.TypeInstanceToName((*interfaces.IntervalClient)(nil))

// IntervalClientFrom helper function queries the DIC and returns the IntervalClient's implementation.
func IntervalClientFrom(get di.Get) interfaces.IntervalClient {
	if get(IntervalClientName) == nil {
		return nil
	}

	return get(IntervalClientName).(interfaces.IntervalClient)
}

// IntervalActionClientName contains the name of the IntervalActionClient's implementation in the DIC.
var IntervalActionClientName = di.TypeInstanceToName((*interfaces.IntervalActionClient)(nil))

// IntervalActionClientFrom helper function queries the DIC and returns the IntervalActionClient's implementation.
func IntervalActionClientFrom(get di.Get) interfaces.IntervalActionClient {
	if get(IntervalActionClientName) == nil {
		return nil
	}

	return get(IntervalActionClientName).(interfaces.IntervalActionClient)
}

// DeviceServiceCallbackClientName contains the name of the DeviceServiceCallbackClient instance in the DIC.
var DeviceServiceCallbackClientName = di.TypeInstanceToName((*interfaces.DeviceServiceCallbackClient)(nil))

// DeviceServiceCommandClientName contains the name of the DeviceServiceCommandClient instance in the DIC.
var DeviceServiceCommandClientName = di.TypeInstanceToName((*interfaces.DeviceServiceCommandClient)(nil))

// DeviceServiceCallbackClientFrom helper function queries the DIC and returns the DeviceServiceCallbackClient instance.
func DeviceServiceCallbackClientFrom(get di.Get) interfaces.DeviceServiceCallbackClient {
	client, ok := get(DeviceServiceCallbackClientName).(interfaces.DeviceServiceCallbackClient)
	if !ok {
		return nil
	}

	return client
}

// DeviceServiceCommandClientFrom helper function queries the DIC and returns the DeviceServiceCommandClient instance.
func DeviceServiceCommandClientFrom(get di.Get) interfaces.DeviceServiceCommandClient {
	client, ok := get(DeviceServiceCommandClientName).(interfaces.DeviceServiceCommandClient)
	if !ok {
		return nil
	}

	return client
}
