package http

import (
	"net/http"

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
	addEventReqDTOs, err := reader.ReadAddEventRequest(r.Body, &ctx)
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
		reqId := addEventReqDTOs[i].RequestID

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
