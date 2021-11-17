//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// SystemManagementClient defines the interface for interactions with the API endpoint on the EdgeX Foundry sys-mgmt-agent service.
type SystemManagementClient interface {
	// GetHealth obtain health information of services via registry by their name
	GetHealth(ctx context.Context, services []string) ([]common.BaseWithServiceNameResponse, errors.EdgeX)
	// GetMetrics obtain metrics information from services by their name
	GetMetrics(ctx context.Context, services []string) ([]common.BaseWithMetricsResponse, errors.EdgeX)
	// GetConfig obtain configuration from services by their name
	GetConfig(ctx context.Context, services []string) ([]common.BaseWithConfigResponse, errors.EdgeX)
	// DoOperation issue a start, stop, restart action to the targeted services
	DoOperation(ctx context.Context, reqs []requests.OperationRequest) ([]common.BaseResponse, errors.EdgeX)
}
