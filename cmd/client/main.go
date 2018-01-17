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

	"github.com/edgexfoundry/export-go/client"
	"github.com/edgexfoundry/export-go/mongo"

	"go.uber.org/zap"
	"gopkg.in/mgo.v2"
)

const (
	port                   int    = 48071
	defMongoURL            string = "0.0.0.0"
	defMongoUsername       string = ""
	defMongoPassword       string = ""
	defMongoDatabase       string = "coredata"
	defMongoPort           int    = 27017
	defMongoConnectTimeout int    = 5000
	defMongoSocketTimeout  int    = 5000
	envMongoURL            string = "EXPORT_CLIENT_MONGO_URL"
	envDistroHost          string = "EXPORT_CLIENT_DISTRO_HOST"
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
}

func main() {
	cfg, clientCfg := loadConfig()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	client.InitLogger(logger)

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

	cfg := config{
		MongoURL:            env(envMongoURL, defMongoURL),
		MongoUser:           defMongoUsername,
		MongoPass:           defMongoPassword,
		MongoDatabase:       defMongoDatabase,
		MongoPort:           defMongoPort,
		MongoConnectTimeout: defMongoConnectTimeout,
		MongoSocketTimeout:  defMongoSocketTimeout,
	}

	clientCfg := client.GetDefaultConfig()
	clientCfg.DistroHost = env(envDistroHost, clientCfg.DistroHost)

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
