//
// Copyright (c) 2017
// Cavium
//
// SPDX-License-Identifier: Apache-2.0

package distro

import (
	"encoding/json"
	"encoding/xml"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"go.uber.org/zap"
)

type jsonFormater struct {
}

func (jsonTr jsonFormater) Format(event *models.Event) []byte {

	b, err := json.Marshal(event)
	if err != nil {
		logger.Error("Error parsing JSON", zap.Error(err))
		return nil
	}
	return b
}

type xmlFormater struct {
}

func (xmlTr xmlFormater) Format(event *models.Event) []byte {
	b, err := xml.Marshal(event)
	if err != nil {
		logger.Error("Error parsing XML", zap.Error(err))
		return nil
	}
	return b
}
