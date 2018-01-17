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
	"syscall"

	"github.com/edgexfoundry/export-go"
	"github.com/edgexfoundry/export-go/distro"

	"go.uber.org/zap"
)

const (
	envClientHost string = "EXPORT_DISTRO_CLIENT_HOST"
	envDataHost   string = "EXPORT_DISTRO_DATA_HOST"
)

var logger *zap.Logger

func main() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()

	distro.InitLogger(logger)

	logger.Info("Starting distro")
	cfg := loadConfig()

	errs := make(chan error, 2)
	eventCh := make(chan *export.Event, 10)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// There can be another receivers that can be initialiced here
	distro.ZeroMQReceiver(eventCh)

	distro.Loop(cfg, errs, eventCh)

	logger.Info("terminated")
}

func loadConfig() distro.Config {
	cfg := distro.GetDefaultConfig()
	cfg.ClientHost = env(envClientHost, cfg.ClientHost)
	cfg.DataHost = env(envDataHost, cfg.DataHost)
	return cfg
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
