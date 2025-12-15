//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/edgex-go/internal/core/keeper/container"
)

type Registry struct {
	ctx   context.Context
	wg    *sync.WaitGroup
	dic   *di.Container
	mutex sync.Mutex
	table map[string]*healthCheckRunner
}

func NewRegistry(ctx context.Context, wg *sync.WaitGroup, dic *di.Container) *Registry {
	return &Registry{
		ctx:   ctx,
		wg:    wg,
		dic:   dic,
		table: make(map[string]*healthCheckRunner),
	}
}

func (r *Registry) Register(registration models.Registration) {
	lc := bootstrapContainer.LoggingClientFrom(r.dic.Get)
	runner := newHealthCheckRunner(registration, r.dic)
	r.mutex.Lock()
	r.table[registration.ServiceId] = runner
	r.mutex.Unlock()
	go runner.start(r.ctx, r.wg)

	lc.Infof("Registered service: %s", registration.ServiceId)
}

func (r *Registry) DeregisterByServiceId(id string) {
	if _, ok := r.table[id]; !ok {
		// service has already been deregistered
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	runner := r.table[id]
	runner.stop()
	delete(r.table, id)
}

type healthCheckRunner struct {
	done     chan struct{}
	registry models.Registration
	dic      *di.Container
}

func newHealthCheckRunner(r models.Registration, dic *di.Container) *healthCheckRunner {
	return &healthCheckRunner{
		done:     make(chan struct{}, 1),
		registry: r,
		dic:      dic,
	}
}

func (h *healthCheckRunner) start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	dbClient := container.DBClientFrom(h.dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(h.dic.Get)
	duration, _ := time.ParseDuration(h.registry.HealthCheck.Interval)
	// set a ticker using the 1/2 health check interval
	preSvcUpTicker := time.NewTicker(duration / 2)

	configuration := container.ConfigurationFrom(h.dic.Get)
	reqTimeout, err := time.ParseDuration(configuration.Service.RequestTimeout)
	if err != nil {
		lc.Errorf("Unable to parse RequestTimeout value of '%s' duration: %v", configuration.Service.RequestTimeout, err)
	}

	// use 1/2 health check interval to check the service health repeatedly before the status is UP
preServiceUpLoop:
	for {
		select {
		case <-ctx.Done():
			preSvcUpTicker.Stop()
			return
		case <-h.done:
			preSvcUpTicker.Stop()
			lc.Infof("Deregistered service: %s", h.registry.ServiceId)
			return
		case <-preSvcUpTicker.C:
			h.registry.Status = healthCheck(h.registry, lc, reqTimeout)
			err := dbClient.UpdateRegistration(h.registry)
			if err != nil {
				lc.Error("Failed to update health check status for %s: %s", h.registry.ServiceId, err.Error())
				continue
			}

			if h.registry.Status == models.Up {
				break preServiceUpLoop
			}
		}
	}

	preSvcUpTicker.Stop()
	// use the full health check interval to check service status periodically
	ticker := time.NewTicker(duration)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-h.done:
			ticker.Stop()
			lc.Infof("Deregistered service: %s", h.registry.ServiceId)
			return
		case <-ticker.C:
			h.registry.Status = healthCheck(h.registry, lc, reqTimeout)
			err := dbClient.UpdateRegistration(h.registry)
			if err != nil {
				lc.Error("Failed to update health check status for %s: %s", h.registry.ServiceId, err.Error())
			}
		}
	}
}

func (h *healthCheckRunner) stop() {
	close(h.done)
}

func healthCheck(r models.Registration, lc logger.LoggingClient, timeout time.Duration) string {
	client := http.Client{
		Timeout: timeout,
	}
	path := r.HealthCheck.Type + "://" + r.Host + ":" + strconv.Itoa(r.Port) + r.HealthCheck.Path
	req, err := http.NewRequest(http.MethodGet, path, http.NoBody)
	if err != nil {
		lc.Errorf("failed to create get request for %s: %v", path, err)
		return models.Down
	}

	resp, err := client.Do(req)
	if err != nil {
		lc.Errorf("Failed to health check service %s: %s", r.ServiceId, err.Error())
		return models.Down
	}

	// Ensure response body is always closed to prevent resource leaks
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		lc.Debugf("service %s status healthy", r.ServiceId)
		return models.Up
	} else {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			lc.Error("Failed to read %s response body: %s", path, err.Error())
		}
		lc.Errorf("service %s is unhealthy: %s", r.ServiceId, string(bodyBytes))
		return models.Down
	}
}
