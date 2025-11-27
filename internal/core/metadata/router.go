//
// Copyright (C) 2021-2025 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package metadata

import (
	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	metadataController "github.com/edgexfoundry/edgex-go/internal/core/metadata/controller/http"

	"github.com/labstack/echo/v4"
)

func LoadRestRoutes(r *echo.Echo, dic *di.Container, serviceName string) {
	authenticationHook := handlers.AutoConfigAuthenticationFunc(dic)

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
	r.GET(common.ApiDeviceProfileByNameRoute, dc.DeviceProfileByName, authenticationHook)
	r.DELETE(common.ApiDeviceProfileByNameRoute, dc.DeleteDeviceProfileByName, authenticationHook)
	r.GET(common.ApiAllDeviceProfileRoute, dc.AllDeviceProfiles, authenticationHook)
	r.GET(common.ApiDeviceProfileByModelRoute, dc.DeviceProfilesByModel, authenticationHook)
	r.GET(common.ApiDeviceProfileByManufacturerRoute, dc.DeviceProfilesByManufacturer, authenticationHook)
	r.GET(common.ApiDeviceProfileByManufacturerAndModelRoute, dc.DeviceProfilesByManufacturerAndModel, authenticationHook)
	r.PATCH(common.ApiDeviceProfileBasicInfoRoute, dc.PatchDeviceProfileBasicInfo, authenticationHook)
	r.GET(common.ApiAllDeviceProfileBasicInfoRoute, dc.AllDeviceProfileBasicInfos, authenticationHook)
	r.PATCH(common.ApiDeviceProfileTagsByNameRoute, dc.PatchDeviceProfileTags, authenticationHook)

	// Device Resource
	dr := metadataController.NewDeviceResourceController(dic)
	r.GET(common.ApiDeviceResourceByProfileAndResourceRoute, dr.DeviceResourceByProfileNameAndResourceName, authenticationHook)
	r.POST(common.ApiDeviceProfileResourceRoute, dr.AddDeviceProfileResource, authenticationHook)
	r.PATCH(common.ApiDeviceProfileResourceRoute, dr.PatchDeviceProfileResource, authenticationHook)
	r.DELETE(common.ApiDeviceProfileResourceByNameRoute, dr.DeleteDeviceResourceByName, authenticationHook)

	// Deivce Command
	dcm := metadataController.NewDeviceCommandController(dic)
	r.POST(common.ApiDeviceProfileDeviceCommandRoute, dcm.AddDeviceProfileDeviceCommand, authenticationHook)
	r.PATCH(common.ApiDeviceProfileDeviceCommandRoute, dcm.PatchDeviceProfileDeviceCommand, authenticationHook)
	r.DELETE(common.ApiDeviceProfileDeviceCommandByNameRoute, dcm.DeleteDeviceCommandByName, authenticationHook)

	// Device Service
	ds := metadataController.NewDeviceServiceController(dic)
	r.POST(common.ApiDeviceServiceRoute, ds.AddDeviceService, authenticationHook)
	r.PATCH(common.ApiDeviceServiceRoute, ds.PatchDeviceService, authenticationHook)
	r.GET(common.ApiDeviceServiceByNameRoute, ds.DeviceServiceByName, authenticationHook)
	r.DELETE(common.ApiDeviceServiceByNameRoute, ds.DeleteDeviceServiceByName, authenticationHook)
	r.GET(common.ApiAllDeviceServiceRoute, ds.AllDeviceServices, authenticationHook)

	// Device
	d := metadataController.NewDeviceController(dic)
	r.POST(common.ApiDeviceRoute, d.AddDevice, authenticationHook)
	r.DELETE(common.ApiDeviceByNameRoute, d.DeleteDeviceByName, authenticationHook)
	r.GET(common.ApiDeviceByServiceNameRoute, d.DevicesByServiceName, authenticationHook)
	r.GET(common.ApiDeviceNameExistsRoute, d.DeviceNameExists, authenticationHook)
	r.PATCH(common.ApiDeviceRoute, d.PatchDevice, authenticationHook)
	r.GET(common.ApiAllDeviceRoute, d.AllDevices, authenticationHook)
	r.GET(common.ApiDeviceByNameRoute, d.DeviceByName, authenticationHook)
	r.GET(common.ApiDeviceByProfileNameRoute, d.DevicesByProfileName, authenticationHook)

	// ProvisionWatcher
	pwc := metadataController.NewProvisionWatcherController(dic)
	r.POST(common.ApiProvisionWatcherRoute, pwc.AddProvisionWatcher, authenticationHook)
	r.GET(common.ApiProvisionWatcherByNameRoute, pwc.ProvisionWatcherByName, authenticationHook)
	r.GET(common.ApiProvisionWatcherByServiceNameRoute, pwc.ProvisionWatchersByServiceName, authenticationHook)
	r.GET(common.ApiProvisionWatcherByProfileNameRoute, pwc.ProvisionWatchersByProfileName, authenticationHook)
	r.GET(common.ApiAllProvisionWatcherRoute, pwc.AllProvisionWatchers, authenticationHook)
	r.DELETE(common.ApiProvisionWatcherByNameRoute, pwc.DeleteProvisionWatcherByName, authenticationHook)
	r.PATCH(common.ApiProvisionWatcherRoute, pwc.PatchProvisionWatcher, authenticationHook)
}
