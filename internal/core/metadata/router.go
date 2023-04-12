//
// Copyright (C) 2021-2023 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package metadata

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	metadataController "github.com/edgexfoundry/edgex-go/internal/core/metadata/controller/http"
	commonController "github.com/edgexfoundry/edgex-go/internal/pkg/controller/http"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
)

func LoadRestRoutes(r *mux.Router, dic *di.Container, serviceName string) {
	// r.UseEncodedPath() tells the router to match the encoded original path to the routes
	r.UseEncodedPath()

	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(secretProvider, lc)

	// Common
	cc := commonController.NewCommonController(dic, serviceName)
	r.HandleFunc(common.ApiPingRoute, cc.Ping).Methods(http.MethodGet) // Health check is always unauthenticated
	r.HandleFunc(common.ApiVersionRoute, authenticationHook(cc.Version)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiConfigRoute, authenticationHook(cc.Config)).Methods(http.MethodGet)

	// Units of Measure
	uc := metadataController.NewUnitOfMeasureController(dic)
	r.HandleFunc(common.ApiUnitsOfMeasureRoute, authenticationHook(uc.UnitsOfMeasure)).Methods(http.MethodGet)

	// Device Profile
	dc := metadataController.NewDeviceProfileController(dic)
	r.HandleFunc(common.ApiDeviceProfileRoute, authenticationHook(dc.AddDeviceProfile)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceProfileRoute, authenticationHook(dc.UpdateDeviceProfile)).Methods(http.MethodPut)
	r.HandleFunc(common.ApiDeviceProfileUploadFileRoute, authenticationHook(dc.AddDeviceProfileByYaml)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceProfileUploadFileRoute, authenticationHook(dc.UpdateDeviceProfileByYaml)).Methods(http.MethodPut)
	r.HandleFunc(common.ApiDeviceProfileByNameRoute, authenticationHook(dc.DeviceProfileByName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileByNameRoute, authenticationHook(dc.DeleteDeviceProfileByName)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiAllDeviceProfileRoute, authenticationHook(dc.AllDeviceProfiles)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileByModelRoute, authenticationHook(dc.DeviceProfilesByModel)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileByManufacturerRoute, authenticationHook(dc.DeviceProfilesByManufacturer)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileByManufacturerAndModelRoute, authenticationHook(dc.DeviceProfilesByManufacturerAndModel)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileBasicInfoRoute, authenticationHook(dc.PatchDeviceProfileBasicInfo)).Methods(http.MethodPatch)

	// Device Resource
	dr := metadataController.NewDeviceResourceController(dic)
	r.HandleFunc(common.ApiDeviceResourceByProfileAndResourceRoute, authenticationHook(dr.DeviceResourceByProfileNameAndResourceName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceProfileResourceRoute, authenticationHook(dr.AddDeviceProfileResource)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceProfileResourceRoute, authenticationHook(dr.PatchDeviceProfileResource)).Methods(http.MethodPatch)
	r.HandleFunc(common.ApiDeviceProfileResourceByNameRoute, authenticationHook(dr.DeleteDeviceResourceByName)).Methods(http.MethodDelete)

	// Deivce Command
	dcm := metadataController.NewDeviceCommandController(dic)
	r.HandleFunc(common.ApiDeviceProfileDeviceCommandRoute, authenticationHook(dcm.AddDeviceProfileDeviceCommand)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceProfileDeviceCommandRoute, authenticationHook(dcm.PatchDeviceProfileDeviceCommand)).Methods(http.MethodPatch)
	r.HandleFunc(common.ApiDeviceProfileDeviceCommandByNameRoute, authenticationHook(dcm.DeleteDeviceCommandByName)).Methods(http.MethodDelete)

	// Device Service
	ds := metadataController.NewDeviceServiceController(dic)
	r.HandleFunc(common.ApiDeviceServiceRoute, authenticationHook(ds.AddDeviceService)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceServiceRoute, authenticationHook(ds.PatchDeviceService)).Methods(http.MethodPatch)
	r.HandleFunc(common.ApiDeviceServiceByNameRoute, authenticationHook(ds.DeviceServiceByName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceServiceByNameRoute, authenticationHook(ds.DeleteDeviceServiceByName)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiAllDeviceServiceRoute, authenticationHook(ds.AllDeviceServices)).Methods(http.MethodGet)

	// Device
	d := metadataController.NewDeviceController(dic)
	r.HandleFunc(common.ApiDeviceRoute, authenticationHook(d.AddDevice)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiDeviceByNameRoute, authenticationHook(d.DeleteDeviceByName)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiDeviceByServiceNameRoute, authenticationHook(d.DevicesByServiceName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceNameExistsRoute, authenticationHook(d.DeviceNameExists)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceRoute, authenticationHook(d.PatchDevice)).Methods(http.MethodPatch)
	r.HandleFunc(common.ApiAllDeviceRoute, authenticationHook(d.AllDevices)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceByNameRoute, authenticationHook(d.DeviceByName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiDeviceByProfileNameRoute, authenticationHook(d.DevicesByProfileName)).Methods(http.MethodGet)

	// ProvisionWatcher
	pwc := metadataController.NewProvisionWatcherController(dic)
	r.HandleFunc(common.ApiProvisionWatcherRoute, authenticationHook(pwc.AddProvisionWatcher)).Methods(http.MethodPost)
	r.HandleFunc(common.ApiProvisionWatcherByNameRoute, authenticationHook(pwc.ProvisionWatcherByName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiProvisionWatcherByServiceNameRoute, authenticationHook(pwc.ProvisionWatchersByServiceName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiProvisionWatcherByProfileNameRoute, authenticationHook(pwc.ProvisionWatchersByProfileName)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiAllProvisionWatcherRoute, authenticationHook(pwc.AllProvisionWatchers)).Methods(http.MethodGet)
	r.HandleFunc(common.ApiProvisionWatcherByNameRoute, authenticationHook(pwc.DeleteProvisionWatcherByName)).Methods(http.MethodDelete)
	r.HandleFunc(common.ApiProvisionWatcherRoute, authenticationHook(pwc.PatchProvisionWatcher)).Methods(http.MethodPatch)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.LoggingMiddleware(container.LoggingClientFrom(dic.Get)))
	r.Use(correlation.UrlDecodeMiddleware(container.LoggingClientFrom(dic.Get)))
}
