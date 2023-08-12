//
// Copyright (C) 2021-2023 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package metadata

import (
	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"

	metadataController "github.com/edgexfoundry/edgex-go/internal/core/metadata/controller/http"

	"github.com/labstack/echo/v4"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	lc := container.LoggingClientFrom(dic.Get)
	secretProvider := container.SecretProviderExtFrom(dic.Get)
	authenticationHook := handlers.AutoConfigAuthenticationFunc(secretProvider, lc)

	// Common
	_ = controller.NewCommonController(dic, r, serviceName, edgex.Version)

	// Units of Measure
	uc := metadataController.NewUnitOfMeasureController(dic)
	r.GET(common.ApiUnitsOfMeasureRoute, uc.UnitsOfMeasure, authenticationHook)

	// Device Profile
	dc := metadataController.NewDeviceProfileController(dic)
	r.POST(common.ApiDeviceProfileRoute, dc.AddDeviceProfile, authenticationHook)
	r.PUT(common.ApiDeviceProfileRoute, dc.UpdateDeviceProfile, authenticationHook)
	r.POST(common.ApiDeviceProfileUploadFileRoute, dc.AddDeviceProfileByYaml, authenticationHook)
	r.PUT(common.ApiDeviceProfileUploadFileRoute, dc.UpdateDeviceProfileByYaml, authenticationHook)
	r.GET(common.ApiDeviceProfileByNameEchoRoute, dc.DeviceProfileByName, authenticationHook)
	r.DELETE(common.ApiDeviceProfileByNameEchoRoute, dc.DeleteDeviceProfileByName, authenticationHook)
	r.GET(common.ApiAllDeviceProfileRoute, dc.AllDeviceProfiles, authenticationHook)
	r.GET(common.ApiDeviceProfileByModelEchoRoute, dc.DeviceProfilesByModel, authenticationHook)
	r.GET(common.ApiDeviceProfileByManufacturerEchoRoute, dc.DeviceProfilesByManufacturer, authenticationHook)
	r.GET(common.ApiDeviceProfileByManufacturerAndModelEchoRoute, dc.DeviceProfilesByManufacturerAndModel, authenticationHook)
	r.PATCH(common.ApiDeviceProfileBasicInfoRoute, dc.PatchDeviceProfileBasicInfo, authenticationHook)

	// Device Resource
	dr := metadataController.NewDeviceResourceController(dic)
	r.GET(common.ApiDeviceResourceByProfileAndResourceEchoRoute, dr.DeviceResourceByProfileNameAndResourceName, authenticationHook)
	r.POST(common.ApiDeviceProfileResourceRoute, dr.AddDeviceProfileResource, authenticationHook)
	r.PATCH(common.ApiDeviceProfileResourceRoute, dr.PatchDeviceProfileResource, authenticationHook)
	r.DELETE(common.ApiDeviceProfileResourceByNameEchoRoute, dr.DeleteDeviceResourceByName, authenticationHook)

	// Deivce Command
	dcm := metadataController.NewDeviceCommandController(dic)
	r.POST(common.ApiDeviceProfileDeviceCommandRoute, dcm.AddDeviceProfileDeviceCommand, authenticationHook)
	r.PATCH(common.ApiDeviceProfileDeviceCommandRoute, dcm.PatchDeviceProfileDeviceCommand, authenticationHook)
	r.DELETE(common.ApiDeviceProfileDeviceCommandByNameEchoRoute, dcm.DeleteDeviceCommandByName, authenticationHook)

	// Device Service
	ds := metadataController.NewDeviceServiceController(dic)
	r.POST(common.ApiDeviceServiceRoute, ds.AddDeviceService, authenticationHook)
	r.PATCH(common.ApiDeviceServiceRoute, ds.PatchDeviceService, authenticationHook)
	r.GET(common.ApiDeviceServiceByNameEchoRoute, ds.DeviceServiceByName, authenticationHook)
	r.DELETE(common.ApiDeviceServiceByNameEchoRoute, ds.DeleteDeviceServiceByName, authenticationHook)
	r.GET(common.ApiAllDeviceServiceRoute, ds.AllDeviceServices, authenticationHook)

	// Device
	d := metadataController.NewDeviceController(dic)
	r.POST(common.ApiDeviceRoute, d.AddDevice, authenticationHook)
	r.DELETE(common.ApiDeviceByNameEchoRoute, d.DeleteDeviceByName, authenticationHook)
	r.GET(common.ApiDeviceByServiceNameEchoRoute, d.DevicesByServiceName, authenticationHook)
	r.GET(common.ApiDeviceNameExistsEchoRoute, d.DeviceNameExists, authenticationHook)
	r.PATCH(common.ApiDeviceRoute, d.PatchDevice, authenticationHook)
	r.GET(common.ApiAllDeviceRoute, d.AllDevices, authenticationHook)
	r.GET(common.ApiDeviceByNameEchoRoute, d.DeviceByName, authenticationHook)
	r.GET(common.ApiDeviceByProfileNameEchoRoute, d.DevicesByProfileName, authenticationHook)

	// ProvisionWatcher
	pwc := metadataController.NewProvisionWatcherController(dic)
	r.POST(common.ApiProvisionWatcherRoute, pwc.AddProvisionWatcher, authenticationHook)
	r.GET(common.ApiProvisionWatcherByNameEchoRoute, pwc.ProvisionWatcherByName, authenticationHook)
	r.GET(common.ApiProvisionWatcherByServiceNameEchoRoute, pwc.ProvisionWatchersByServiceName, authenticationHook)
	r.GET(common.ApiProvisionWatcherByProfileNameEchoRoute, pwc.ProvisionWatchersByProfileName, authenticationHook)
	r.GET(common.ApiAllProvisionWatcherRoute, pwc.AllProvisionWatchers, authenticationHook)
	r.DELETE(common.ApiProvisionWatcherByNameEchoRoute, pwc.DeleteProvisionWatcherByName, authenticationHook)
	r.PATCH(common.ApiProvisionWatcherRoute, pwc.PatchProvisionWatcher, authenticationHook)
}
