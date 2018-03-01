//
// Copyright (c) 2017
// Cavium
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/edgexfoundry/edgex-go/export/distro"
	consulclient "github.com/edgexfoundry/edgex-go/support/consul-client"

	"go.uber.org/zap"
)

const (
	envDistroHost string = "EXPORT_DISTRO_HOST"
	envClientHost string = "EXPORT_DISTRO_CLIENT_HOST"
	envDataHost   string = "EXPORT_DISTRO_DATA_HOST"
	envConsulHost string = "EXPORT_DISTRO_CONSUL_HOST"
	envConsulPort string = "EXPORT_DISTRO_CONSUL_PORT"
	envMQTTSCert  string = "EXPORT_DISTRO_MQTTS_CERT_FILE"
	envMQTTSKey   string = "EXPORT_DISTRO_MQTTS_KEY_FILE"

	applicationName string = "export-distro"
	consulProfile   string = "go"

	defConsulHost string = "127.0.0.1"
	defConsulPort int    = 8500
	defHostname   string = "127.0.0.1"
)

type config struct {
	ConsulHost string
	ConsulPort int
	Hostname   string
}

var logger *zap.Logger

func main() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting edgex export client", zap.String("version", edgex.Version))

	distro.InitLogger(logger)

	distroCfg, cfg := loadConfig()

	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    applicationName,
		ServicePort:    distroCfg.Port,
		ServiceAddress: cfg.Hostname,
		CheckAddress:   "http://" + cfg.Hostname + ":" + strconv.Itoa(distroCfg.Port) + "/api/v1/ping",
		CheckInterval:  "10s",
		ConsulAddress:  cfg.ConsulHost,
		ConsulPort:     cfg.ConsulPort,
	})

	if err == nil {
		logger.Info("Registered microservice in consul",
			zap.String("consulHost", cfg.ConsulHost),
			zap.Int("consulPort", cfg.ConsulPort))

		consulProfiles := []string{consulProfile}
		if err := consulclient.CheckKeyValuePairs(&distroCfg, applicationName, consulProfiles); err != nil {
			logger.Warn("Error getting key/values from Consul", zap.Error(err),
				zap.String("consulHost", cfg.ConsulHost),
				zap.Int("consulPort", cfg.ConsulPort))
		} else {
			logger.Info("Updated configuration from consul",
				zap.String("consulHost", cfg.ConsulHost),
				zap.Int("consulPort", cfg.ConsulPort))
		}
	} else {
		logger.Warn("Error registering to consul", zap.Error(err),
			zap.String("consulHost", cfg.ConsulHost),
			zap.Int("consulPort", cfg.ConsulPort))
	}

	logger.Info("Starting distro")
	errs := make(chan error, 2)
	eventCh := make(chan *models.Event, 10)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// There can be another receivers that can be initialiced here
	distro.ZeroMQReceiver(eventCh)

	distro.Loop(distroCfg, errs, eventCh)

	logger.Info("terminated")
}

func loadConfig() (distro.Config, config) {
	distroCfg := distro.GetDefaultConfig()

	distroCfg.ClientHost = env(envClientHost, distroCfg.ClientHost)
	distroCfg.DataHost = env(envDataHost, distroCfg.DataHost)
	distroCfg.MQTTSCert = env(envMQTTSCert, distroCfg.MQTTSCert)
	distroCfg.MQTTSKey = env(envMQTTSKey, distroCfg.MQTTSKey)

	cfg := config{
		ConsulHost: env(envConsulHost, defConsulHost),
		ConsulPort: defConsulPort,
		Hostname:   env(envDistroHost, defHostname),
	}
	hostname, err := os.Hostname()
	if err == nil {
		cfg.Hostname = hostname
	}

	portStr := env(envConsulPort, strconv.Itoa(cfg.ConsulPort))
	port, err := strconv.Atoi(portStr)
	if err == nil {
		cfg.ConsulPort = port
	} else {
		logger.Warn("Could not parse port", zap.String("port", portStr))
	}
	return distroCfg, cfg
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
