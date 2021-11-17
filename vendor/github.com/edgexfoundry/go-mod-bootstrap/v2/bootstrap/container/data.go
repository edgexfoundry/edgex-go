//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
)

// DataEventClientName contains the name of the EventClient instance in the DIC.
var DataEventClientName = di.TypeInstanceToName((*interfaces.EventClient)(nil))

// DataReadingClientName contains the name of the ReadingClient instance in the DIC.
var DataReadingClientName = di.TypeInstanceToName((*interfaces.ReadingClient)(nil))

// DataEventClientFrom helper function queries the DIC and returns the EventClient instance.
func DataEventClientFrom(get di.Get) interfaces.EventClient {
	client, ok := get(DataEventClientName).(interfaces.EventClient)
	if !ok {
		return nil
	}

	return client
}

// DataReadingClientFrom helper function queries the DIC and returns the ReadingClient instance.
func DataReadingClientFrom(get di.Get) interfaces.ReadingClient {
	client, ok := get(DataReadingClientName).(interfaces.ReadingClient)
	if !ok {
		return nil
	}

	return client
}
