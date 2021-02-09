package http

import (
	"math"
	"net/http"

	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"
	"github.com/gorilla/mux"
)

type ReadingController struct {
	dic *di.Container
}

// NewReadingController creates and initializes a ReadingController
func NewReadingController(dic *di.Container) *ReadingController {
	return &ReadingController{
		dic: dic,
	}
}

func (rc *ReadingController) ReadingTotalCount(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(rc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var countResponse interface{}
	var statusCode int

	// Count readings
	count, err := application.ReadingTotalCount(rc.dic)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		countResponse = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		countResponse = commonDTO.NewCountResponse("", "", http.StatusOK, count)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(countResponse, w, lc) // encode and send out the countResponse
}

func (rc *ReadingController) AllReadings(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	var response interface{}
	var statusCode int

	// parse URL query string for offset, and limit, and labels
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		readings, err := application.AllReadings(offset, limit, rc.dic)
		if err != nil {
			if errors.Kind(err) != errors.KindEntityDoesNotExist {
				lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			}
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (rc *ReadingController) ReadingsByTimeRange(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	var response interface{}
	var statusCode int

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		readings, err := application.ReadingsByTimeRange(start, end, offset, limit, rc.dic)
		if err != nil {
			if errors.Kind(err) != errors.KindEntityDoesNotExist {
				lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			}
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (rc *ReadingController) ReadingsByResourceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	vars := mux.Vars(r)
	resourceName := vars[v2.ResourceName]

	var response interface{}
	var statusCode int

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		readings, err := application.ReadingsByResourceName(offset, limit, resourceName, rc.dic)
		if err != nil {
			if errors.Kind(err) != errors.KindEntityDoesNotExist {
				lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			}
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (rc *ReadingController) ReadingsByDeviceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	vars := mux.Vars(r)
	name := vars[v2.Name]

	var response interface{}
	var statusCode int

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		readings, err := application.ReadingsByDeviceName(offset, limit, name, rc.dic)
		if err != nil {
			if errors.Kind(err) != errors.KindEntityDoesNotExist {
				lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			}
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (rc *ReadingController) ReadingCountByDeviceName(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(rc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	deviceName := vars[v2.Name]

	var countResponse interface{}
	var statusCode int

	// Count the event by device
	count, err := application.ReadingCountByDeviceName(deviceName, rc.dic)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		countResponse = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		countResponse = commonDTO.NewCountResponse("", "", http.StatusOK, count)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(countResponse, w, lc) // encode and send out the response
}
