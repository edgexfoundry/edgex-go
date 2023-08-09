package http

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/core/data/application"
	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	edgexIO "github.com/edgexfoundry/edgex-go/internal/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/utils"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	"github.com/labstack/echo/v4"
)

type EventController struct {
	readers map[string]edgexIO.DtoReader
	mux     sync.RWMutex
	dic     *di.Container
	app     *application.CoreDataApp
}

// NewEventController creates and initializes an EventController
func NewEventController(dic *di.Container) *EventController {
	app := application.CoreDataAppFrom(dic.Get)
	return &EventController{
		readers: make(map[string]edgexIO.DtoReader),
		dic:     dic,
		app:     app,
	}
}

func (ec *EventController) getReader(r *http.Request) edgexIO.DtoReader {
	contentType := strings.ToLower(r.Header.Get(common.ContentType))
	ec.mux.RLock()
	reader, ok := ec.readers[contentType]
	ec.mux.RUnlock()
	if ok {
		return reader
	}

	ec.mux.Lock()
	defer ec.mux.Unlock()
	reader = edgexIO.NewDtoReader(contentType)
	ec.readers[contentType] = reader
	return reader
}

func (ec *EventController) AddEvent(c echo.Context) error {
	r := c.Request()
	w := c.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(ec.dic.Get)

	// URL parameters
	serviceName := c.Param(common.ServiceName)
	profileName := c.Param(common.ProfileName)
	deviceName := c.Param(common.DeviceName)
	sourceName := c.Param(common.SourceName)

	var addEventReqDTO requestDTO.AddEventRequest
	var err errors.EdgeX

	if len(strings.TrimSpace(serviceName)) == 0 {
		err = errors.NewCommonEdgeX(errors.KindContractInvalid, "service name sending event can not be empty", nil)
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	if config.MaxEventSize > 0 && r.ContentLength > config.MaxEventSize*1024 {
		err = errors.NewCommonEdgeX(errors.KindLimitExceeded, fmt.Sprintf("request size exceed %d KB", config.MaxEventSize), nil)
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	dataBytes, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		err = errors.NewCommonEdgeX(errors.KindIOError, "AddEventRequest I/O reading failed", nil)
	} else if r.ContentLength == -1 { // only check the payload byte array size when the Content-Length of Request is unknown
		err = utils.CheckPayloadSize(dataBytes, config.MaxEventSize*1024)
	}

	if err == nil {
		// Per https://github.com/edgexfoundry/edgex-go/pull/3202#discussion_r587618347
		// it is decided to asynchronously publish initially encoded payload (not re-encoding) to message bus
		go ec.app.PublishEvent(dataBytes, serviceName, profileName, deviceName, sourceName, ctx, ec.dic)
		// unmarshal bytes to AddEventRequest
		reader := ec.getReader(r)
		err = reader.Read(bytes.NewReader(dataBytes), &addEventReqDTO)
	}
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	event := requestDTO.AddEventReqToEventModel(addEventReqDTO)
	err = ec.app.ValidateEvent(event, profileName, deviceName, sourceName, ctx, ec.dic)
	if err == nil {
		err = ec.app.AddEvent(event, ctx, ec.dic)
	}
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, addEventReqDTO.RequestId)
	}

	response := commonDTO.NewBaseWithIdResponse(addEventReqDTO.RequestId, "", http.StatusCreated, event.Id)
	utils.WriteHttpHeader(w, ctx, http.StatusCreated)
	// encode and send out the response
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ec *EventController) EventById(c echo.Context) error {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	id := c.Param(common.Id)

	// Get the event
	e, err := ec.app.EventById(id, ec.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewEventResponse("", "", http.StatusOK, e)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	// encode and send out the response
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ec *EventController) DeleteEventById(c echo.Context) error {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	id := c.Param(common.Id)

	// Delete the event
	err := ec.app.DeleteEventById(id, ec.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	// encode and send out the response
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ec *EventController) EventTotalCount(c echo.Context) error {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// Count the event
	count, err := ec.app.EventTotalCount(ec.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewCountResponse("", "", http.StatusOK, count)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc) // encode and send out the response
}

func (ec *EventController) EventCountByDeviceName(c echo.Context) error {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	// URL parameters
	deviceName := c.Param(common.Name)

	// Count the event by device
	count, err := ec.app.EventCountByDeviceName(deviceName, ec.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewCountResponse("", "", http.StatusOK, count)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc) // encode and send out the response
}

func (ec *EventController) AllEvents(c echo.Context) error {
	lc := container.LoggingClientFrom(ec.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(ec.dic.Get)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	events, totalCount, err := ec.app.AllEvents(offset, limit, ec.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	response := responseDTO.NewMultiEventsResponse("", "", http.StatusOK, totalCount, events)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ec *EventController) EventsByDeviceName(c echo.Context) error {
	lc := container.LoggingClientFrom(ec.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(ec.dic.Get)

	name := c.Param(common.Name)

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	events, totalCount, err := ec.app.EventsByDeviceName(offset, limit, name, ec.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiEventsResponse("", "", http.StatusOK, totalCount, events)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ec *EventController) DeleteEventsByDeviceName(c echo.Context) error {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	deviceName := c.Param(common.Name)

	// Delete events with associated Device deviceName
	err := ec.app.DeleteEventsByDeviceName(deviceName, ec.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusAccepted)
	utils.WriteHttpHeader(w, ctx, http.StatusAccepted)
	// encode and send out the response
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ec *EventController) EventsByTimeRange(c echo.Context) error {
	lc := container.LoggingClientFrom(ec.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()
	config := dataContainer.ConfigurationFrom(ec.dic.Get)

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(c, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	events, totalCount, err := ec.app.EventsByTimeRange(start, end, offset, limit, ec.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := responseDTO.NewMultiEventsResponse("", "", http.StatusOK, totalCount, events)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	return pkg.EncodeAndWriteResponse(response, w, lc)
}

func (ec *EventController) DeleteEventsByAge(c echo.Context) error {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)
	r := c.Request()
	w := c.Response()
	ctx := r.Context()

	age, parsingErr := strconv.ParseInt(c.Param(common.Age), 10, 64)

	if parsingErr != nil {
		err := errors.NewCommonEdgeX(errors.KindContractInvalid, "age format parsing failed", parsingErr)
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}
	err := ec.app.DeleteEventsByAge(age, ec.dic)
	if err != nil {
		return utils.WriteErrorResponse(w, ctx, lc, err, "")
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusAccepted)
	utils.WriteHttpHeader(w, ctx, http.StatusAccepted)
	// encode and send out the response
	return pkg.EncodeAndWriteResponse(response, w, lc)
}
