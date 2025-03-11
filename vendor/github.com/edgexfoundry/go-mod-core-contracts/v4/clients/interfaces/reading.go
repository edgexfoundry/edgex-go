//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package interfaces

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

// ReadingClient defines the interface for interactions with the Reading endpoint on the EdgeX Foundry core-data service.
type ReadingClient interface {
	// AllReadings returns all readings sorted in descending order of created time.
	// The result can be limited in a certain range by specifying the offset and limit parameters.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	AllReadings(ctx context.Context, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX)
	// ReadingCount returns a count of all readings currently stored in the database.
	ReadingCount(ctx context.Context) (common.CountResponse, errors.EdgeX)
	// ReadingCountByDeviceName returns a count of all readings currently stored in the database, sourced from the specified device.
	ReadingCountByDeviceName(ctx context.Context, name string) (common.CountResponse, errors.EdgeX)
	// ReadingsByDeviceName returns a portion of the entire readings according to the device name, offset and limit parameters. Readings are sorted in descending order of created time.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	ReadingsByDeviceName(ctx context.Context, name string, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX)
	// ReadingsByResourceName returns a portion of the entire readings according to the device resource name, offset and limit parameters. Readings are sorted in descending order of created time.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	ReadingsByResourceName(ctx context.Context, name string, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX)
	// ReadingsByTimeRange returns readings between a given start and end date/time. Readings are sorted in descending order of created time.
	// start, end: Unix timestamp, indicating the date/time range.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	ReadingsByTimeRange(ctx context.Context, start, end int64, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX)
	// ReadingsByResourceNameAndTimeRange returns readings by resource name and specified time range. Readings are sorted in descending order of origin time.
	// start, end: Unix timestamp, indicating the date/time range
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	ReadingsByResourceNameAndTimeRange(ctx context.Context, name string, start, end int64, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX)
	// ReadingsByDeviceNameAndResourceName returns readings by device name and resource name. Readings are sorted in descending order of origin time.
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	ReadingsByDeviceNameAndResourceName(ctx context.Context, deviceName, resourceName string, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX)
	// ReadingsByDeviceNameAndResourceNameAndTimeRange returns readings by device name, resource name and specified time range. Readings are sorted in descending order of origin time.
	// start, end: Unix timestamp, indicating the date/time range
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	ReadingsByDeviceNameAndResourceNameAndTimeRange(ctx context.Context, deviceName, resourceName string, start, end int64, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX)
	// ReadingsByDeviceNameAndResourceNamesAndTimeRange returns readings by device name, multiple resource names and specified time range. Readings are sorted in descending order of origin time.
	// If none of resourceNames is specified, return all Readings under specified deviceName and within specified time range
	// start, end: Unix timestamp, indicating the date/time range
	// offset: The number of items to skip before starting to collect the result set. Default is 0.
	// limit: The number of items to return. Specify -1 will return all remaining items after offset. The maximum will be the MaxResultCount as defined in the configuration of service. Default is 20.
	ReadingsByDeviceNameAndResourceNamesAndTimeRange(ctx context.Context, deviceName string, resourceNames []string, start, end int64, offset, limit int) (responses.MultiReadingsResponse, errors.EdgeX)
}
