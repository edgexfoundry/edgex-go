package v2

import (
	"net/http"

	metadataController "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/controller/http"
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
	dc := metadataController.NewDeviceProfileController(dic)
	r.HandleFunc(v2Constant.ApiDeviceProfileRoute, dc.AddDeviceProfile).Methods(http.MethodPost)
	r.HandleFunc(v2Constant.ApiDeviceProfileUploadFileRoute, dc.AddDeviceProfileByYaml).Methods(http.MethodPost)
	r.HandleFunc(v2Constant.ApiDeviceProfileUploadFileRoute, dc.UpdateDeviceProfileByYaml).Methods(http.MethodPut)
	r.HandleFunc(v2Constant.ApiDeviceProfileByNameRoute, dc.GetDeviceProfileByName).Methods(http.MethodGet)

	// Device Service
	ds := metadataController.NewDeviceServiceController(dic)
	r.HandleFunc(v2Constant.ApiDeviceServiceRoute, ds.AddDeviceService).Methods(http.MethodPost)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)
}
