//
// Copyright (C) 2021-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package dtos

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type Address struct {
	Type string `json:"type" validate:"oneof='REST' 'MQTT' 'EMAIL' 'ZeroMQ'"`

	Scheme string `json:"scheme,omitempty"`
	Host   string `json:"host,omitempty" validate:"required_unless=Type EMAIL"`
	Port   int    `json:"port,omitempty" validate:"required_unless=Type EMAIL"`

	RESTAddress    `json:",inline" validate:"-"`
	MQTTPubAddress `json:",inline" validate:"-"`
	EmailAddress   `json:",inline" validate:"-"`
	ZeroMQAddress  `json:",inline" validate:"-"`
	MessageBus     `json:",inline" validate:"-"`
	Security       `json:",inline" validate:"-"`
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
	case common.MQTT:
		err = common.Validate(a.MQTTPubAddress)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid MQTTPubAddress.", err)
		}
		err = common.Validate(a.MessageBus)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid MQTTPubAddress.", err)
		}
	case common.ZeroMQ:
		err = common.Validate(a.MessageBus)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid ZeroMQAddress.", err)
		}
	case common.EMAIL:
		err = common.Validate(a.EmailAddress)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid EmailAddress.", err)
		}
	}

	return nil
}

type RESTAddress struct {
	Path            string `json:"path,omitempty"`
	HTTPMethod      string `json:"httpMethod,omitempty" validate:"required,oneof='GET' 'HEAD' 'POST' 'PUT' 'PATCH' 'DELETE' 'TRACE' 'CONNECT'"`
	InjectEdgeXAuth bool   `json:"injectEdgeXAuth,omitempty"`
}

func NewRESTAddress(host string, port int, httpMethod string, scheme string) Address {
	if scheme == "" {
		scheme = common.HTTP
	}
	return Address{
		Type:   common.REST,
		Host:   host,
		Port:   port,
		Scheme: scheme,
		RESTAddress: RESTAddress{
			HTTPMethod: httpMethod,
		},
	}
}

type MQTTPubAddress struct {
	Publisher      string `json:"publisher,omitempty" validate:"required"`
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
		},
		MessageBus: MessageBus{Topic: topic},
	}
}

func NewMQTTAddressWithSecurity(scheme string, host string, port int, publisher string, topic string, authMode string, secretPath string, skipCertVerify bool) Address {
	return Address{
		Type:   common.MQTT,
		Scheme: scheme,
		Host:   host,
		Port:   port,
		MQTTPubAddress: MQTTPubAddress{
			Publisher: publisher,
		},
		MessageBus: MessageBus{Topic: topic},
		Security: Security{
			AuthMode:       authMode,
			SecretPath:     secretPath,
			SkipCertVerify: skipCertVerify,
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

type MessageBus struct {
	Topic string `json:"topic,omitempty" validate:"required"`
}

type Security struct {
	SecretPath     string `json:"secretPath,omitempty" validate:"required"`
	AuthMode       string `json:"authMode,omitempty" validate:"required,oneof='none' 'usernamepassword' 'cacert' 'clientcert'"`
	SkipCertVerify bool   `json:"skipCertVerify,omitempty"`
}

type ZeroMQAddress struct {
}

func NewZeroMQAddress(host string, port int, topic string) Address {
	return Address{
		Type:       common.ZeroMQ,
		Host:       host,
		Port:       port,
		MessageBus: MessageBus{Topic: topic},
	}
}

func ToAddressModel(a Address) models.Address {
	var address models.Address

	switch a.Type {
	case common.REST:
		address = models.RESTAddress{
			BaseAddress: models.BaseAddress{
				Type: a.Type, Host: a.Host, Port: a.Port, Scheme: a.Scheme,
			},
			Path:            a.RESTAddress.Path,
			HTTPMethod:      a.RESTAddress.HTTPMethod,
			InjectEdgeXAuth: a.RESTAddress.InjectEdgeXAuth,
		}
	case common.MQTT:
		address = models.MQTTPubAddress{
			BaseAddress: models.BaseAddress{
				Type: a.Type, Scheme: a.Scheme, Host: a.Host, Port: a.Port,
			},
			Security: models.Security{
				SecretPath:     a.SecretPath,
				AuthMode:       a.AuthMode,
				SkipCertVerify: a.SkipCertVerify,
			},
			MessageBus:     models.MessageBus{Topic: a.Topic},
			Publisher:      a.MQTTPubAddress.Publisher,
			QoS:            a.QoS,
			KeepAlive:      a.KeepAlive,
			Retained:       a.Retained,
			AutoReconnect:  a.AutoReconnect,
			ConnectTimeout: a.ConnectTimeout,
		}
	case common.ZeroMQ:
		address = models.ZeroMQAddress{
			BaseAddress: models.BaseAddress{
				Type: a.Type, Host: a.Host, Port: a.Port,
			},
			MessageBus: models.MessageBus{Topic: a.Topic},
		}
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
		Type:   address.GetBaseAddress().Type,
		Scheme: address.GetBaseAddress().Scheme,
		Host:   address.GetBaseAddress().Host,
		Port:   address.GetBaseAddress().Port,
	}

	switch a := address.(type) {
	case models.RESTAddress:
		dto.RESTAddress = RESTAddress{
			Path:            a.Path,
			HTTPMethod:      a.HTTPMethod,
			InjectEdgeXAuth: a.InjectEdgeXAuth,
		}
	case models.MQTTPubAddress:
		dto.MQTTPubAddress = MQTTPubAddress{
			Publisher:      a.Publisher,
			QoS:            a.QoS,
			KeepAlive:      a.KeepAlive,
			Retained:       a.Retained,
			AutoReconnect:  a.AutoReconnect,
			ConnectTimeout: a.ConnectTimeout,
		}
		dto.MessageBus = MessageBus{Topic: a.Topic}
		dto.Security = Security{
			SecretPath:     a.SecretPath,
			AuthMode:       a.AuthMode,
			SkipCertVerify: a.SkipCertVerify,
		}
	case models.ZeroMQAddress:
		dto.MessageBus = MessageBus{Topic: a.Topic}
	case models.EmailAddress:
		dto.EmailAddress = EmailAddress{
			Recipients: a.Recipients,
		}
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
