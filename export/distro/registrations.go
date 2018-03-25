//
// Copyright (c) 2017
// Cavium
// Mainflux
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

	export "github.com/edgexfoundry/edgex-go/export"

	"github.com/edgexfoundry/edgex-go/core/domain/models"

	"go.uber.org/zap"
)

var registrationChanges chan export.NotifyUpdate = make(chan export.NotifyUpdate, 2)

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
		reg.format = jsonFormater{}
	case export.FormatXML:
		reg.format = xmlFormater{}
	case export.FormatSerialized:
		// TODO reg.format = distro.NewSerializedFormat()
	case export.FormatIoTCoreJSON:
		// TODO reg.format = distro.NewIotCoreFormat()
	case export.FormatAzureJSON:
		// TODO reg.format = distro.NewAzureFormat()
	case export.FormatCSV:
		// TODO reg.format = distro.NewCsvFormat()
	default:
		logger.Warn("Format not supported: ", zap.String("format", newReg.Format))
		return false
	}

	reg.compression = nil
	switch newReg.Compression {
	case export.CompNone:
		reg.compression = nil
	case export.CompGzip:
		reg.compression = &gzipTransformer{}
	case export.CompZip:
		reg.compression = &zlibTransformer{}
	default:
		logger.Warn("Compression not supported: ", zap.String("compression", newReg.Compression))
		return false
	}

	reg.sender = nil
	switch newReg.Destination {
	case export.DestMQTT:
		reg.sender = NewMqttSender(newReg.Addressable)
	case export.DestZMQ:
		logger.Info("Destination ZMQ is not supported")
	case export.DestIotCoreMQTT:
		// TODO reg.sender = distro.NewIotCoreSender("TODO URL")
	case export.DestAzureMQTT:
		// TODO reg.sender = distro.NewAzureSender("TODO URL")
	case export.DestRest:
		reg.sender = NewHTTPSender(newReg.Addressable)
	default:
		logger.Warn("Destination not supported: ", zap.String("destination", newReg.Destination))
		return false
	}

	if reg.sender == nil {
		return false
	}

	reg.encrypt = nil
	switch newReg.Encryption.Algo {
	case export.EncNone:
		reg.encrypt = nil
	case export.EncAes:
		reg.encrypt = NewAESEncryption(newReg.Encryption)
	default:
		logger.Warn("Encryption not supported: ", zap.String("Algorithm", newReg.Encryption.Algo))
		return false
	}

	reg.filter = nil

	if len(newReg.Filter.DeviceIDs) > 0 {
		reg.filter = append(reg.filter, newDevIdFilter(newReg.Filter))
		logger.Debug("Device ID filter added: ", zap.Any("filters", newReg.Filter.DeviceIDs))
	}

	if len(newReg.Filter.ValueDescriptorIDs) > 0 {
		reg.filter = append(reg.filter, newValueDescFilter(newReg.Filter))
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

	reg.sender.Send(encrypted)
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
func Loop(config Config, errChan chan error, eventCh chan *models.Event) {

	cfg = config

	go func() {
		p := fmt.Sprintf(":%d", cfg.Port)
		logger.Info("Starting Export Distro", zap.String("url", p))
		errChan <- http.ListenAndServe(p, httpServer())
	}()

	registrations := make(map[string]*registrationInfo)

	allRegs := getRegistrations()

	for allRegs == nil {
		logger.Info("Waiting for client microservice")
		select {
		case e := <-errChan:
			logger.Info("exit msg", zap.Error(e))
			return
		case <-time.After(time.Second):
		}
		allRegs = getRegistrations()
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
