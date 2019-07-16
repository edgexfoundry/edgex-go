/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @author: Tingyu Zeng, Dell
 * @version: 1.1.0
 *******************************************************************************/
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	worker "github.com/edgexfoundry/edgex-go/internal/security/proxy"
	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	model "github.com/edgexfoundry/go-mod-core-contracts/models"
)

var lc = CreateLogging()

func CreateLogging() logger.LoggingClient {
	return logger.NewClient(worker.SecurityService, false, fmt.Sprintf("%s-%s.log", worker.SecurityService, time.Now().Format("2006-01-02")), model.InfoLog)
}

func main() {

	if len(os.Args) < 2 {
		worker.HelpCallback()
	}
	insecureSkipVerify := flag.Bool("insureskipverify", true, "skip server side SSL verification, mainly for self-signed cert")
	initNeeded := flag.Bool("init", false, "run init procedure for security service.")
	resetNeeded := flag.Bool("reset", false, "reset reverse proxy by removing all services/routes/consumers")
	userTobeCreated := flag.String("useradd", "", "user that needs to be added to consume the edgex services")
	userofGroup := flag.String("group", "user", "group that the user belongs to. By default it is in user group")
	userTobeDeleted := flag.String("userdel", "", "user that needs to be deleted from the edgex services")
	configFileLocation := flag.String("configfile", "res/configuration.toml", "configuration file")

	flag.Usage = worker.HelpCallback
	flag.Parse()

	config, err := worker.LoadTomlConfig(*configFileLocation)
	if err != nil {
		lc.Error("failed to retrieve config data from local file. Please make sure res/configuration.toml file exists with correct formats")
		return
	}

	client := getNewClient(*insecureSkipVerify)
	er := worker.EdgeXRequestor{ProxyBaseURL: config.GetProxyBaseURL(), SecretSvcBaseURL: config.GetSecretSvcBaseURL(), Client: client}
	s := &worker.Service{Connect: &er, CertCfg: config, ServiceCfg: config}

	err = s.CheckProxyServiceStatus()
	if err != nil {
		lc.Error(err.Error())
		return
	}

	if *initNeeded == true && *resetNeeded == true {
		lc.Error("can't run initialization and reset at the same time for security service")
		return
	}

	if *initNeeded == true {
		err = s.Init()
		if err != nil {
			lc.Error(err.Error())
		}
	}

	if *resetNeeded == true {
		err = s.ResetProxy()
		if err != nil {
			lc.Error(err.Error())
		}
	}

	if *userTobeCreated != "" && *userofGroup != "" {
		c := &worker.Consumer{Name: *userTobeCreated, Connect: &er, Cfg: config}

		err := c.Create(worker.EdgeXService)
		if err != nil {
			lc.Error(err.Error())
			return
		}

		err = c.AssociateWithGroup(*userofGroup)
		if err != nil {
			lc.Error(err.Error())
			return
		}

		t, err := c.CreateToken()
		if err != nil {
			lc.Error(fmt.Sprintf("failed to create access token for edgex service due to error %s", err.Error()))
			return
		}

		fmt.Println(fmt.Sprintf("the access token for user %s is: %s. Please keep the token for accessing edgex services", *userTobeCreated, t))

		utp := &worker.UserTokenPair{User: *userTobeCreated, Token: t}
		file, err := os.Create(config.GetProxyAuthOutputPath())
		if err != nil {
			lc.Error(err.Error())
			return
		}
		err = utp.Save(file)
		if err != nil {
			lc.Error(err.Error())
		}
	}

	if *userTobeDeleted != "" {
		t := &worker.Consumer{Name: *userTobeDeleted, Connect: &er, Cfg: config}
		t.Delete()
	}
}

func getNewClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	return &http.Client{Timeout: 10 * time.Second, Transport: tr}
}
