package v2

import (
	"net/http"

	dataController "github.com/edgexfoundry/edgex-go/internal/core/data/v2/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/v2/controller/http"

	"github.com/edgexfoundry/go-mod-bootstrap/di"
	v2Constant "github.com/edgexfoundry/go-mod-core-contracts/v2"

	"github.com/gorilla/mux"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container) {
	// v2 API routes
	// Common
	cc := commonController.NewV2CommonController(dic)
	r.HandleFunc(v2Constant.ApiPingRoute, cc.Ping).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiVersionRoute, cc.Version).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiConfigRoute, cc.Config).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiMetricsRoute, cc.Metrics).Methods(http.MethodGet)

	// Events
	ec := dataController.NewEventController(dic)
	r.HandleFunc(v2Constant.ApiEventRoute, ec.AddEvent).Methods(http.MethodPost)
	r.HandleFunc(v2Constant.ApiEventIdRoute, ec.EventById).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiEventIdRoute, ec.DeleteEventById).Methods(http.MethodDelete)
	r.HandleFunc(v2Constant.ApiEventCountRoute, ec.EventTotalCount).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiEventCountByDeviceNameRoute, ec.EventCountByDevice).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiAllEventRoute, ec.AllEvents).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiEventByDeviceNameRoute, ec.EventsByDeviceName).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiEventByDeviceNameRoute, ec.DeleteEventsByDeviceName).Methods(http.MethodDelete)
	r.HandleFunc(v2Constant.ApiEventByTimeRangeRoute, ec.EventsByTimeRange).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiEventByAgeRoute, ec.DeleteEventsByAge).Methods(http.MethodDelete)

	// Readings
	rc := dataController.NewReadingController(dic)
	r.HandleFunc(v2Constant.ApiReadingCountRoute, rc.ReadingTotalCount).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiAllReadingRoute, rc.AllReadings).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiReadingByDeviceNameRoute, rc.ReadingsByDeviceName).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiReadingByTimeRangeRoute, rc.ReadingsByTimeRange).Methods(http.MethodGet)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)
}
