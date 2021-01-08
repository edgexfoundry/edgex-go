//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"

	v2MetadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/v2/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// AddProvisionWatcher function accepts the new provision watcher model from the controller function
// and then invokes AddProvisionWatcher function of infrastructure layer to add new device service
func AddProvisionWatcher(pw models.ProvisionWatcher, ctx context.Context, dic *di.Container) (id string, err errors.EdgeX) {
	dbClient := v2MetadataContainer.DBClientFrom(dic.Get)
	lc := container.LoggingClientFrom(dic.Get)
	correlationId := correlation.FromContext(ctx)

	addProvisionWatcher, err := dbClient.AddProvisionWatcher(pw)
	if err != nil {
		return "", errors.NewCommonEdgeXWrapper(err)
	}

	lc.Debugf("ProvisionWatcher created on DB successfully. ProvisionWatcher ID: %s, Correlation-ID: %s ",
		addProvisionWatcher.Id,
		correlationId,
	)

	return addProvisionWatcher.Id, nil
}
