//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/core/keeper/container"
)

func AddRegistration(r models.Registration, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	r, err := dbClient.AddRegistration(r)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	registry := container.RegistryFrom(dic.Get)
	registry.Register(r)

	return nil
}

func UpdateRegistration(r models.Registration, dic *di.Container) errors.EdgeX {
	dbClient := container.DBClientFrom(dic.Get)
	err := dbClient.UpdateRegistration(r)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	registry := container.RegistryFrom(dic.Get)
	// remove the old service health check runner first, and then create a new one based on the updated registry
	registry.DeregisterByServiceId(r.ServiceId)

	if r.Status != models.Halt {
		registry.Register(r)
	}

	return nil
}

func DeleteRegistration(id string, dic *di.Container) errors.EdgeX {
	if id == "" {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "serviceId is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	err := dbClient.DeleteRegistrationByServiceId(id)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	registry := container.RegistryFrom(dic.Get)
	registry.DeregisterByServiceId(id)

	return nil
}

func Registrations(dic *di.Container, deregistered bool) ([]dtos.Registration, errors.EdgeX) {
	dbClient := container.DBClientFrom(dic.Get)
	registrations, err := dbClient.Registrations()
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	var filteredRegistrations []models.Registration
	if deregistered {
		filteredRegistrations = registrations
	} else {
		// if deregistered is false, filter out registrations with status is HALT which have been deregistered
		// otherwise, registrations with all health statuses will be included
		for _, r := range registrations {
			if r.Status != models.Halt {
				filteredRegistrations = append(filteredRegistrations, r)
			}
		}
	}

	res := make([]dtos.Registration, len(filteredRegistrations))
	for idx, r := range filteredRegistrations {
		dto := dtos.FromRegistrationModelToDTO(r)
		res[idx] = dto
	}

	return res, nil
}

func RegistrationByServiceId(id string, dic *di.Container) (dtos.Registration, errors.EdgeX) {
	if id == "" {
		return dtos.Registration{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "serviceId is empty", nil)
	}

	dbClient := container.DBClientFrom(dic.Get)
	r, err := dbClient.RegistrationByServiceId(id)
	if err != nil {
		return dtos.Registration{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, "registration not found by serviceId", err)
	}

	return dtos.FromRegistrationModelToDTO(r), nil
}
