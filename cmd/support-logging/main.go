//
// Copyright (c) 2018
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

	edgex "github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/support/logging"
)

const ()

func main() {
	cfg := loadConfig()

	fmt.Printf("Starting support-logging %s\n", edgex.Version)
	errs := make(chan error, 2)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logging.StartHTTPServer(cfg, errs)

	c := <-errs
	fmt.Println("terminated: ", c)
}

func loadConfig() logging.Config {
	cfg := logging.GetDefaultConfig()
	return cfg
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
