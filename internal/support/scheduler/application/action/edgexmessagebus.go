//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"context"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

func publishEdgeXMessageBus(dic *di.Container, action models.EdgeXMessageBusAction) errors.EdgeX {
	messageBus := bootstrapContainer.MessagingClientFrom(dic.Get)
	contentType := action.ContentType
	if contentType == "" {
		contentType = common.ContentTypeJSON
	}
	ctx := context.WithValue(context.Background(), common.ContentType, contentType) //nolint: staticcheck

	envelope := types.NewMessageEnvelope(action.Payload, ctx)

	if err := messageBus.Publish(envelope, action.Topic); err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to publish to EdgeX message bus", err)
	}

	return nil
}
