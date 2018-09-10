//
// Copyright (c) 2017
// Cavium
// Mainflux
// IOTech
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

// TODO:
// - Event buffer management per sender(do not block distro.Loop on full
//   registration channel)

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/internal/export/interfaces"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"net/http"
	"time"
)

const (
	awsMQTTPort         int    = 8883
	awsThingUpdateTopic string = "$aws/things/%s/shadow/update"
)

var registrationChanges chan export.NotifyUpdate = make(chan export.NotifyUpdate, 2)

// RegistrationInfo - registration info
type registrationInfo struct {
	registration export.Registration
	format       interfaces.Formatter
	compression  interfaces.Transformer
	encrypt      interfaces.Transformer
	sender       interfaces.Sender
	filter       []interfaces.Filterer

	chRegistration chan *export.Registration
	chEvent        chan *models.Event

	deleteFlag bool
}

func RefreshRegistrations(update export.NotifyUpdate) {
	// TODO make it not blocking, return bool?
	registrationChanges <- update
}

func newRegistrationInfo() *registrationInfo {
	reg := &registrationInfo{}

	reg.chRegistration = make(chan *export.Registration)
	reg.chEvent = make(chan *models.Event)
	return reg
}

func (reg *registrationInfo) update(newReg export.Registration) bool {
	reg.registration = newReg

	reg.format = nil
	switch newReg.Format {
	case export.FormatJSON:
		reg.format = jsonFormatter{}
	case export.FormatXML:
		reg.format = xmlFormatter{}
	case export.FormatSerialized:
		// TODO reg.format = distro.NewSerializedFormat()
	case export.FormatIoTCoreJSON:
		reg.format = jsonFormatter{}
	case export.FormatAzureJSON:
		reg.format = azureFormatter{}
	case export.FormatAWSJSON:
		reg.format = awsFormatter{}
	case export.FormatCSV:
		// TODO reg.format = distro.NewCsvFormat()
	case export.FormatThingsBoardJSON:
		reg.format = thingsboardJSONFormatter{}
	case export.FormatNOOP:
		reg.format = noopFormatter{}
	default:
		LoggingClient.Warn(fmt.Sprintf("Format not supported: %s", newReg.Format))
		return false
	}

	reg.compression = nil
	switch newReg.Compression {
	case "":
		fallthrough
	case export.CompNone:
		reg.compression = nil
	case export.CompGzip:
		reg.compression = &gzipTransformer{}
	case export.CompZip:
		reg.compression = &zlibTransformer{}
	default:
		LoggingClient.Warn(fmt.Sprintf("Compression not supported: %s", newReg.Compression))
		return false
	}

	reg.sender = nil
	switch newReg.Destination {
	case export.DestMQTT, export.DestAzureMQTT:
		reg.sender = NewMqttSender(newReg.Addressable, configuration.MQTTSCert, configuration.MQTTSKey)
	case export.DestAWSMQTT:
		newReg.Addressable.Protocol = "tls"
		newReg.Addressable.Path = ""
		newReg.Addressable.Topic = fmt.Sprintf(awsThingUpdateTopic, newReg.Addressable.Topic)
		newReg.Addressable.Port = awsMQTTPort
		reg.sender = NewMqttSender(newReg.Addressable, configuration.AWSCert, configuration.AWSKey)
	case export.DestZMQ:
		LoggingClient.Info("Destination ZMQ is not supported")
	case export.DestIotCoreMQTT:
		reg.sender = NewIoTCoreSender(newReg.Addressable)
	case export.DestRest:
		reg.sender = NewHTTPSender(newReg.Addressable)
	case export.DestXMPP:
		reg.sender = NewXMPPSender(newReg.Addressable)
	case export.DestInfluxDB:
		reg.sender = NewInfluxDBSender(newReg.Addressable)

	default:
		LoggingClient.Warn(fmt.Sprintf("Destination not supported: %s", newReg.Destination))
		return false
	}

	if reg.sender == nil {
		return false
	}

	reg.encrypt = nil
	switch newReg.Encryption.Algo {
	case "":
		fallthrough
	case export.EncNone:
		reg.encrypt = nil
	case export.EncAes:
		reg.encrypt = export.NewAESEncryption(newReg.Encryption)
	default:
		LoggingClient.Warn(fmt.Sprintf("Encryption not supported: %s", newReg.Encryption.Algo))
		return false
	}

	reg.filter = nil

	if len(newReg.Filter.DeviceIDs) > 0 {
		reg.filter = append(reg.filter, NewDevIdFilter(newReg.Filter))
		LoggingClient.Debug(fmt.Sprintf("Device ID filter added: %s", newReg.Filter.DeviceIDs))
	}

	if len(newReg.Filter.ValueDescriptorIDs) > 0 {
		reg.filter = append(reg.filter, NewValueDescFilter(newReg.Filter))
		LoggingClient.Debug(fmt.Sprintf("Value descriptor filter added: %s", newReg.Filter.ValueDescriptorIDs))
	}

	return true
}

