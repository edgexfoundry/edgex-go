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
	"net/http"
	"time"

	"go.uber.org/zap"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

var registrationChanges chan models.NotifyUpdate = make(chan models.NotifyUpdate, 2)

// RegistrationInfo - registration info
type registrationInfo struct {
	registration models.Registration
	format       models.Formatter
	compression  models.Transformer
	encrypt      models.Transformer
	sender       models.Sender
	filter       []models.Filterer

	chRegistration chan *models.Registration
	chEvent        chan *models.Event

	deleteMe bool
}

func RefreshRegistrations(update models.NotifyUpdate) {
	// TODO make it not blocking, return bool?
	registrationChanges <- update
}

func newRegistrationInfo() *registrationInfo {
	reg := &registrationInfo{}

	reg.chRegistration = make(chan *models.Registration)
	reg.chEvent = make(chan *models.Event)
	return reg
}

func (reg *registrationInfo) update(newReg models.Registration) bool {
	reg.registration = newReg

	reg.format = nil
	switch newReg.Format {
	case models.FormatJSON:
		reg.format = jsonFormatter{}
	case models.FormatXML:
		reg.format = xmlFormatter{}
	case models.FormatSerialized:
		// TODO reg.format = distro.NewSerializedFormat()
	case models.FormatIoTCoreJSON:
		reg.format = jsonFormatter{}
	case models.FormatAzureJSON:
		reg.format = azureFormatter{}
	case models.FormatCSV:
		// TODO reg.format = distro.NewCsvFormat()
	case models.FormatThingsBoardJSON:
		reg.format = thingsboardJSONFormatter{}
	case models.FormatNOOP:
		reg.format = noopFormatter{}
	default:
		logger.Warn("Format not supported: ", zap.String("format", newReg.Format))
		return false
	}

	reg.compression = nil
	switch newReg.Compression {
	case models.CompNone:
		reg.compression = nil
	case models.CompGzip:
		reg.compression = &gzipTransformer{}
	case models.CompZip:
		reg.compression = &zlibTransformer{}
	default:
		logger.Warn("Compression not supported: ", zap.String("compression", newReg.Compression))
		return false
	}

	reg.sender = nil
	switch newReg.Destination {
	case models.DestMQTT, models.DestAzureMQTT:
		reg.sender = NewMqttSender(newReg.Addressable)
	case models.DestZMQ:
		logger.Info("Destination ZMQ is not supported")
	case models.DestIotCoreMQTT:
		reg.sender = NewIoTCoreSender(newReg.Addressable)
	case models.DestRest:
		reg.sender = models.NewHTTPSender(newReg.Addressable)
	case models.DestXMPP:
		reg.sender = models.NewXMPPSender(newReg.Addressable)
	case models.DestInfluxDB:
		reg.sender = models.NewInfluxDBSender(newReg.Addressable)

	default:
		logger.Warn("Destination not supported: ", zap.String("destination", newReg.Destination))
		return false
	}

	if reg.sender == nil {
		return false
	}

	reg.encrypt = nil
	switch newReg.Encryption.Algo {
	case models.EncNone:
		reg.encrypt = nil
	case models.EncAes:
		reg.encrypt = models.NewAESEncryption(newReg.Encryption)
	default:
		logger.Warn("Encryption not supported: ", zap.String("Algorithm", newReg.Encryption.Algo))
		return false
	}

	reg.filter = nil

	if len(newReg.Filter.DeviceIDs) > 0 {
		reg.filter = append(reg.filter, models.NewDevIdFilter(newReg.Filter))
		logger.Debug("Device ID filter added: ", zap.Any("filters", newReg.Filter.DeviceIDs))
	}

	if len(newReg.Filter.ValueDescriptorIDs) > 0 {
		reg.filter = append(reg.filter, models.NewValueDescFilter(newReg.Filter))
		logger.Debug("Value descriptor filter added: ", zap.Any("filters", newReg.Filter.ValueDescriptorIDs))
	}

	return true
}

func (reg registrationInfo) processEvent(event *models.Event) {
	// Valid Event Filter, needed?

	for _, f := range reg.filter {
		var accepted bool
		accepted, event = f.Filter(event)
		if !accepted {
			logger.Info("Event filtered")
			return
		}
	}

	if reg.format == nil {
		logger.Warn("registrationInfo with nil format")
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
			logger.Error(fmt.Sprintf("Failed to mark event as pushed : event ID = %s: %s", id, err))
		}
	}

	logger.Debug("Sent event with registration:",
		zap.Any("Event", event),
		zap.String("Name", reg.registration.Name))
}

func registrationLoop(reg *registrationInfo) {
	logger.Info("registration loop started",
		zap.String("Name", reg.registration.Name))
	for {
		select {
		case event := <-reg.chEvent:
			reg.processEvent(event)

		case newReg := <-reg.chRegistration:
			if newReg == nil {
				logger.Info("Terminating registration goroutine")
				return
			} else {
				if reg.update(*newReg) {
					logger.Info("Registration updated: OK",
						zap.String("Name", reg.registration.Name))
				} else {
					logger.Info("Registration updated: KO, terminating goroutine",
						zap.String("Name", reg.registration.Name))
					reg.deleteMe = true
					return
				}
			}
		}
	}
}

func updateRunningRegistrations(running map[string]*registrationInfo,
	update models.NotifyUpdate) error {

	switch update.Operation {
	case models.NotifyUpdateDelete:
		for k, v := range running {
			if k == update.Name {
				v.chRegistration <- nil
				delete(running, k)
				return nil
			}
		}
		return fmt.Errorf("delete update not processed")
	case models.NotifyUpdateUpdate:
		reg := GetRegistrationByName(update.Name)
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
	case models.NotifyUpdateAdd:
		reg := GetRegistrationByName(update.Name)
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
		logger.Info("Starting Export Distro", zap.String("url", p))
		errChan <- http.ListenAndServe(p, httpServer())
	}()

	registrations := make(map[string]*registrationInfo)

	allRegs := GetRegistrations()

	for allRegs == nil {
		logger.Info("Waiting for client microservice")
		select {
		case e := <-errChan:
			logger.Info("exit msg", zap.Error(e))
			return
		case <-time.After(time.Second):
		}
		allRegs = GetRegistrations()
	}

	// Create new goroutines for each registration
	for _, reg := range allRegs {
		regInfo := newRegistrationInfo()
		if regInfo.update(reg) {
			registrations[reg.Name] = regInfo
			go registrationLoop(regInfo)
		}
	}

	logger.Info("Starting registration loop")
	for {
		select {
		case e := <-errChan:
			// kill all registration goroutines
			for k, reg := range registrations {
				if !reg.deleteMe {
					// Do not write in channel that will not be read
					reg.chRegistration <- nil
				}
				delete(registrations, k)
			}
			logger.Info("exit msg", zap.Error(e))
			return

		case update := <-registrationChanges:
			logger.Info("Registration changes")
			err := updateRunningRegistrations(registrations, update)
			if err != nil {
				logger.Warn("Error updating registration", zap.Error(err),
					zap.Any("update", update))
			}

		case event := <-eventCh:
			logger.Info("EVENT")
			for k, reg := range registrations {
				if reg.deleteMe {
					delete(registrations, k)
				} else {
					// TODO only sent event if it is not blocking
					reg.chEvent <- event
				}
			}
		}
	}
}
