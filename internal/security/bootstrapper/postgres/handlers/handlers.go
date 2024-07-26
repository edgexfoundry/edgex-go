//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/security/bootstrapper/postgres/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v3/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"
)

// Handler is the redis bootstrapping handler
type Handler struct {
	credentials bootstrapConfig.Credentials
}

// NewHandler instantiates a new Handler
func NewHandler() *Handler {
	return &Handler{}
}

// GetCredentials retrieves the postgres database credentials from secretstore
func (handler *Handler) GetCredentials(ctx context.Context, _ *sync.WaitGroup, startupTimer startup.Timer,
	dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)

	var credentials = bootstrapConfig.Credentials{
		Username: "unset",
		Password: "unset",
	}

	for startupTimer.HasNotElapsed() {
		// retrieve database credentials from secretstore
		secrets, err := secretProvider.GetSecret(config.Database.Type)
		if err == nil {
			credentials.Username = secrets[secret.UsernameKey]
			credentials.Password = secrets[secret.PasswordKey]
			break
		}

		lc.Warnf("Could not retrieve database credentials (startup timer has not expired): %s", err.Error())
		startupTimer.SleepForInterval()
	}

	if credentials.Password == "unset" {
		lc.Error("Failed to retrieve database credentials before startup timer expired")
		return false
	}

	handler.credentials = credentials
	return true
}
