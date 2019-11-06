/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package metadata

import (
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/gorilla/mux"
)

func LoadRestRoutes(dic *di.Container) *mux.Router {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, func(w http.ResponseWriter, r *http.Request) {
		configHandler(w, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	b := r.PathPrefix(clients.ApiBase).Subrouter()

	loadDeviceRoutes(b, dic)
	loadDeviceProfileRoutes(b, dic)
	loadDeviceReportRoutes(b, dic)
	loadDeviceServiceRoutes(b, dic)
	loadProvisionWatcherRoutes(b, dic)
	loadAddressableRoutes(b, dic)
	loadCommandRoutes(b, dic)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}
func loadDeviceRoutes(b *mux.Router, dic *di.Container) {
	// /api/v1/" + DEVICE
	b.HandleFunc("/"+DEVICE, func(w http.ResponseWriter, r *http.Request) {
		restAddNewDevice(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPost)
	b.HandleFunc("/"+DEVICE, func(w http.ResponseWriter, r *http.Request) {
		restUpdateDevice(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	b.HandleFunc("/"+DEVICE, func(w http.ResponseWriter, r *http.Request) {
		restGetAllDevices(w, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	d := b.PathPrefix("/" + DEVICE).Subrouter()

	d.HandleFunc("/"+LABEL+"/{"+LABEL+"}", restGetDevicesWithLabel).Methods(http.MethodGet)
	d.HandleFunc("/"+PROFILE+"/{"+PROFILEID+"}", restGetDeviceByProfileId).Methods(http.MethodGet)
	d.HandleFunc("/"+SERVICE+"/{"+SERVICEID+"}", restGetDeviceByServiceId).Methods(http.MethodGet)
	d.HandleFunc("/"+SERVICENAME+"/{"+SERVICENAME+"}", restGetDeviceByServiceName).Methods(http.MethodGet)
	d.HandleFunc("/"+PROFILENAME+"/{"+PROFILENAME+"}", restGetDeviceByProfileName).Methods(http.MethodGet)

	// /api/v1/" + DEVICE" + ID + "
	d.HandleFunc("/{"+ID+"}", restGetDeviceById).Methods(http.MethodGet)
	d.HandleFunc("/{"+ID+"}", func(w http.ResponseWriter, r *http.Request) {
		restSetDeviceStateById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	d.HandleFunc("/"+ID+"/{"+ID+"}", func(w http.ResponseWriter, r *http.Request) {
		restDeleteDeviceById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodDelete)
	d.HandleFunc("/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", func(w http.ResponseWriter, r *http.Request) {
		restSetDeviceLastReportedById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}/{"+LASTREPORTEDNOTIFY+"}", func(w http.ResponseWriter, r *http.Request) {
		restSetDeviceLastReportedByIdNotify(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", func(w http.ResponseWriter, r *http.Request) {
		restSetDeviceLastConnectedById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}/{"+LASTCONNECTEDNOTIFY+"}", func(w http.ResponseWriter, r *http.Request) {
		restSetLastConnectedByIdNotify(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	d.HandleFunc("/"+CHECK+"/{"+ID+"}", func(w http.ResponseWriter, r *http.Request) {
		restCheckForDevice(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	// /api/v1/" + DEVICE/" + NAME + "
	n := d.PathPrefix("/" + NAME).Subrouter()

	n.HandleFunc("/{"+NAME+"}", restGetDeviceByName).Methods(http.MethodGet)
	n.HandleFunc("/{"+NAME+"}", func(w http.ResponseWriter, r *http.Request) {
		restDeleteDeviceByName(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodDelete)
	n.HandleFunc("/{"+NAME+"}", func(w http.ResponseWriter, r *http.Request) {
		restSetDeviceStateByDeviceName(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", func(w http.ResponseWriter, r *http.Request) {
		restSetDeviceLastReportedByName(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}/{"+LASTREPORTEDNOTIFY+"}", func(w http.ResponseWriter, r *http.Request) {
		restSetDeviceLastReportedByNameNotify(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", func(w http.ResponseWriter, r *http.Request) {
		restSetDeviceLastConnectedByName(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}/{"+LASTCONNECTEDNOTIFY+"}", func(w http.ResponseWriter, r *http.Request) {
		restSetDeviceLastConnectedByNameNotify(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)

}

func loadDeviceProfileRoutes(b *mux.Router, dic *di.Container) {
	///api/v1/" + DEVICEPROFILE + "
	b.HandleFunc("/"+DEVICEPROFILE+"", func(w http.ResponseWriter, r *http.Request) {
		restGetAllDeviceProfiles(w, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)
	b.HandleFunc("/"+DEVICEPROFILE+"", func(w http.ResponseWriter, r *http.Request) {
		restAddDeviceProfile(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPost)
	b.HandleFunc("/"+DEVICEPROFILE+"", func(w http.ResponseWriter, r *http.Request) {
		restUpdateDeviceProfile(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)

	dp := b.PathPrefix("/" + DEVICEPROFILE).Subrouter()
	dp.HandleFunc("/{"+ID+"}", restGetProfileByProfileId).Methods(http.MethodGet)
	dp.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteProfileByProfileId).Methods(http.MethodDelete)
	dp.HandleFunc("/"+UPLOADFILE, restAddProfileByYaml).Methods(http.MethodPost)
	dp.HandleFunc("/"+UPLOAD, restAddProfileByYamlRaw).Methods(http.MethodPost)
	dp.HandleFunc("/"+MODEL+"/{"+MODEL+"}", restGetProfileByModel).Methods(http.MethodGet)
	dp.HandleFunc("/"+LABEL+"/{"+LABEL+"}", restGetProfileWithLabel).Methods(http.MethodGet)

	// /api/v1/" + DEVICEPROFILE + "/"  + MANUFACTURER + "
	dpm := dp.PathPrefix("/" + MANUFACTURER).Subrouter()
	dpm.HandleFunc("/{"+MANUFACTURER+"}/"+MODEL+"/{"+MODEL+"}", restGetProfileByManufacturerModel).Methods(http.MethodGet)
	dpm.HandleFunc("/{"+MANUFACTURER+"}", restGetProfileByManufacturer).Methods(http.MethodGet)

	// /api/v1/" + DEVICEPROFILE + "/" + NAME + "
	dpn := dp.PathPrefix("/" + NAME).Subrouter()
	dpn.HandleFunc("/{"+NAME+"}", restGetProfileByName).Methods(http.MethodGet)
	dpn.HandleFunc("/{"+NAME+"}", restDeleteProfileByName).Methods(http.MethodDelete)

	// /api/v1/" + DEVICEPROFILE + "/"  + YAML
	dpy := dp.PathPrefix("/" + YAML).Subrouter()
	// TODO add functionality
	dpy.HandleFunc("/"+NAME+"/{"+NAME+"}", restGetYamlProfileByName).Methods(http.MethodGet)
	dpy.HandleFunc("/{"+ID+"}", func(w http.ResponseWriter, r *http.Request) {
		restGetYamlProfileById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)
}
func loadDeviceReportRoutes(b *mux.Router, dic *di.Container) {
	// /api/v1/devicereport
	b.HandleFunc("/"+DEVICEREPORT, func(w http.ResponseWriter, r *http.Request) {
		restGetAllDeviceReports(w, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	b.HandleFunc("/"+DEVICEREPORT, func(w http.ResponseWriter, r *http.Request) {
		restAddDeviceReport(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPost)
	b.HandleFunc("/"+DEVICEREPORT, func(w http.ResponseWriter, r *http.Request) {
		restUpdateDeviceReport(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)

	dr := b.PathPrefix("/" + DEVICEREPORT).Subrouter()
	dr.HandleFunc("/{"+ID+"}", restGetReportById).Methods(http.MethodGet)
	dr.HandleFunc("/"+ID+"/{"+ID+"}", func(w http.ResponseWriter, r *http.Request) {
		restDeleteReportById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodDelete)

	dr.HandleFunc("/"+DEVICENAME+"/{"+DEVICENAME+"}", restGetDeviceReportByDeviceName).Methods(http.MethodGet)

	// /api/v1/devicereport/" + NAME + "
	drn := dr.PathPrefix("/" + NAME).Subrouter()
	drn.HandleFunc("/{"+NAME+"}", restGetReportByName).Methods(http.MethodGet)
	drn.HandleFunc("/{"+NAME+"}", func(w http.ResponseWriter, r *http.Request) {
		restDeleteReportByName(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodDelete)

	// /api/v1/devicereport/valueDescriptorsFor/devicename
	drvd := dr.PathPrefix("/" + VALUEDESCRIPTORSFOR).Subrouter()
	drvd.HandleFunc("/{"+DEVICENAME+"}", restGetValueDescriptorsForDeviceName).Methods(http.MethodGet)
}
func loadDeviceServiceRoutes(b *mux.Router, dic *di.Container) {
	// /api/v1/deviceservice
	b.HandleFunc("/"+DEVICESERVICE, func(w http.ResponseWriter, r *http.Request) {
		restGetAllDeviceServices(w, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	b.HandleFunc("/"+DEVICESERVICE, restAddDeviceService).Methods(http.MethodPost)
	b.HandleFunc("/"+DEVICESERVICE, func(w http.ResponseWriter, r *http.Request) {
		restUpdateDeviceService(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)

	ds := b.PathPrefix("/" + DEVICESERVICE).Subrouter()
	ds.HandleFunc("/"+ADDRESSABLENAME+"/{"+ADDRESSABLENAME+"}", restGetServiceByAddressableName).Methods(http.MethodGet)
	ds.HandleFunc("/"+ADDRESSABLE+"/{"+ADDRESSABLEID+"}", restGetServiceByAddressableId).Methods(http.MethodGet)
	ds.HandleFunc("/"+LABEL+"/{"+LABEL+"}", restGetServiceWithLabel).Methods(http.MethodGet)

	// /api/v1/deviceservice/" + NAME + "
	dsn := ds.PathPrefix("/" + NAME).Subrouter()
	dsn.HandleFunc("/{"+NAME+"}", restGetServiceByName).Methods(http.MethodGet)
	dsn.HandleFunc("/{"+NAME+"}", func(w http.ResponseWriter, r *http.Request) {
		restDeleteServiceByName(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodDelete)

	dsn.HandleFunc("/{"+NAME+"}/"+OPSTATE+"/{"+OPSTATE+"}", restUpdateServiceOpStateByName).Methods(http.MethodPut)
	dsn.HandleFunc("/{"+NAME+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restUpdateServiceAdminStateByName).Methods(http.MethodPut)
	dsn.HandleFunc("/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", func(w http.ResponseWriter, r *http.Request) {
		restUpdateServiceLastReportedByName(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	dsn.HandleFunc("/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", func(w http.ResponseWriter, r *http.Request) {
		restUpdateServiceLastConnectedByName(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)

	// /api/v1/"  + DEVICESERVICE + ID + "
	ds.HandleFunc("/{"+ID+"}", restGetServiceById).Methods(http.MethodGet)
	ds.HandleFunc("/"+ID+"/{"+ID+"}", func(w http.ResponseWriter, r *http.Request) {
		restDeleteServiceById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodDelete)

	ds.HandleFunc("/"+ID+"/{"+ID+"}", func(w http.ResponseWriter, r *http.Request) {
		restDeleteServiceById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodDelete)

	ds.HandleFunc("/{"+ID+"}/"+OPSTATE+"/{"+OPSTATE+"}", restUpdateServiceOpStateById).Methods(http.MethodPut)
	ds.HandleFunc("/{"+ID+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restUpdateServiceAdminStateById).Methods(http.MethodPut)
	ds.HandleFunc("/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", func(w http.ResponseWriter, r *http.Request) {
		restUpdateServiceLastReportedById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	ds.HandleFunc("/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", func(w http.ResponseWriter, r *http.Request) {
		restUpdateServiceLastConnectedById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)

}

func loadProvisionWatcherRoutes(b *mux.Router, dic *di.Container) {
	b.HandleFunc("/"+PROVISIONWATCHER, func(w http.ResponseWriter, r *http.Request) {
		restAddProvisionWatcher(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPost)
	b.HandleFunc("/"+PROVISIONWATCHER, func(w http.ResponseWriter, r *http.Request) {
		restUpdateProvisionWatcher(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPut)
	b.HandleFunc("/"+PROVISIONWATCHER, func(w http.ResponseWriter, r *http.Request) {
		restGetProvisionWatchers(w)
	}).Methods(http.MethodGet)

	pw := b.PathPrefix("/" + PROVISIONWATCHER).Subrouter()
	// /api/v1/provisionwatcher
	pw.HandleFunc("/"+ID+"/{"+ID+"}", func(w http.ResponseWriter, r *http.Request) {
		restDeleteProvisionWatcherById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodDelete)

	pw.HandleFunc("/{"+ID+"}", restGetProvisionWatcherById).Methods(http.MethodGet)
	pw.HandleFunc("/"+NAME+"/{"+NAME+"}", func(w http.ResponseWriter, r *http.Request) {
		restDeleteProvisionWatcherByName(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodDelete)

	pw.HandleFunc("/"+NAME+"/{"+NAME+"}", restGetProvisionWatcherByName).Methods(http.MethodGet)
	pw.HandleFunc("/"+PROFILENAME+"/{"+NAME+"}", restGetProvisionWatchersByProfileName).Methods(http.MethodGet)
	pw.HandleFunc("/"+PROFILE+"/{"+ID+"}", restGetProvisionWatchersByProfileId).Methods(http.MethodGet)
	pw.HandleFunc("/"+SERVICE+"/{"+ID+"}", restGetProvisionWatchersByServiceId).Methods(http.MethodGet)
	pw.HandleFunc("/"+SERVICENAME+"/{"+NAME+"}", restGetProvisionWatchersByServiceName).Methods(http.MethodGet)
	pw.HandleFunc("/"+IDENTIFIER+"/{"+KEY+"}/{"+VALUE+"}", restGetProvisionWatchersByIdentifier).Methods(http.MethodGet)

}
func loadAddressableRoutes(b *mux.Router, dic *di.Container) {
	// /api/v1/" + ADDRESSABLE + "
	b.HandleFunc("/"+ADDRESSABLE, func(w http.ResponseWriter, r *http.Request) {
		restGetAllAddressables(w, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	b.HandleFunc("/"+ADDRESSABLE, func(w http.ResponseWriter, r *http.Request) {
		restAddAddressable(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPost)

	b.HandleFunc("/"+ADDRESSABLE, func(w http.ResponseWriter, r *http.Request) {
		restUpdateAddressable(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodPost)

	a := b.PathPrefix("/" + ADDRESSABLE).Subrouter()
	a.HandleFunc("/{"+ID+"}", restGetAddressableById).Methods(http.MethodGet)
	a.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteAddressableById).Methods(http.MethodDelete)
	a.HandleFunc("/"+NAME+"/{"+NAME+"}", restGetAddressableByName).Methods(http.MethodGet)
	a.HandleFunc("/"+NAME+"/{"+NAME+"}", restDeleteAddressableByName).Methods(http.MethodDelete)
	a.HandleFunc("/"+TOPIC+"/{"+TOPIC+"}", restGetAddressableByTopic).Methods(http.MethodGet)
	a.HandleFunc("/"+PORT+"/{"+PORT+"}", restGetAddressableByPort).Methods(http.MethodGet)
	a.HandleFunc("/"+PUBLISHER+"/{"+PUBLISHER+"}", restGetAddressableByPublisher).Methods(http.MethodGet)
	a.HandleFunc("/"+ADDRESS+"/{"+ADDRESS+"}", restGetAddressableByAddress).Methods(http.MethodGet)
}
func loadCommandRoutes(b *mux.Router, dic *di.Container) {

	// /api/v1/command
	b.HandleFunc("/"+COMMAND, func(w http.ResponseWriter, r *http.Request) {
		restGetAllCommands(w, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	b.HandleFunc("/"+COMMAND, func(w http.ResponseWriter, r *http.Request) {
		restGetAllCommands(w, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	c := b.PathPrefix("/" + COMMAND).Subrouter()
	c.HandleFunc("/{"+ID+"}", func(w http.ResponseWriter, r *http.Request) {
		restGetCommandById(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)
	c.HandleFunc("/"+NAME+"/{"+NAME+"}", func(w http.ResponseWriter, r *http.Request) {
		restGetCommandsByName(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

	d := c.PathPrefix("/" + DEVICE).Subrouter()
	d.HandleFunc("/{"+ID+"}", func(w http.ResponseWriter, r *http.Request) {
		restGetCommandsByDeviceId(w, r, container.LoggingClientFrom(dic.Get))
	}).Methods(http.MethodGet)

}
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(clients.ContentType, clients.ContentTypeText)
	w.Write([]byte("pong"))
}

func configHandler(
	w http.ResponseWriter,
	loggingClient logger.LoggingClient) {

	pkg.Encode(Configuration, w, loggingClient)
}

func metricsHandler(
	w http.ResponseWriter,
	loggingClient logger.LoggingClient) {

	s := telemetry.NewSystemUsage()

	pkg.Encode(s, w, loggingClient)

	return
}
