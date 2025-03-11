//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type Registration struct {
	DBTimestamp `json:",inline"`
	ServiceId   string      `json:"serviceId" validate:"required"`
	Status      string      `json:"status"`
	Host        string      `json:"host" validate:"required"`
	Port        int         `json:"port" validate:"required"`
	HealthCheck HealthCheck `json:",inline"`
}

type HealthCheck struct {
	Interval string `json:"interval" validate:"required,edgex-dto-duration"`
	Path     string `json:"path" validate:"required"`
	Type     string `json:"type" validate:"required"`
}

func ToRegistrationModel(dto Registration) models.Registration {
	var r models.Registration
	r.ServiceId = dto.ServiceId
	r.Status = dto.Status
	r.Host = dto.Host
	r.Port = dto.Port
	r.HealthCheck.Type = dto.HealthCheck.Type
	r.HealthCheck.Path = dto.HealthCheck.Path
	r.HealthCheck.Interval = dto.HealthCheck.Interval

	return r
}

func FromRegistrationModelToDTO(r models.Registration) Registration {
	var dto Registration
	dto.DBTimestamp = DBTimestamp(r.DBTimestamp)
	dto.ServiceId = r.ServiceId
	dto.Status = r.Status
	dto.Host = r.Host
	dto.Port = r.Port
	dto.HealthCheck.Type = r.HealthCheck.Type
	dto.HealthCheck.Path = r.HealthCheck.Path
	dto.HealthCheck.Interval = r.HealthCheck.Interval

	return dto
}

func (r *Registration) Validate() error {
	err := common.Validate(r)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid Registration.", err)
	}
	err = common.Validate(r.HealthCheck)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid Registration HealthCheck.", err)
	}
	// check if the health status value is UP, DOWN, UNKNOWN, or HALT
	// if the value is invalid or empty, assign UNKNOWN to the status value
	switch r.Status {
	case models.Up, models.Down, models.Unknown, models.Halt:
		break
	default:
		r.Status = models.Unknown
	}
	return nil
}
