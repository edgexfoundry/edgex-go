//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"net/http"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	v2HttpClient "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/http"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

func newDeviceServiceCallbackClient(ctx context.Context, dic *di.Container, deviceServiceName string) (interfaces.DeviceServiceCallbackClient, errors.EdgeX) {
	ds, err := DeviceServiceByName(deviceServiceName, ctx, dic)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return v2HttpClient.NewDeviceServiceCallbackClient(ds.BaseAddress), nil
}

// addDeviceCallback invoke device service's callback function for adding new device
func addDeviceCallback(ctx context.Context, dic *di.Container, device dtos.Device) {
	lc := container.LoggingClientFrom(dic.Get)
	deviceServiceCallbackClient, err := newDeviceServiceCallbackClient(ctx, dic, device.ServiceName)
	if err != nil {
		lc.Errorf("fail to new a device service callback client by serviceName %s, err: %v", device.ServiceName, err)
	}
	response, err := deviceServiceCallbackClient.AddDeviceCallback(ctx, requests.AddDeviceRequest{Device: device})
	if err != nil {
		lc.Errorf("fail to invoke device service callback for adding device %s, err: %v", device.Name, err)
	}
	if response.StatusCode != http.StatusOK {
		lc.Errorf("fail to invoke device service callback for adding device %s, err: %s", device.Name, response.Message)
	}
}

// updateDeviceCallback invoke device service's callback function for updating device
func updateDeviceCallback(ctx context.Context, dic *di.Container, serviceName string, device models.Device) {
	lc := container.LoggingClientFrom(dic.Get)
	deviceServiceCallbackClient, err := newDeviceServiceCallbackClient(ctx, dic, serviceName)
	if err != nil {
		lc.Errorf("fail to new a device service callback client by serviceName %s, err: %v", serviceName, err)
	}
	updateDevice := deviceModelToUpdateDTO(device)
	response, err := deviceServiceCallbackClient.UpdateDeviceCallback(ctx, requests.UpdateDeviceRequest{Device: updateDevice})
	if err != nil {
		lc.Errorf("fail to invoke device service callback for updating device %s, err: %v", device.Name, err)
		return
	}
	if response.StatusCode != http.StatusOK {
		lc.Errorf("fail to invoke device service callback for updating device %s, err: %s", device.Name, response.Message)
	}
}

// deleteDeviceCallback invoke device service's callback function for deleting device
func deleteDeviceCallback(ctx context.Context, dic *di.Container, device models.Device) {
	lc := container.LoggingClientFrom(dic.Get)
	deviceServiceCallbackClient, err := newDeviceServiceCallbackClient(ctx, dic, device.ServiceName)
	if err != nil {
		lc.Errorf("fail to new a device service callback client by serviceName %s, err: %v", device.ServiceName, err)
	}
	response, err := deviceServiceCallbackClient.DeleteDeviceCallback(ctx, device.Id)
	if err != nil {
		lc.Errorf("fail to invoke device service callback for deleting device %s, err: %v", device.Name, err)
		return
	}
	if response.StatusCode != http.StatusOK {
		lc.Errorf("fail to invoke device service callback for deleting device %s, err: %s", device.Name, response.Message)
	}
}

func deviceModelToUpdateDTO(device models.Device) dtos.UpdateDevice {
	adminState := string(device.AdminState)
	operatingState := string(device.OperatingState)
	return dtos.UpdateDevice{
		Id:             &device.Id,
		Name:           &device.Name,
		Description:    &device.Description,
		AdminState:     &adminState,
		OperatingState: &operatingState,
		LastConnected:  &device.LastConnected,
		LastReported:   &device.LastReported,
		ServiceName:    &device.ServiceName,
		ProfileName:    &device.ProfileName,
		Labels:         device.Labels,
		Location:       device.Location,
		AutoEvents:     dtos.FromAutoEventModelsToDTOs(device.AutoEvents),
		Protocols:      dtos.FromProtocolModelsToDTOs(device.Protocols),
		Notify:         &device.Notify,
	}
}

// updateDeviceProfileCallback invoke device service's callback function for updating device profile
func updateDeviceProfileCallback(ctx context.Context, dic *di.Container, deviceProfile dtos.DeviceProfile) {
	lc := container.LoggingClientFrom(dic.Get)
	devices, err := DevicesByProfileName(0, -1, deviceProfile.Name, dic)
	if err != nil {
		lc.Errorf("fail to query associated devices by deviceProfile name %s, err: %v", deviceProfile.Name, err)
	}
	// Invoke callback for each device service
	dsMap := make(map[string]bool)
	for _, d := range devices {
		if _, ok := dsMap[d.ServiceName]; ok {
			return
		}
		dsMap[d.ServiceName] = true

		deviceServiceCallbackClient, err := newDeviceServiceCallbackClient(ctx, dic, d.ServiceName)
		if err != nil {
			lc.Errorf("fail to new a device service callback client by serviceName %s, err: %v", d.ServiceName, err)
		}
		response, err := deviceServiceCallbackClient.UpdateDeviceProfileCallback(ctx, requests.DeviceProfileRequest{Profile: deviceProfile})
		if err != nil {
			lc.Errorf("fail to invoke device service callback for updating device profile %s, err: %v", deviceProfile.Name, err)
			return
		}
		if response.StatusCode != http.StatusOK {
			lc.Errorf("fail to invoke device service callback for updating device profile %s, err: %s", deviceProfile.Name, response.Message)
		}
	}
}

