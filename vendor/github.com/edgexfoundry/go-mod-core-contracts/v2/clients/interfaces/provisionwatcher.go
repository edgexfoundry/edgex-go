//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// ProvisionWatcherClient defines the interface for interactions with the ProvisionWatcher endpoint on the EdgeX Foundry core-metadata service.
type ProvisionWatcherClient interface {
	// Add adds a new provision watcher.
	Add(ctx context.Context, reqs []requests.AddProvisionWatcherRequest) ([]common.BaseWithIdResponse, errors.EdgeX)
	// Update updates provision watchers.
	Update(ctx context.Context, reqs []requests.UpdateProvisionWatcherRequest) ([]common.BaseResponse, errors.EdgeX)
	// AllProvisionWatchers returns all provision watchers. ProvisionWatchers can also be filtered by labels.
	// The result can be limited in a certain range by specifying the offset and limit parameters.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	AllProvisionWatchers(ctx context.Context, labels []string, offset int, limit int) (responses.MultiProvisionWatchersResponse, errors.EdgeX)
	// ProvisionWatcherByName returns a provision watcher by name.
	ProvisionWatcherByName(ctx context.Context, name string) (responses.ProvisionWatcherResponse, errors.EdgeX)
	// DeleteProvisionWatcherByName deletes a provision watcher by name.
	DeleteProvisionWatcherByName(ctx context.Context, name string) (common.BaseResponse, errors.EdgeX)
	// ProvisionWatchersByProfileName returns provision watchers associated with the specified device profile name.
	// The result can be limited in a certain range by specifying the offset and limit parameters.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	ProvisionWatchersByProfileName(ctx context.Context, name string, offset int, limit int) (responses.MultiProvisionWatchersResponse, errors.EdgeX)
	// ProvisionWatchersByServiceName returns provision watchers associated with the specified device service name.
	// The result can be limited in a certain range by specifying the offset and limit parameters.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	ProvisionWatchersByServiceName(ctx context.Context, name string, offset int, limit int) (responses.MultiProvisionWatchersResponse, errors.EdgeX)
}
