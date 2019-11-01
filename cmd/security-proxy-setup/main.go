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
 *******************************************************************************/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/internal/security/proxy"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func main() {

	if len(os.Args) < 2 {
		usage.HelpCallbackSecurityProxy()
	}
	var initNeeded bool
	var insecureSkipVerify bool
	var resetNeeded bool
	var configDir, profileDir string
	var userTobeCreated string
	var userOfGroup string
	var userToBeDeleted string
	var useRegistry bool

	flag.BoolVar(&insecureSkipVerify, "insecureSkipVerify", false, "skip server side SSL verification, mainly for self-signed cert")
	flag.BoolVar(&initNeeded, "init", false, "run init procedure for security service.")
	flag.BoolVar(&resetNeeded, "reset", false, "reset reverse proxy by removing all services/routes/consumers")
	flag.StringVar(&userTobeCreated, "useradd", "", "user that needs to be added to consume the edgex services")
	flag.StringVar(&userOfGroup, "group", "user", "group that the user belongs to. By default it is in user group")
	flag.StringVar(&userToBeDeleted, "userdel", "", "user that needs to be deleted from the edgex services")
	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry service.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry service.")
	flag.StringVar(&profileDir, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&profileDir, "p", "", "Specify a profile other than default.")
	flag.StringVar(&configDir, "confdir", "", "Specify local configuration directory")

	flag.Usage = usage.HelpCallbackSecurityProxy
	flag.Parse()

	params := startup.BootParams{
		UseRegistry: useRegistry,
		ConfigDir:   configDir,
		ProfileDir:  profileDir,
		BootTimeout: internal.BootTimeoutDefault,
	}
	startup.Bootstrap(params, proxy.Retry, logBeforeInit)
	if proxy.Configuration == nil {
		// proxy.LoggingClient wasn't initialized either
		os.Exit(1)
	}

	req := proxy.NewRequestor(insecureSkipVerify, proxy.Configuration.Writable.RequestTimeout)

	s := proxy.NewService(req)

	err := s.CheckProxyServiceStatus()
	if err != nil {
		proxy.LoggingClient.Error(err.Error())
		os.Exit(1)
	}

	if initNeeded {
		if resetNeeded {
			proxy.LoggingClient.Error("can't run initialization and reset at the same time for security service")
			os.Exit(1)
		}
		cert := proxy.NewCertificateLoader(req, proxy.Configuration.SecretService.CertPath, proxy.Configuration.SecretService.TokenPath)
		err = s.Init(cert) // Where the Service init is called
		if err != nil {
			proxy.LoggingClient.Error(err.Error())
			os.Exit(1)
		}
	} else if resetNeeded {
		err = s.ResetProxy()
		if err != nil {
			proxy.LoggingClient.Error(err.Error())
			os.Exit(1)
		}
	}

	if userTobeCreated != "" && userOfGroup != "" {
		c := proxy.NewConsumer(userTobeCreated, req)

		err := c.Create(proxy.EdgeXKong)
		if err != nil {
			proxy.LoggingClient.Error(err.Error())
			os.Exit(1)
		}

		err = c.AssociateWithGroup(userOfGroup)
		if err != nil {
			proxy.LoggingClient.Error(err.Error())
			os.Exit(1)
		}

		t, err := c.CreateToken()
		if err != nil {
			proxy.LoggingClient.Error(fmt.Sprintf("failed to create access token for edgex service due to error %s", err.Error()))
			os.Exit(1)
		}

		fmt.Println(fmt.Sprintf("the access token for user %s is: %s. Please keep the token for accessing edgex services", userTobeCreated, t))

		utp := &proxy.UserTokenPair{User: userTobeCreated, Token: t}
		if err != nil {
			proxy.LoggingClient.Error(err.Error())
			os.Exit(1)
		}
		file, err := os.Create(proxy.Configuration.KongAuth.OutputPath)
		if err != nil {
			proxy.LoggingClient.Error(err.Error())
			os.Exit(1)
		}

		err = utp.Save(file)
		if err != nil {
			proxy.LoggingClient.Error(err.Error())
			os.Exit(1)
		}
	}

	if userToBeDeleted != "" {
		t := proxy.NewConsumer(userToBeDeleted, req)
		t.Delete()
		if err != nil {
			proxy.LoggingClient.Error(err.Error())
			os.Exit(1)
		}
	}
}

func logBeforeInit(err error) {
	l := logger.NewClient("security-proxy-setup", false, "", models.InfoLog)
	l.Error(err.Error())
}
