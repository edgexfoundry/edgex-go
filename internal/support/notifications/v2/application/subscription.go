//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	v2NotificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/bootstrap/container"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

// The AddSubscription function accepts the new Subscription model from the controller function
// and then invokes AddSubscription function of infrastructure layer to add new Subscription
func AddSubscription(d models.Subscription, ctx context.Context, dic *di.Container) (id string, edgeXerr errors.EdgeX) {
	dbClient := v2NotificationsContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)

	addedSubscription, err := dbClient.AddSubscription(d)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("Subscription created on DB successfully. Subscription ID: %s, Correlation-ID: %s ",
		addedSubscription.Id,
		correlation.FromContext(ctx))

	return addedSubscription.Id, nil
}
