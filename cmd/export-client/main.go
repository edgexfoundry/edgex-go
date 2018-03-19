//
// Copyright (c) 2017
// Mainflux
// Cavium
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
	"time"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/export/client"
	"github.com/edgexfoundry/edgex-go/export/mongo"
	consulclient "github.com/edgexfoundry/edgex-go/support/consul-client"

	"go.uber.org/zap"
	"gopkg.in/mgo.v2"
)

const (
	port                   int    = 48071
	defHostname            string = "127.0.0.1"
	defMongoURL            string = "0.0.0.0"
	defMongoUsername       string = ""
	defMongoPassword       string = ""
	defMongoDatabase       string = "coredata"
	defMongoPort           int    = 27017
	defMongoConnectTimeout int    = 5000
	defMongoSocketTimeout  int    = 5000
	defConsulHost          string = "127.0.0.1"
	defConsulPort          int    = 8500

	envClientHost string = "EXPORT_CLIENT_HOST"
	envMongoURL   string = "EXPORT_CLIENT_MONGO_URL"
	envDistroHost string = "EXPORT_CLIENT_DISTRO_HOST"
	envConsulHost string = "EXPORT_CLIENT_CONSUL_HOST"
	envConsulPort string = "EXPORT_CLIENT_CONSUL_PORT"

	applicationName string = "export-client"
	consulProfile   string = "go"
)

type config struct {
	Port                int
	MongoURL            string
	MongoUser           string
	MongoPass           string
	MongoDatabase       string
	MongoPort           int
	MongoConnectTimeout int
	MongoSocketTimeout  int

	ConsulHost string
	ConsulPort int
	Hostname   string
}

var logger *zap.Logger

func main() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting edgex export client", zap.String("version", edgex.Version))

	client.InitLogger(logger)

	cfg, clientCfg := loadConfig()

	// Initialize service on Consul
	err := consulclient.ConsulInit(consulclient.ConsulConfig{
		ServiceName:    applicationName,
		ServicePort:    cfg.Port,
		ServiceAddress: cfg.Hostname,
		CheckAddress:   "http://" + cfg.Hostname + ":" + strconv.Itoa(cfg.Port) + "/api/v1/ping",
		CheckInterval:  "10s",
		ConsulAddress:  cfg.ConsulHost,
		ConsulPort:     cfg.ConsulPort,
	})

	if err == nil {
		logger.Info("Registered microservice in consul",
			zap.String("consulHost", cfg.ConsulHost),
			zap.Int("consulPort", cfg.ConsulPort))

		consulProfiles := []string{consulProfile}
		if err := consulclient.CheckKeyValuePairs(clientCfg, applicationName, consulProfiles); err != nil {
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

	ms, err := connectToMongo(cfg)
	if err != nil {
		logger.Error("Failed to connect to Mongo.", zap.Error(err))
		return
	}
	defer ms.Close()

	repo := mongo.NewRepository(ms)
	client.InitMongoRepository(repo)

	errs := make(chan error, 2)

	client.StartHTTPServer(*clientCfg, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	c := <-errs
	logger.Info("terminated", zap.String("error", c.Error()))
}

func loadConfig() (*config, *client.Config) {
	clientCfg := client.GetDefaultConfig()
	clientCfg.DistroHost = env(envDistroHost, clientCfg.DistroHost)

	cfg := config{
		MongoURL:            env(envMongoURL, defMongoURL),
		MongoUser:           defMongoUsername,
		MongoPass:           defMongoPassword,
		MongoDatabase:       defMongoDatabase,
		MongoPort:           defMongoPort,
		MongoConnectTimeout: defMongoConnectTimeout,
		MongoSocketTimeout:  defMongoSocketTimeout,
		ConsulHost:          env(envConsulHost, defConsulHost),
		ConsulPort:          defConsulPort,
	}

	hostname, err := os.Hostname()
	if err == nil {
		cfg.Hostname = hostname
	}
	cfg.Hostname = env(envClientHost, cfg.Hostname)

	portStr := env(envConsulPort, strconv.Itoa(cfg.ConsulPort))

	port, err := strconv.Atoi(portStr)
	if err == nil {
		cfg.ConsulPort = port
	} else {
		logger.Warn("Could not parse port", zap.String("port", portStr), zap.Error(err))
	}
	return &cfg, &clientCfg
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func connectToMongo(cfg *config) (*mgo.Session, error) {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{cfg.MongoURL + ":" + strconv.Itoa(cfg.MongoPort)},
		Timeout:  time.Duration(cfg.MongoConnectTimeout) * time.Millisecond,
		Database: cfg.MongoDatabase,
		Username: cfg.MongoUser,
		Password: cfg.MongoPass,
	}

	ms, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return nil, err
	}

	ms.SetSocketTimeout(time.Duration(cfg.MongoSocketTimeout) * time.Millisecond)
	ms.SetMode(mgo.Monotonic, true)

	return ms, nil
}