// addProvisionWatcherCallback invoke device service's callback function for adding new provision watcher
func addProvisionWatcherCallback(ctx context.Context, dic *di.Container, pw dtos.ProvisionWatcher) {
	lc := container.LoggingClientFrom(dic.Get)
	deviceServiceCallbackClient, err := newDeviceServiceCallbackClient(ctx, dic, pw.ServiceName)
	if err != nil {
		lc.Errorf("fail to new a device service callback client by serviceName %s, err: %v", pw.ServiceName, err)
	}
	response, err := deviceServiceCallbackClient.AddProvisionWatcherCallback(ctx, requests.AddProvisionWatcherRequest{ProvisionWatcher: pw})
	if err != nil {
		lc.Errorf("fail to invoke device service callback for adding  provision watcher %s, err: %v", pw.Name, err)
	}
	if response.StatusCode != http.StatusOK {
		lc.Errorf("fail to invoke device service callback for adding  provision watcher %s, err: %s", pw.Name, response.Message)
	}
}

// updateProvisionWatcherCallback invoke device service's callback function for updating provision watcher
func updateProvisionWatcherCallback(ctx context.Context, dic *di.Container, serviceName string, pw models.ProvisionWatcher) {
	lc := container.LoggingClientFrom(dic.Get)
	deviceServiceCallbackClient, err := newDeviceServiceCallbackClient(ctx, dic, serviceName)
	if err != nil {
		lc.Errorf("fail to new a device service callback client by serviceName %s, err: %v", serviceName, err)
	}
	dto := dtos.FromProvisionWatcherModelToUpdateDTO(pw)
	response, err := deviceServiceCallbackClient.UpdateProvisionWatcherCallback(ctx, requests.UpdateProvisionWatcherRequest{ProvisionWatcher: dto})
	if err != nil {
		lc.Errorf("fail to invoke device service callback for updating provision watcher %s, err: %v", pw.Name, err)
		return
	}
	if response.StatusCode != http.StatusOK {
		lc.Errorf("fail to invoke device service callback for updating provision watcher %s, err: %s", pw.Name, response.Message)
	}
}

// deleteProvisionWatcherCallback invoke device service's callback function for deleting provision watcher
func deleteProvisionWatcherCallback(ctx context.Context, dic *di.Container, pw models.ProvisionWatcher) {
	lc := container.LoggingClientFrom(dic.Get)
	deviceServiceCallbackClient, err := newDeviceServiceCallbackClient(ctx, dic, pw.ServiceName)
	if err != nil {
		lc.Errorf("fail to new a device service callback client by serviceName %s, err: %v", pw.ServiceName, err)
	}
	response, err := deviceServiceCallbackClient.DeleteProvisionWatcherCallback(ctx, pw.Id)
	if err != nil {
		lc.Errorf("fail to invoke device service callback for deleting provision watcher %s, err: %v", pw.Name, err)
		return
	}
	if response.StatusCode != http.StatusOK {
		lc.Errorf("fail to invoke device service callback for deleting provision watcher %s, err: %s", pw.Name, response.Message)
	}
}

// updateDeviceServiceCallback invoke device service's callback function for updating device service
func updateDeviceServiceCallback(ctx context.Context, dic *di.Container, ds models.DeviceService) {
	lc := container.LoggingClientFrom(dic.Get)
	deviceServiceCallbackClient, err := newDeviceServiceCallbackClient(ctx, dic, ds.Name)
	if err != nil {
		lc.Errorf("fail to new a device service callback client by serviceName %s, err: %v", ds.Name, err)
	}
	dto := deviceServiceModelToUpdateDTO(ds)
	response, err := deviceServiceCallbackClient.UpdateDeviceServiceCallback(ctx, requests.UpdateDeviceServiceRequest{Service: dto})
	if err != nil {
		lc.Errorf("fail to invoke device service callback for updating device service %s, err: %v", ds.Name, err)
	}
	if response.StatusCode != http.StatusOK {
		lc.Errorf("fail to invoke device service callback for updating device service %s, err: %s", ds.Name, response.Message)
	}
}

func deviceServiceModelToUpdateDTO(ds models.DeviceService) dtos.UpdateDeviceService {
	adminState := string(ds.AdminState)
	return dtos.UpdateDeviceService{
		Id:          &ds.Id,
		Name:        &ds.Name,
		BaseAddress: &ds.BaseAddress,
		Labels:      ds.Labels,
		AdminState:  &adminState,
	}
}