func (reg registrationInfo) processEvent(event *models.Event) {
	// Valid Event Filter, needed?

	for _, f := range reg.filter {
		var accepted bool
		accepted, event = f.Filter(event)
		if !accepted {
			LoggingClient.Info("Event filtered")
			return
		}
	}

	if reg.format == nil {
		LoggingClient.Warn("registrationInfo with nil format")
		return
	}
	formated := reg.format.Format(event)

	compressed := formated
	if reg.compression != nil {
		compressed = reg.compression.Transform(formated)
	}

	encrypted := compressed
	if reg.encrypt != nil {
		encrypted = reg.encrypt.Transform(compressed)
	}

	if reg.sender.Send(encrypted, event) && configuration.MarkPushed {
		id := event.ID.Hex()
		err := ec.MarkPushed(id)

		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Failed to mark event as pushed : event ID = %s: %s", id, err))
		}
	}

	LoggingClient.Debug(fmt.Sprintf("Sent event with registration: %s", reg.registration.Name))
}

func registrationLoop(reg *registrationInfo) {
	LoggingClient.Info(fmt.Sprintf("registration loop started: %s", reg.registration.Name))
	for {
		select {
		case event := <-reg.chEvent:
			reg.processEvent(event)

		case newReg := <-reg.chRegistration:
			if newReg == nil {
				LoggingClient.Info("Terminating registration goroutine")
				return
			} else {
				if reg.update(*newReg) {
					LoggingClient.Info(fmt.Sprintf("Registration %s updated: OK", reg.registration.Name))
				} else {
					LoggingClient.Info(fmt.Sprintf("Registration %s updated: OK, terminating goroutine", reg.registration.Name))
					reg.deleteFlag = true
					return
				}
			}
		}
	}
}

func updateRunningRegistrations(running map[string]*registrationInfo,
	update export.NotifyUpdate) error {

	switch update.Operation {
	case export.NotifyUpdateDelete:
		for k, v := range running {
			if k == update.Name {
				v.chRegistration <- nil
				delete(running, k)
				return nil
			}
		}
		return fmt.Errorf("delete update not processed")
	case export.NotifyUpdateUpdate:
		reg := getRegistrationByName(update.Name)
		if reg == nil {
			return fmt.Errorf("Could not find registration")
		}
		for k, v := range running {
			if k == update.Name {
				v.chRegistration <- reg
				return nil
			}
		}
		return fmt.Errorf("Could not find running registration")
	case export.NotifyUpdateAdd:
		reg := getRegistrationByName(update.Name)
		if reg == nil {
			return fmt.Errorf("Could not find registration")
		}
		regInfo := newRegistrationInfo()
		if regInfo.update(*reg) {
			running[reg.Name] = regInfo
			go registrationLoop(regInfo)
		}
		return nil
	default:
		return fmt.Errorf("Invalid update operation")
	}
}

// Loop - registration loop
func Loop(errChan chan error, eventCh chan *models.Event) {
	go func() {
		p := fmt.Sprintf(":%d", configuration.Port)
		LoggingClient.Info(fmt.Sprintf("Starting Export Distro %s", p))
		errChan <- http.ListenAndServe(p, httpServer())
	}()

	registrations := make(map[string]*registrationInfo)

	allRegs, err := getRegistrations()

	for allRegs == nil {
		LoggingClient.Info("Waiting for client microservice")
		select {
		case e := <-errChan:
			LoggingClient.Error(fmt.Sprintf("exit msg: %s", e.Error()))
			if err != nil {
				LoggingClient.Error(fmt.Sprintf("with error: %s",  err.Error()))
			}
			return
		case <-time.After(time.Second):
		}
		allRegs, err = getRegistrations()
	}

	// Create new goroutines for each registration
	for _, reg := range allRegs {
		regInfo := newRegistrationInfo()
		if regInfo.update(reg) {
			registrations[reg.Name] = regInfo
			go registrationLoop(regInfo)
		}
	}

	LoggingClient.Info("Starting registration loop")
	for {
		select {
		case e := <-errChan:
			// kill all registration goroutines
			for k, reg := range registrations {
				if !reg.deleteFlag {
					// Do not write in channel that will not be read
					reg.chRegistration <- nil
				}
				delete(registrations, k)
			}
			LoggingClient.Error(fmt.Sprintf("exit msg: %s", e.Error()))
			return

		case update := <-registrationChanges:
			LoggingClient.Info("Registration changes")
			err := updateRunningRegistrations(registrations, update)
			if err != nil {
				LoggingClient.Error(err.Error())
				LoggingClient.Warn(fmt.Sprintf("Error updating registration %s", update.Name))
			}

		case event := <-eventCh:
			LoggingClient.Info("EVENT")
			for k, reg := range registrations {
				if reg.deleteFlag {
					delete(registrations, k)
				} else {
					// TODO only sent event if it is not blocking
					reg.chEvent <- event
				}
			}
		}
	}
}
