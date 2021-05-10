//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	"github.com/gorilla/mux"

	metadataController "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/v2/controller/http"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container) {
	// v2 API routes
	// Common
	cc := commonController.NewV2CommonController(dic)
	r.HandleFunc(v2.ApiPingRoute, cc.Ping).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiVersionRoute, cc.Version).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiConfigRoute, cc.Config).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiMetricsRoute, cc.Metrics).Methods(http.MethodGet)

	// Device Profile
	dc := metadataController.NewDeviceProfileController(dic)
	r.HandleFunc(v2.ApiDeviceProfileRoute, dc.AddDeviceProfile).Methods(http.MethodPost)
	r.HandleFunc(v2.ApiDeviceProfileRoute, dc.UpdateDeviceProfile).Methods(http.MethodPut)
	r.HandleFunc(v2.ApiDeviceProfileUploadFileRoute, dc.AddDeviceProfileByYaml).Methods(http.MethodPost)
	r.HandleFunc(v2.ApiDeviceProfileUploadFileRoute, dc.UpdateDeviceProfileByYaml).Methods(http.MethodPut)
	r.HandleFunc(v2.ApiDeviceProfileByNameRoute, dc.DeviceProfileByName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceProfileByNameRoute, dc.DeleteDeviceProfileByName).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiAllDeviceProfileRoute, dc.AllDeviceProfiles).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceProfileByModelRoute, dc.DeviceProfilesByModel).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceProfileByManufacturerRoute, dc.DeviceProfilesByManufacturer).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceProfileByManufacturerAndModelRoute, dc.DeviceProfilesByManufacturerAndModel).Methods(http.MethodGet)

	// Device Resource
	dr := metadataController.NewDeviceResourceController(dic)
	r.HandleFunc(v2.ApiDeviceResourceByProfileAndResourceRoute, dr.DeviceResourceByProfileNameAndResourceName).Methods(http.MethodGet)

	// Device Service
	ds := metadataController.NewDeviceServiceController(dic)
	r.HandleFunc(v2.ApiDeviceServiceRoute, ds.AddDeviceService).Methods(http.MethodPost)
	r.HandleFunc(v2.ApiDeviceServiceRoute, ds.PatchDeviceService).Methods(http.MethodPatch)
	r.HandleFunc(v2.ApiDeviceServiceByNameRoute, ds.DeviceServiceByName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceServiceByNameRoute, ds.DeleteDeviceServiceByName).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiAllDeviceServiceRoute, ds.AllDeviceServices).Methods(http.MethodGet)

	// Device
	d := metadataController.NewDeviceController(dic)
	r.HandleFunc(v2.ApiDeviceRoute, d.AddDevice).Methods(http.MethodPost)
	r.HandleFunc(v2.ApiDeviceByNameRoute, d.DeleteDeviceByName).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiDeviceByServiceNameRoute, d.DevicesByServiceName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceNameExistsRoute, d.DeviceNameExists).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceRoute, d.PatchDevice).Methods(http.MethodPatch)
	r.HandleFunc(v2.ApiAllDeviceRoute, d.AllDevices).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceByNameRoute, d.DeviceByName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiDeviceByProfileNameRoute, d.DevicesByProfileName).Methods(http.MethodGet)

	// ProvisionWatcher
	pwc := metadataController.NewProvisionWatcherController(dic)
	r.HandleFunc(v2.ApiProvisionWatcherRoute, pwc.AddProvisionWatcher).Methods(http.MethodPost)
	r.HandleFunc(v2.ApiProvisionWatcherByNameRoute, pwc.ProvisionWatcherByName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiProvisionWatcherByServiceNameRoute, pwc.ProvisionWatchersByServiceName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiProvisionWatcherByProfileNameRoute, pwc.ProvisionWatchersByProfileName).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiAllProvisionWatcherRoute, pwc.AllProvisionWatchers).Methods(http.MethodGet)
	r.HandleFunc(v2.ApiProvisionWatcherByNameRoute, pwc.DeleteProvisionWatcherByName).Methods(http.MethodDelete)
	r.HandleFunc(v2.ApiProvisionWatcherRoute, pwc.PatchProvisionWatcher).Methods(http.MethodPatch)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
