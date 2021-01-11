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
	r.HandleFunc(v2Constant.ApiDeviceProfileRoute, dc.UpdateDeviceProfile).Methods(http.MethodPut)
	r.HandleFunc(v2Constant.ApiDeviceProfileUploadFileRoute, dc.AddDeviceProfileByYaml).Methods(http.MethodPost)
	r.HandleFunc(v2Constant.ApiDeviceProfileUploadFileRoute, dc.UpdateDeviceProfileByYaml).Methods(http.MethodPut)
	r.HandleFunc(v2Constant.ApiDeviceProfileByNameRoute, dc.DeviceProfileByName).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiDeviceProfileByIdRoute, dc.DeleteDeviceProfileById).Methods(http.MethodDelete)
	r.HandleFunc(v2Constant.ApiDeviceProfileByNameRoute, dc.DeleteDeviceProfileByName).Methods(http.MethodDelete)
	r.HandleFunc(v2Constant.ApiAllDeviceProfileRoute, dc.AllDeviceProfiles).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiDeviceProfileByModelRoute, dc.DeviceProfilesByModel).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiDeviceProfileByManufacturerRoute, dc.DeviceProfilesByManufacturer).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiDeviceProfileByManufacturerAndModelRoute, dc.DeviceProfilesByManufacturerAndModel).Methods(http.MethodGet)

	// Device Service
	ds := metadataController.NewDeviceServiceController(dic)
	r.HandleFunc(v2Constant.ApiDeviceServiceRoute, ds.AddDeviceService).Methods(http.MethodPost)
	r.HandleFunc(v2Constant.ApiDeviceServiceRoute, ds.PatchDeviceService).Methods(http.MethodPatch)
	r.HandleFunc(v2Constant.ApiDeviceServiceByNameRoute, ds.DeviceServiceByName).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiDeviceServiceByIdRoute, ds.DeleteDeviceServiceById).Methods(http.MethodDelete)
	r.HandleFunc(v2Constant.ApiDeviceServiceByNameRoute, ds.DeleteDeviceServiceByName).Methods(http.MethodDelete)
	r.HandleFunc(v2Constant.ApiAllDeviceServiceRoute, ds.AllDeviceServices).Methods(http.MethodGet)

	// Device
	d := metadataController.NewDeviceController(dic)
	r.HandleFunc(v2Constant.ApiDeviceRoute, d.AddDevice).Methods(http.MethodPost)
	r.HandleFunc(v2Constant.ApiDeviceByIdRoute, d.DeleteDeviceById).Methods(http.MethodDelete)
	r.HandleFunc(v2Constant.ApiDeviceByNameRoute, d.DeleteDeviceByName).Methods(http.MethodDelete)
	r.HandleFunc(v2Constant.ApiDeviceByServiceNameRoute, d.DevicesByServiceName).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiDeviceIdExistsRoute, d.DeviceIdExists).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiDeviceNameExistsRoute, d.DeviceNameExists).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiDeviceRoute, d.PatchDevice).Methods(http.MethodPatch)
	r.HandleFunc(v2Constant.ApiAllDeviceRoute, d.AllDevices).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiDeviceByNameRoute, d.DeviceByName).Methods(http.MethodGet)
	r.HandleFunc(v2Constant.ApiDeviceByProfileNameRoute, d.DevicesByProfileName).Methods(http.MethodGet)

	// ProvisionWatcher
	pwc := metadataController.NewProvisionWatcherController(dic)
	r.HandleFunc(v2Constant.ApiProvisionWatcherRoute, pwc.AddProvisionWatcher).Methods(http.MethodPost)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)
}
