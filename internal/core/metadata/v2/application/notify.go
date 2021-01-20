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

func newDeviceServiceCallbackClient(ctx context.Context, dic *di.Container, deviceServiceName string) (interfaces.DeviceServiceCallbackClient, errors.EdgeX) {
	ds, err := DeviceServiceByName(deviceServiceName, ctx, dic)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return v2HttpClient.NewDeviceServiceCallbackClient(ds.BaseAddress), nil
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
