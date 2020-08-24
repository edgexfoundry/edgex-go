package http

import (
	"net/http"

	v2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/core/data/v2/io"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
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

	reader := io.NewEventRequestReader()
	addEventReqDTOs, err := reader.ReadAddEventRequest(r.Body, &ctx)
	if err != nil {
		lc.Error(err.Error())
		lc.Debug(err.DebugMessages())
		http.Error(w, err.Message(), err.Code())
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
			lc.Error(err.Error())
			lc.Debug(err.DebugMessages())
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

	w.Header().Set(clients.CorrelationHeader, correlation.FromContext(ctx))
	w.WriteHeader(http.StatusMultiStatus)
	pkg.Encode(addResponses, w, lc) // encode and send out the response
}

func (ec *EventController) EventById(w http.ResponseWriter, r *http.Request) {
	// retrieve all the service injections from bootstrap
	lc := container.LoggingClientFrom(ec.dic.Get)

	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	id := vars[v2.Id]

	var eventResponse interface{}

	// Get the event
	e, err := application.EventById(id, ec.dic)
	if err != nil {
		lc.Error(err.Error())
		lc.Debug(err.DebugMessages())
		eventResponse = commonDTO.NewBaseResponse("", err.Message(), err.Code())
		w.WriteHeader(err.Code())
	} else {
		eventResponse = responseDTO.NewEventResponse("", "", http.StatusOK, e)
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set(clients.CorrelationHeader, correlation.FromContext(ctx))
	pkg.Encode(eventResponse, w, lc) // encode and send out the response
}
