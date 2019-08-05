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
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy"
	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

var lc = CreateLogging()

func CreateLogging() logger.LoggingClient {
	return logger.NewClient(proxy.SecurityService, false, fmt.Sprintf("%s-%s.log", proxy.SecurityService, time.Now().Format("2006-01-02")), models.InfoLog)
}

func main() {

	if len(os.Args) < 2 {
		proxy.HelpCallback()
	}
	var initNeeded bool
	var insecureSkipVerify bool
	var resetNeeded bool
	var useProfile string
	var userTobeCreated string
	var userofGroup string
	var userTobeDeleted string
	var useRegistry bool

	flag.BoolVar(&insecureSkipVerify,"insureskipverify", true, "skip server side SSL verification, mainly for self-signed cert")
	flag.BoolVar(&initNeeded,"init", false, "run init procedure for security service.")
	flag.BoolVar(&resetNeeded,"reset", false, "reset reverse proxy by removing all services/routes/consumers")
	flag.StringVar(&userTobeCreated,"useradd", "", "user that needs to be added to consume the edgex services")
	flag.StringVar(&userofGroup,"group", "user", "group that the user belongs to. By default it is in user group")
	flag.StringVar(&userTobeDeleted,"userdel", "", "user that needs to be deleted from the edgex services")
	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")

	flag.Usage = proxy.HelpCallback
	flag.Parse()

	params := startup.BootParams{UseRegistry: useRegistry, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, proxy.Retry, logBeforeInit)

	//er := proxy.EdgeXRequestor{ProxyBaseURL: config.GetProxyBaseURL(), SecretSvcBaseURL: config.GetSecretSvcBaseURL(), Client: client}
	//s := &proxy.Service{Connect: &er, CertCfg: config, ServiceCfg: config}
	req := proxy.NewRequestor(insecureSkipVerify)
	s := proxy.NewService(req)

	err := s.CheckProxyServiceStatus()
	if err != nil {
		lc.Error(err.Error())
		return
	}

	if initNeeded {
		if resetNeeded {
			lc.Error("can't run initialization and reset at the same time for security service")
			return
		}
		err = s.Init()
		if err != nil {
			lc.Error(err.Error())
		}
	}

	if resetNeeded {
		err = s.ResetProxy()
		if err != nil {
			lc.Error(err.Error())
		}
	}

	if userTobeCreated != "" && userofGroup != "" {
		c := proxy.NewConsumer(userTobeCreated, req)

		err := c.Create(proxy.EdgeXService)
		if err != nil {
			lc.Error(err.Error())
			return
		}

		err = c.AssociateWithGroup(userofGroup)
		if err != nil {
			lc.Error(err.Error())
			return
		}

		t, err := c.CreateToken()
		if err != nil {
			lc.Error(fmt.Sprintf("failed to create access token for edgex service due to error %s", err.Error()))
			return
		}

		fmt.Println(fmt.Sprintf("the access token for user %s is: %s. Please keep the token for accessing edgex services", userTobeCreated, t))

		utp := &proxy.UserTokenPair{User: userTobeCreated, Token: t}
		if err != nil {
			lc.Error(err.Error())
			return
		}
		err = utp.Save()
		if err != nil {
			lc.Error(err.Error())
		}
	}

	if userTobeDeleted != "" {
		t := proxy.NewConsumer(userTobeDeleted, req)
		t.Delete()
	}
}

func logBeforeInit(err error) {
	l := logger.NewClient("security-proxy-setup", false, "", models.InfoLog)
	l.Error(err.Error())
}