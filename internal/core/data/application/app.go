/*******************************************************************************
 * Copyright 2022 Intel Corp.
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
 *******************************************************************************/

package application

import (
	"context"
	"sync"

	gometrics "github.com/rcrowley/go-metrics"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

const (
	eventsPersistedMetricName   = "EventsPersisted"
	readingsPersistedMetricName = "ReadingsPersisted"
)

// CoreDataApp encapsulates the Core Data Application functionality
// TODO: Extend this App usage beyond Events.
type CoreDataApp struct {
	lc                       logger.LoggingClient
	eventsPersistedCounter   gometrics.Counter
	readingsPersistedCounter gometrics.Counter
}

// NewCoreDataApp create a new initialized Core Data application
func NewCoreDataApp(dic *di.Container) *CoreDataApp {
	app := &CoreDataApp{
		lc: bootstrapContainer.LoggingClientFrom(dic.Get),
	}

	app.eventsPersistedCounter = gometrics.NewCounter()
	app.readingsPersistedCounter = gometrics.NewCounter()
	metricsManager := bootstrapContainer.MetricsManagerFrom(dic.Get)
	if metricsManager == nil {
		app.lc.Error("Metric Manager not available. Events and Readings metrics will not be collected.")
		return app
	}

	if err := metricsManager.Register(eventsPersistedMetricName, app.eventsPersistedCounter, nil); err != nil {
		app.lc.Errorf("%s metrics will not be collected: %s", eventsPersistedMetricName, err.Error())
	}
	app.lc.Infof("Registered metrics counter %s", eventsPersistedMetricName)

	app.readingsPersistedCounter = gometrics.NewCounter()
	if err := metricsManager.Register(readingsPersistedMetricName, app.readingsPersistedCounter, nil); err != nil {
		app.lc.Errorf("%s metrics will not be collected: %s", readingsPersistedMetricName, err.Error())
	}
	app.lc.Infof("Registered metrics counter %s", readingsPersistedMetricName)

	return app
}

// CoreDataAppName contains the name of data's application.CoreDataApp{} instance in the DIC.
var CoreDataAppName = di.TypeInstanceToName(CoreDataApp{})

// CoreDataAppFrom helper function queries the DIC and returns the application.CoreDataApp instance.
func CoreDataAppFrom(get di.Get) *CoreDataApp {
	return get(CoreDataAppName).(*CoreDataApp)
}

// BootstrapHandler fulfills the BootstrapHandler contract and performs creation of the CoreDataApp.
func BootstrapHandler(_ context.Context, _ *sync.WaitGroup, _ startup.Timer, dic *di.Container) bool {
	app := NewCoreDataApp(dic)

	dic.Update(di.ServiceConstructorMap{
		CoreDataAppName: func(get di.Get) interface{} {
			return app
		},
	})

	return true
}
