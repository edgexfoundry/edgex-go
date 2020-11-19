package http

import (
	"math"
	"net/http"

	dataContainer "github.com/edgexfoundry/edgex-go/internal/core/data/container"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type EventController struct {
	reader io.EventReader
	dic    *di.Container
}

// NewEventController creates and initializes an EventController
func NewEventController(dic *di.Container) *EventController {
	return &EventController{
		reader: io.NewEventRequestReader(),
		dic:    dic,
	}
}

func (ec *EventController) AddEvent(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	reader := io.NewEventRequestReader()
	addEventReqDTOs, err := reader.ReadAddEventRequest(r.Body)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		errResponses := commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		// encode and send out the response
		pkg.Encode(errResponses, w, lc)
		return
	}
	events := requestDTO.AddEventReqToEventModels(addEventReqDTOs)

	// map Event models to AddEventResponse DTOs
	var addResponses []interface{}
	for i, e := range events {
		newId, err := application.AddEvent(e, ctx, ec.dic)
		var addEventResponse interface{}
		// get the requestID from AddEventRequestDTO
		reqId := addEventReqDTOs[i].RequestId

		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			addEventResponse = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			addEventResponse = commonDTO.NewBaseWithIdResponse(
				reqId,
				"",
				http.StatusCreated,
				newId)
		}
		addResponses = append(addResponses, addEventResponse)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	// encode and send out the response
	pkg.Encode(addResponses, w, lc)
}

func (ec *EventController) EventById(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	id := vars[v2.Id]

	var eventResponse interface{}
	var statusCode int

	// Get the event
	e, err := application.EventById(id, ec.dic)
	if err != nil {
		// Event not found is not a real error, so the error message should not be printed out
		if errors.Kind(err) != errors.KindEntityDoesNotExist {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		}
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		eventResponse = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		eventResponse = responseDTO.NewEventResponse("", "", http.StatusOK, e)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	// encode and send out the response
	pkg.Encode(eventResponse, w, lc)
}

func (ec *EventController) DeleteEventById(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	id := vars[v2.Id]

	var response interface{}
	var statusCode int

	// Delete the event
	err := application.DeleteEventById(id, ec.dic)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = commonDTO.NewBaseResponse("", "", http.StatusOK)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	// encode and send out the response
	pkg.Encode(response, w, lc)
}

func (ec *EventController) EventTotalCount(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var eventResponse interface{}
	var statusCode int

	// Count the event
	count, err := application.EventTotalCount(ec.dic)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		eventResponse = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		eventResponse = responseDTO.NewEventCountResponse("", "", http.StatusOK, count, "")
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(eventResponse, w, lc) // encode and send out the response
}

func (ec *EventController) EventCountByDevice(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	// URL parameters
	vars := mux.Vars(r)
	deviceName := vars[v2.DeviceName]

	var eventResponse interface{}
	var statusCode int

	// Count the event by device
	count, err := application.EventCountByDevice(deviceName, ec.dic)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		eventResponse = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		eventResponse = responseDTO.NewEventCountResponse("", "", http.StatusOK, count, deviceName)
		statusCode = http.StatusOK
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(eventResponse, w, lc) // encode and send out the response
}

func (ec *EventController) DeletePushedEvents(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	var response interface{}
	var statusCode int

	err := application.DeletePushedEvents(ec.dic)
	if err != nil {
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = commonDTO.NewBaseResponse("", "", http.StatusAccepted)
		statusCode = http.StatusAccepted
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	// encode and send out the response
	pkg.Encode(response, w, lc)
}

func (ec *EventController) UpdateEventPushedById(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	updateEventPushedReqs, err := ec.reader.ReadUpdateEventPushedByIdRequest(r.Body)
	if err != nil {
		lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		errResponses := commonDTO.NewBaseResponse(
			"",
			err.Message(),
			err.Code())
		utils.WriteHttpHeader(w, ctx, err.Code())
		// encode and send out the response
		pkg.Encode(errResponses, w, lc)
		return
	}

	var updatedResponses []interface{}
	for _, req := range updateEventPushedReqs {
		err := application.UpdateEventPushedById(req.Id, ec.dic)
		var updateEventPushedResponse interface{}
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			updateEventPushedResponse = responseDTO.NewUpdateEventPushedByIdResponse(req.RequestId, err.Message(), err.Code(), req.Id)
		} else {
			updateEventPushedResponse = responseDTO.NewUpdateEventPushedByIdResponse(req.RequestId, "", http.StatusOK, req.Id)
		}
		updatedResponses = append(updatedResponses, updateEventPushedResponse)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	// encode and send out the response
	pkg.Encode(updatedResponses, w, lc)
}

func (ec *EventController) AllEvents(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(ec.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := dataContainer.ConfigurationFrom(ec.dic.Get)

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
		events, err := application.AllEvents(offset, limit, ec.dic)
		if err != nil {
			if errors.Kind(err) != errors.KindEntityDoesNotExist {
				lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			}
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiEventsResponse("", "", http.StatusOK, events)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (ec *EventController) EventsByDeviceName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(ec.dic.Get)
	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)
	config := dataContainer.ConfigurationFrom(ec.dic.Get)

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
		events, err := application.EventsByDeviceName(offset, limit, name, ec.dic)
		if err != nil {
			if errors.Kind(err) != errors.KindEntityDoesNotExist {
				lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			}
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
			statusCode = err.Code()
		} else {
			response = responseDTO.NewMultiEventsResponse("", "", http.StatusOK, events)
			statusCode = http.StatusOK
		}
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	pkg.Encode(response, w, lc)
}

func (ec *EventController) DeleteEventsByDeviceName(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	vars := mux.Vars(r)
	deviceName := vars[v2.Name]

	var response interface{}
	var statusCode int

	// Delete events with associated Device deviceName
	err := application.DeleteEventsByDeviceName(deviceName, ec.dic)
	if err != nil {
		lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
		response = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		statusCode = err.Code()
	} else {
		response = commonDTO.NewBaseResponse("", "", http.StatusAccepted)
		statusCode = http.StatusAccepted
	}

	utils.WriteHttpHeader(w, ctx, statusCode)
	// encode and send out the response
	pkg.Encode(response, w, lc)
}
