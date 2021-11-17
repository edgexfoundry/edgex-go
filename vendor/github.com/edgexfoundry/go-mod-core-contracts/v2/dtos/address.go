//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

type Address struct {
	Type string `json:"type" validate:"oneof='REST' 'MQTT' 'EMAIL'"`

	Host string `json:"host" validate:"required_unless=Type EMAIL"`
	Port int    `json:"port" validate:"required_unless=Type EMAIL"`

	RESTAddress    `json:",inline" validate:"-"`
	MQTTPubAddress `json:",inline" validate:"-"`
	EmailAddress   `json:",inline" validate:"-"`
}

// Validate satisfies the Validator interface
func (a *Address) Validate() error {
	err := common.Validate(a)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid Address.", err)
	}
	switch a.Type {
	case common.REST:
		err = common.Validate(a.RESTAddress)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid RESTAddress.", err)
		}
		break
	case common.MQTT:
		err = common.Validate(a.MQTTPubAddress)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid MQTTPubAddress.", err)
		}
		break
	case common.EMAIL:
		err = common.Validate(a.EmailAddress)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid EmailAddress.", err)
		}
		break
	}

	return nil
}

type RESTAddress struct {
	Path       string `json:"path,omitempty"`
	HTTPMethod string `json:"httpMethod,omitempty" validate:"required,oneof='GET' 'HEAD' 'POST' 'PUT' 'DELETE' 'TRACE' 'CONNECT'"`
}

func NewRESTAddress(host string, port int, httpMethod string) Address {
	return Address{
		Type: common.REST,
		Host: host,
		Port: port,
		RESTAddress: RESTAddress{
			HTTPMethod: httpMethod,
		},
	}
}

type MQTTPubAddress struct {
	Publisher      string `json:"publisher,omitempty" validate:"required"`
	Topic          string `json:"topic,omitempty" validate:"required"`
	QoS            int    `json:"qos,omitempty"`
	KeepAlive      int    `json:"keepAlive,omitempty"`
	Retained       bool   `json:"retained,omitempty"`
	AutoReconnect  bool   `json:"autoReconnect,omitempty"`
	ConnectTimeout int    `json:"connectTimeout,omitempty"`
}

func NewMQTTAddress(host string, port int, publisher string, topic string) Address {
	return Address{
		Type: common.MQTT,
		Host: host,
		Port: port,
		MQTTPubAddress: MQTTPubAddress{
			Publisher: publisher,
			Topic:     topic,
		},
	}
}

type EmailAddress struct {
	Recipients []string `json:"recipients,omitempty" validate:"gt=0,dive,email"`
}

func NewEmailAddress(recipients []string) Address {
	return Address{
		Type: common.EMAIL,
		EmailAddress: EmailAddress{
			Recipients: recipients,
		},
	}
}

func ToAddressModel(a Address) models.Address {
	var address models.Address

	switch a.Type {
	case common.REST:
		address = models.RESTAddress{
			BaseAddress: models.BaseAddress{
				Type: a.Type, Host: a.Host, Port: a.Port,
			},
			Path:       a.RESTAddress.Path,
			HTTPMethod: a.RESTAddress.HTTPMethod,
		}
		break
	case common.MQTT:
		address = models.MQTTPubAddress{
			BaseAddress: models.BaseAddress{
				Type: a.Type, Host: a.Host, Port: a.Port,
			},
			Publisher:      a.MQTTPubAddress.Publisher,
			Topic:          a.MQTTPubAddress.Topic,
			QoS:            a.QoS,
			KeepAlive:      a.KeepAlive,
			Retained:       a.Retained,
			AutoReconnect:  a.AutoReconnect,
			ConnectTimeout: a.ConnectTimeout,
		}
		break
	case common.EMAIL:
		address = models.EmailAddress{
			BaseAddress: models.BaseAddress{
				Type: a.Type,
			},
			Recipients: a.EmailAddress.Recipients,
		}
	}
	return address
}

func FromAddressModelToDTO(address models.Address) Address {
	dto := Address{
		Type: address.GetBaseAddress().Type,
		Host: address.GetBaseAddress().Host,
		Port: address.GetBaseAddress().Port,
	}

	switch a := address.(type) {
	case models.RESTAddress:
		dto.RESTAddress = RESTAddress{
			Path:       a.Path,
			HTTPMethod: a.HTTPMethod,
		}
		break
	case models.MQTTPubAddress:
		dto.MQTTPubAddress = MQTTPubAddress{
			Publisher:      a.Publisher,
			Topic:          a.Topic,
			QoS:            a.QoS,
			KeepAlive:      a.KeepAlive,
			Retained:       a.Retained,
			AutoReconnect:  a.AutoReconnect,
			ConnectTimeout: a.ConnectTimeout,
		}
		break
	case models.EmailAddress:
		dto.EmailAddress = EmailAddress{
			Recipients: a.Recipients,
		}
		break
	}
	return dto
}

func ToAddressModels(dtos []Address) []models.Address {
	models := make([]models.Address, len(dtos))
	for i, c := range dtos {
		models[i] = ToAddressModel(c)
	}
	return models
}

func FromAddressModelsToDTOs(models []models.Address) []Address {
	dtos := make([]Address, len(models))
	for i, c := range models {
		dtos[i] = FromAddressModelToDTO(c)
	}
	return dtos
}
