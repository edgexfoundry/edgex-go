package v2

import (
	"net/http"

	httpController "github.com/edgexfoundry/edgex-go/internal/core/data/v2/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"

	"github.com/edgexfoundry/go-mod-bootstrap/di"

	v2Constant "github.com/edgexfoundry/go-mod-core-contracts/v2"

	"github.com/gorilla/mux"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container) {
	// v2 API routes
	// Events
	ec := httpController.NewEventController(dic)

	r.HandleFunc(v2Constant.ApiEventRoute, ec.AddEvent).Methods(http.MethodPost)
	r.HandleFunc(v2Constant.ApiEventIdRoute, ec.EventById).Methods(http.MethodGet)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)
}
