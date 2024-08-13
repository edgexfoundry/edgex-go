//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"context"
	"encoding/json"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

func issueSetCommand(dic *di.Container, action models.DeviceControlAction) (string, errors.EdgeX) {
	if action.DeviceName == "" {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "device name cannot be empty", nil)
	}

	if action.SourceName == "" {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "source name cannot be empty", nil)
	}

	var payload map[string]any
	if err := json.Unmarshal(action.Payload, &payload); err != nil {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to convert payload to map", err)
	}

	cc := bootstrapContainer.CommandClientFrom(dic.Get)
	if cc == nil {
		return "", errors.NewCommonEdgeX(errors.KindServerError, "nil CommandClient returned", nil)
	}

	resp, err := cc.IssueSetCommandByNameWithObject(context.Background(), action.DeviceName, action.SourceName, payload)
	if err != nil {
		return "", err
	}

	return resp.Message, nil
}
