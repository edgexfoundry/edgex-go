//
// Copyright (C) 2021-2022 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package metadata

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

	metadataController "github.com/edgexfoundry/edgex-go/internal/core/metadata/controller/http"
	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container, serviceName string) {
	// Common
	cc := commonController.NewCommonController(dic, serviceName)
	r.HandleFunc(common.ApiPingRoute, cc.Ping).Methods(http.MethodGet)
	r.HandleFunc(common.ApiVersionRoute, cc.Version).Methods(http.MethodGet)
	r.HandleFunc(common.ApiConfigRoute, cc.Config).Methods(http.MethodGet)
	r.HandleFunc(common.ApiMetricsRoute, cc.Metrics).Methods(http.MethodGet)

	// Units of Measure
	uc := metadataController.NewUnitOfMeasureController(dic)
	r.HandleFunc(common.ApiUnitsOfMeasureRoute, uc.UnitsOfMeasure).Methods(http.MethodGet)

	// Device Profile
	dc := metadataController.NewDeviceProfileController(dic)
	r.HandleFunc(common.ApiDeviceProfileRoute, dc.AddDeviceProfile).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceProfileRoute, dc.UpdateDeviceProfile).Methods(http.MethodPut)
	r.HandleFunc(common.ApiDeviceProfileUploadFileRoute, dc.AddDeviceProfileByYaml).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceProfileUploadFileRoute, dc.UpdateDeviceProfileByYaml).Methods(http.MethodPut)
	r.HandleFunc(common.ApiDeviceProfileByNameRoute, dc.DeviceProfileByName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileByNameRoute, dc.DeleteDeviceProfileByName).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiAllDeviceProfileRoute, dc.AllDeviceProfiles).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileByModelRoute, dc.DeviceProfilesByModel).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileByManufacturerRoute, dc.DeviceProfilesByManufacturer).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileByManufacturerAndModelRoute, dc.DeviceProfilesByManufacturerAndModel).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileBasicInfoRoute, dc.PatchDeviceProfileBasicInfo).Methods(http.MethodPatch)

	// Device Resource
	dr := metadataController.NewDeviceResourceController(dic)
	r.HandleFunc(common.ApiDeviceResourceByProfileAndResourceRoute, dr.DeviceResourceByProfileNameAndResourceName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileResourceRoute, dr.AddDeviceProfileResource).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceProfileResourceRoute, dr.PatchDeviceProfileResource).Methods(http.MethodPatch)
	r.HandleFunc(common.ApiDeviceProfileResourceByNameRoute, dr.DeleteDeviceResourceByName).Methods(http.MethodDelete)

	// Deivce Command
	dcm := metadataController.NewDeviceCommandController(dic)
	r.HandleFunc(common.ApiDeviceProfileDeviceCommandRoute, dcm.AddDeviceProfileDeviceCommand).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceProfileDeviceCommandRoute, dcm.PatchDeviceProfileDeviceCommand).Methods(http.MethodPatch)
	r.HandleFunc(common.ApiDeviceProfileDeviceCommandByNameRoute, dcm.DeleteDeviceCommandByName).Methods(http.MethodDelete)

	// Device Service
	ds := metadataController.NewDeviceServiceController(dic)
	r.HandleFunc(common.ApiDeviceServiceRoute, ds.AddDeviceService).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceServiceRoute, ds.PatchDeviceService).Methods(http.MethodPatch)
	r.HandleFunc(common.ApiDeviceServiceByNameRoute, ds.DeviceServiceByName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceServiceByNameRoute, ds.DeleteDeviceServiceByName).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiAllDeviceServiceRoute, ds.AllDeviceServices).Methods(http.MethodGet)

	// Device
	d := metadataController.NewDeviceController(dic)
	r.HandleFunc(common.ApiDeviceRoute, d.AddDevice).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceByNameRoute, d.DeleteDeviceByName).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiDeviceByServiceNameRoute, d.DevicesByServiceName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceNameExistsRoute, d.DeviceNameExists).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceRoute, d.PatchDevice).Methods(http.MethodPatch)
	r.HandleFunc(common.ApiAllDeviceRoute, d.AllDevices).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceByNameRoute, d.DeviceByName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceByProfileNameRoute, d.DevicesByProfileName).Methods(http.MethodGet)

	// ProvisionWatcher
	pwc := metadataController.NewProvisionWatcherController(dic)
	r.HandleFunc(common.ApiProvisionWatcherRoute, pwc.AddProvisionWatcher).Methods(http.MethodPost)
	r.HandleFunc(common.ApiProvisionWatcherByNameRoute, pwc.ProvisionWatcherByName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiProvisionWatcherByServiceNameRoute, pwc.ProvisionWatchersByServiceName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiProvisionWatcherByProfileNameRoute, pwc.ProvisionWatchersByProfileName).Methods(http.MethodGet)
	r.HandleFunc(common.ApiAllProvisionWatcherRoute, pwc.AllProvisionWatchers).Methods(http.MethodGet)
	r.HandleFunc(common.ApiProvisionWatcherByNameRoute, pwc.DeleteProvisionWatcherByName).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiProvisionWatcherRoute, pwc.PatchProvisionWatcher).Methods(http.MethodPatch)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
}
