//
// Copyright (C) 2021-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"fmt"
	"math"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/core/data/application"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/core/data/query"
	"github.com/edgexfoundry/edgex-go/internal/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
)

type ReadingController struct {
	reader io.DtoReader
	dic    *di.Container
	app    *application.CoreDataApp
}

// NewReadingController creates and initializes a ReadingController
func NewReadingController(dic *di.Container) *ReadingController {
	app := application.CoreDataAppFrom(dic.Get)
	return &ReadingController{
		reader: io.NewJsonDtoReader(),
		dic:    dic,
		app:    app,
	}
}

func (rc *ReadingController) ReadingTotalCount(c echo.Context) error {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(rc.dic.Get)

	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// Count readings
	count, err := rc.app.ReadingTotalCount(rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewCountResponse("", "", http.StatusOK, count)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc) // encode and send out the countResponse
}

func (rc *ReadingController) AllReadings(c echo.Context) error {
	lc := container.LoggingClientFrom(rc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	// parse URL query string for offset, and limit, and labels
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, minOffset, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	parms := query.Parameters{
		Offset: offset, Limit: limit,
		Numeric: cast.ToBool(c.QueryParam(common.Numeric))}

	aggFuncParam := c.QueryParam(common.AggregateFunc)
	if aggFuncParam != "" {
		// Specify the app layer function to be invoked to get the aggregated reading values
		aggReadingsFunc := func(aggFunc string) ([]dtos.BaseReading, errors.EdgeX) {
			return rc.app.AllAggregateReadings(aggFunc, rc.dic, parms)
		}
		return handleReadingAggregation(w, ctx, lc, aggFuncParam, aggReadingsFunc)
	}

	readings, totalCount, err := rc.app.AllReadings(parms, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByTimeRange(c echo.Context) error {
	lc := container.LoggingClientFrom(rc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(c, minOffset, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	parms := query.Parameters{
		Start: start, End: end, Offset: offset, Limit: limit,
		Numeric: cast.ToBool(c.QueryParam(common.Numeric))}

	aggFuncParam := c.QueryParam(common.AggregateFunc)
	if aggFuncParam != "" {
		// Specify the app layer function to be invoked to get the aggregated reading values
		aggReadingsFunc := func(aggFunc string) ([]dtos.BaseReading, errors.EdgeX) {
			return rc.app.AllAggregateReadingsByTimeRange(aggFunc, parms, rc.dic)
		}
		return handleReadingAggregation(w, ctx, lc, aggFuncParam, aggReadingsFunc)
	}

	readings, totalCount, err := rc.app.ReadingsByTimeRange(parms, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByResourceName(c echo.Context) error {
	lc := container.LoggingClientFrom(rc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	resourceName := c.Param(common.ResourceName)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, minOffset, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	parms := query.Parameters{
		Offset: offset, Limit: limit,
		Numeric: cast.ToBool(c.QueryParam(common.Numeric))}

	aggFuncParam := c.QueryParam(common.AggregateFunc)
	if aggFuncParam != "" {
		// Specify the app layer function to be invoked to get the aggregated reading values
		aggReadingsFunc := func(aggFunc string) ([]dtos.BaseReading, errors.EdgeX) {
			return rc.app.AggregateReadingsByResourceName(resourceName, aggFunc, rc.dic, parms)
		}
		return handleReadingAggregation(w, ctx, lc, aggFuncParam, aggReadingsFunc)
	}

	readings, totalCount, err := rc.app.ReadingsByResourceName(parms, resourceName, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByDeviceName(c echo.Context) error {
	lc := container.LoggingClientFrom(rc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	name := c.Param(common.Name)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, minOffset, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	parms := query.Parameters{
		Offset: offset, Limit: limit,
		Numeric: cast.ToBool(c.QueryParam(common.Numeric))}

	aggFuncParam := c.QueryParam(common.AggregateFunc)
	if aggFuncParam != "" {
		// Specify the app layer function to be invoked to get the aggregated reading values
		aggReadingsFunc := func(aggFunc string) ([]dtos.BaseReading, errors.EdgeX) {
			return rc.app.AggregateReadingsByDeviceName(name, aggFunc, rc.dic, parms)
		}
		return handleReadingAggregation(w, ctx, lc, aggFuncParam, aggReadingsFunc)
	}

	readings, totalCount, err := rc.app.ReadingsByDeviceName(parms, name, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingCountByDeviceName(c echo.Context) error {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(rc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	deviceName := c.Param(common.Name)

	// Count the event by device
	count, err := rc.app.ReadingCountByDeviceName(deviceName, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewCountResponse("", "", http.StatusOK, count)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc) // encode and send out the response
}

// ReadingsByResourceNameAndTimeRange returns readings by resource name and specified time range. Readings are sorted in descending order of origin time.
func (rc *ReadingController) ReadingsByResourceNameAndTimeRange(c echo.Context) error {
	lc := container.LoggingClientFrom(rc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	resourceName := c.Param(common.ResourceName)

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(c, minOffset, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	parms := query.Parameters{
		Start: start, End: end, Offset: offset, Limit: limit,
		Numeric: cast.ToBool(c.QueryParam(common.Numeric))}

	aggFuncParam := c.QueryParam(common.AggregateFunc)
	if aggFuncParam != "" {
		// Specify the app layer function to be invoked to get the aggregated reading values
		aggReadingsFunc := func(aggFunc string) ([]dtos.BaseReading, errors.EdgeX) {
			return rc.app.AggregateReadingsByResourceNameAndTimeRange(resourceName, aggFunc, parms, rc.dic)
		}
		return handleReadingAggregation(w, ctx, lc, aggFuncParam, aggReadingsFunc)
	}

	readings, totalCount, err := rc.app.ReadingsByResourceNameAndTimeRange(resourceName, parms, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByDeviceNameAndResourceName(c echo.Context) error {
	lc := container.LoggingClientFrom(rc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	deviceName := c.Param(common.Name)
	resourceName := c.Param(common.ResourceName)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, minOffset, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	parms := query.Parameters{
		Offset: offset, Limit: limit,
		Numeric: cast.ToBool(c.QueryParam(common.Numeric))}

	aggFuncParam := c.QueryParam(common.AggregateFunc)
	if aggFuncParam != "" {
		// Specify the app layer function to be invoked to get the aggregated reading values
		aggReadingsFunc := func(aggFunc string) ([]dtos.BaseReading, errors.EdgeX) {
			return rc.app.AggregateReadingsByDeviceNameAndResourceName(deviceName, resourceName, aggFunc, rc.dic, parms)
		}
		return handleReadingAggregation(w, ctx, lc, aggFuncParam, aggReadingsFunc)
	}

	readings, totalCount, err := rc.app.ReadingsByDeviceNameAndResourceName(deviceName, resourceName, parms, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByDeviceNameAndResourceNameAndTimeRange(c echo.Context) error {
	lc := container.LoggingClientFrom(rc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	deviceName := c.Param(common.Name)
	resourceName := c.Param(common.ResourceName)

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(c, minOffset, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	parms := query.Parameters{
		Start: start, End: end, Offset: offset, Limit: limit,
		Numeric: cast.ToBool(c.QueryParam(common.Numeric))}

	aggFuncParam := c.QueryParam(common.AggregateFunc)
	if aggFuncParam != "" {
		// Specify the app layer function to be invoked to get the aggregated reading values
		aggReadingsFunc := func(aggFunc string) ([]dtos.BaseReading, errors.EdgeX) {
			return rc.app.AggregateReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName, resourceName, aggFunc, parms, rc.dic)
		}
		return handleReadingAggregation(w, ctx, lc, aggFuncParam, aggReadingsFunc)
	}

	readings, totalCount, err := rc.app.ReadingsByDeviceNameAndResourceNameAndTimeRange(deviceName, resourceName, parms, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (rc *ReadingController) ReadingsByDeviceNameAndResourceNamesAndTimeRange(c echo.Context) error {
	lc := container.LoggingClientFrom(rc.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	deviceName := c.Param(common.Name)

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(c, minOffset, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	parms := query.Parameters{
		Start: start, End: end, Offset: offset, Limit: limit,
		Numeric: cast.ToBool(c.QueryParam(common.Numeric))}

	aggFuncParam := c.QueryParam(common.AggregateFunc)
	if aggFuncParam != "" {
		// Specify the app layer function to be invoked to get the aggregated reading values
		aggReadingsFunc := func(aggFunc string) ([]dtos.BaseReading, errors.EdgeX) {
			return rc.app.AggregateReadingsByDeviceNameAndTimeRange(deviceName, aggFunc, parms, rc.dic)
		}
		return handleReadingAggregation(w, ctx, lc, aggFuncParam, aggReadingsFunc)
	}

	var queryPayload map[string]interface{}
	if r.Body != http.NoBody { //only parse request body when there are contents provided
		err = rc.reader.Read(r.Body, &queryPayload)
		if err != nil {
			return utils.WriteErrorResponse(w, ctx, lc, err, "")
		}
	}

	var resourceNames []string
	if val, ok := queryPayload[common.ResourceNames]; ok { //look for
		switch t := val.(type) {
		case []interface{}:
			for _, v := range t {
				if strVal, ok := v.(string); ok {
					resourceNames = append(resourceNames, strVal)
				}
			}
		default:
			err = errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("query criteria [%v] not in expected format", common.ResourceNames), nil)
			return utils.WriteErrorResponse(w, ctx, lc, err, "")
		}
	}

	readings, totalCount, err := rc.app.ReadingsByDeviceNameAndResourceNamesAndTimeRange(deviceName, resourceNames, parms, rc.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, totalCount, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

// handleReadingAggregation parses the aggregateFunc query parameter, calls the provided application-layer function
// to compute the aggregated reading values, and returns a MultiReadingsAggregationResponse DTO.
func handleReadingAggregation(
	w *echo.Response,
	ctx context.Context,
	lc logger.LoggingClient,
	aggFuncParam string,
	aggReadingFunc func(string) ([]dtos.BaseReading, errors.EdgeX),
) error {
	aggFunc, err := utils.ParseAggregateFuncQueryString(aggFuncParam)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	readings, err := aggReadingFunc(aggFunc)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiReadingsAggregationResponse("", "", http.StatusOK, aggFuncParam, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}
