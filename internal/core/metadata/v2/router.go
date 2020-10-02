package v2

import (
	"net/http"

	dpController "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/controller/http"
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

	// Device Profile
	dc := dpController.NewDeviceProfileController(dic)
	r.HandleFunc(v2Constant.ApiDeviceProfileRoute, dc.AddDeviceProfile).Methods(http.MethodPost)
	dp := r.PathPrefix(v2Constant.ApiDeviceProfileRoute).Subrouter()
	dp.HandleFunc(dpController.UploadFile, dc.AddDeviceProfileByYaml).Methods(http.MethodPost)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)
}
