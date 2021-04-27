package http

import (
	"math"
	"net/http"

	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
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

	// Count readings
	count, err := application.ReadingTotalCount(rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewCountResponse("", "", http.StatusOK, count)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc) // encode and send out the countResponse
}

func (rc *ReadingController) AllReadings(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	// parse URL query string for offset, and limit, and labels
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	readings, err := application.AllReadings(offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (rc *ReadingController) ReadingsByTimeRange(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	readings, err := application.ReadingsByTimeRange(start, end, offset, limit, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (rc *ReadingController) ReadingsByResourceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	vars := mux.Vars(r)
	resourceName := vars[v2.ResourceName]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	readings, err := application.ReadingsByResourceName(offset, limit, resourceName, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (rc *ReadingController) ReadingsByDeviceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(rc.dic.Get)
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(rc.dic.Get)

	vars := mux.Vars(r)
	name := vars[v2.Name]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	readings, err := application.ReadingsByDeviceName(offset, limit, name, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiReadingsResponse("", "", http.StatusOK, readings)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (rc *ReadingController) ReadingCountByDeviceName(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(rc.dic.Get)

	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	deviceName := vars[v2.Name]

	// Count the event by device
	count, err := application.ReadingCountByDeviceName(deviceName, rc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewCountResponse("", "", http.StatusOK, count)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc) // encode and send out the response
}
