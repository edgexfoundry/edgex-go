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

	metadataContainer "github.com/edgexfoundry/edgex-go/internal/core/metadata/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	errorContainer "github.com/edgexfoundry/edgex-go/internal/pkg/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"

	"github.com/gorilla/mux"
)

func loadRestRoutes(r *mux.Router, dic *di.Container) {
	// Ping Resource
	r.HandleFunc(
		clients.ApiPingRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(clients.ContentType, clients.ContentTypeText)
			_, _ = w.Write([]byte("pong"))
		}).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(
		clients.ApiConfigRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			pkg.Encode(metadataContainer.ConfigurationFrom(dic.Get), w, bootstrapContainer.LoggingClientFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(
		clients.ApiMetricsRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			pkg.Encode(telemetry.NewSystemUsage(), w, bootstrapContainer.LoggingClientFrom(dic.Get))
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
}

func loadDeviceRoutes(b *mux.Router, dic *di.Container) {
	// /api/v1/" + DEVICE
	b.HandleFunc(
		"/"+DEVICE,
		func(w http.ResponseWriter, r *http.Request) {
			restAddNewDevice(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPost)
	b.HandleFunc(
		"/"+DEVICE,
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateDevice(w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)
	b.HandleFunc(
		"/"+DEVICE,
		func(w http.ResponseWriter, r *http.Request) {
			restGetAllDevices(
				w,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)

	d := b.PathPrefix("/" + DEVICE).Subrouter()

	d.HandleFunc(
		"/"+LABEL+"/{"+LABEL+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetDevicesWithLabel(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	d.HandleFunc(
		"/"+PROFILE+"/{"+PROFILEID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetDeviceByProfileId(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	d.HandleFunc(
		"/"+SERVICE+"/{"+SERVICEID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetDeviceByServiceId(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	d.HandleFunc(
		"/"+SERVICENAME+"/{"+SERVICENAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetDeviceByServiceName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	d.HandleFunc(
		"/"+PROFILENAME+"/{"+PROFILENAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetDeviceByProfileName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	// /api/v1/" + DEVICE" + ID + "
	d.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetDeviceById(w, r, container.DBClientFrom(dic.Get), errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	d.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restSetDeviceStateById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)
	d.HandleFunc(
		"/"+ID+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteDeviceById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodDelete)
	d.HandleFunc(
		"/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restSetDeviceLastReportedById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)
	d.HandleFunc(
		"/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}/{"+LASTREPORTEDNOTIFY+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restSetDeviceLastReportedByIdNotify(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)
	d.HandleFunc(
		"/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restSetDeviceLastConnectedById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get), errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)
	d.HandleFunc(
		"/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}/{"+LASTCONNECTEDNOTIFY+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restSetLastConnectedByIdNotify(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)
	d.HandleFunc(
		"/"+CHECK+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restCheckForDevice(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	// /api/v1/" + DEVICE/" + NAME + "
	n := d.PathPrefix("/" + NAME).Subrouter()
	n.HandleFunc(
		"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetDeviceByName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	n.HandleFunc(
		"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteDeviceByName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodDelete)
	n.HandleFunc(
		"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restSetDeviceStateByDeviceName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)
	n.HandleFunc(
		"/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}",
		func(w http.ResponseWriter,
			r *http.Request) {
			restSetDeviceLastReportedByName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)
	n.HandleFunc(
		"/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}/{"+LASTREPORTEDNOTIFY+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restSetDeviceLastReportedByNameNotify(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)
	n.HandleFunc(
		"/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}",
		func(w http.ResponseWriter,
			r *http.Request) {
			restSetDeviceLastConnectedByName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)
	n.HandleFunc(
		"/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}/{"+LASTCONNECTEDNOTIFY+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restSetDeviceLastConnectedByNameNotify(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get), errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)

}

func loadDeviceProfileRoutes(b *mux.Router, dic *di.Container) {
	///api/v1/" + DEVICEPROFILE + "
	b.HandleFunc(
		"/"+DEVICEPROFILE+"",
		func(w http.ResponseWriter, r *http.Request) {
			restGetAllDeviceProfiles(
				w,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+DEVICEPROFILE+"",
		func(w http.ResponseWriter, r *http.Request) {
			restAddDeviceProfile(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.CoreDataValueDescriptorClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPost)
	b.HandleFunc(
		"/"+DEVICEPROFILE+"",
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateDeviceProfile(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.CoreDataValueDescriptorClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPut)

	dp := b.PathPrefix("/" + DEVICEPROFILE).Subrouter()
	dp.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProfileByProfileId(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	dp.HandleFunc(
		"/"+ID+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteProfileByProfileId(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodDelete)
	dp.HandleFunc(
		"/"+UPLOADFILE,
		func(w http.ResponseWriter, r *http.Request) {
			restAddProfileByYaml(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.CoreDataValueDescriptorClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPost)
	dp.HandleFunc(
		"/"+UPLOAD,
		func(w http.ResponseWriter, r *http.Request) {
			restAddProfileByYamlRaw(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.CoreDataValueDescriptorClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPost)
	dp.HandleFunc(
		"/"+MODEL+"/{"+MODEL+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProfileByModel(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	dp.HandleFunc(
		"/"+LABEL+"/{"+LABEL+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProfileWithLabel(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	// /api/v1/" + DEVICEPROFILE + "/"  + MANUFACTURER + "
	dpm := dp.PathPrefix("/" + MANUFACTURER).Subrouter()
	dpm.HandleFunc(
		"/{"+MANUFACTURER+"}/"+MODEL+"/{"+MODEL+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProfileByManufacturerModel(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	dpm.HandleFunc(
		"/{"+MANUFACTURER+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProfileByManufacturer(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	// /api/v1/" + DEVICEPROFILE + "/" + NAME + "
	dpn := dp.PathPrefix("/" + NAME).Subrouter()
	dpn.HandleFunc(
		"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProfileByName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	dpn.HandleFunc(
		"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteProfileByName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodDelete)

	// /api/v1/" + DEVICEPROFILE + "/"  + YAML
	dpy := dp.PathPrefix("/" + YAML).Subrouter()
	// TODO add functionality
	dpy.HandleFunc(
		"/"+NAME+"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetYamlProfileByName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	dpy.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetYamlProfileById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

}
func loadDeviceReportRoutes(b *mux.Router, dic *di.Container) {
	// /api/v1/devicereport
	b.HandleFunc(
		"/"+DEVICEREPORT,
		func(w http.ResponseWriter, r *http.Request) {
			restGetAllDeviceReports(
				w,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)

	b.HandleFunc(
		"/"+DEVICEREPORT,
		func(w http.ResponseWriter, r *http.Request) {
			restAddDeviceReport(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPost)
	b.HandleFunc(
		"/"+DEVICEREPORT,
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateDeviceReport(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)

	dr := b.PathPrefix("/" + DEVICEREPORT).Subrouter()
	dr.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetReportById(w, r, container.DBClientFrom(dic.Get), errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	dr.HandleFunc(
		"/"+ID+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteReportById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodDelete)
	dr.HandleFunc(
		"/"+DEVICENAME+"/{"+DEVICENAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetDeviceReportByDeviceName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	// /api/v1/devicereport/" + NAME + "
	drn := dr.PathPrefix("/" + NAME).Subrouter()
	drn.HandleFunc(
		"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetReportByName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	drn.HandleFunc(
		"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteReportByName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodDelete)

	// /api/v1/devicereport/valueDescriptorsFor/devicename
	drvd := dr.PathPrefix("/" + VALUEDESCRIPTORSFOR).Subrouter()
	drvd.HandleFunc(
		"/{"+DEVICENAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetValueDescriptorsForDeviceName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
}
func loadDeviceServiceRoutes(b *mux.Router, dic *di.Container) {
	// /api/v1/deviceservice
	b.HandleFunc(
		"/"+DEVICESERVICE,
		func(w http.ResponseWriter, r *http.Request) {
			restGetAllDeviceServices(
				w,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+DEVICESERVICE,
		func(w http.ResponseWriter, r *http.Request) {
			restAddDeviceService(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPost)

	b.HandleFunc(
		"/"+DEVICESERVICE,
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateDeviceService(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)

	ds := b.PathPrefix("/" + DEVICESERVICE).Subrouter()
	ds.HandleFunc(
		"/"+ADDRESSABLENAME+"/{"+ADDRESSABLENAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetServiceByAddressableName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	ds.HandleFunc(
		"/"+ADDRESSABLE+"/{"+ADDRESSABLEID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetServiceByAddressableId(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	ds.HandleFunc(
		"/"+LABEL+"/{"+LABEL+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetServiceWithLabel(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	// /api/v1/deviceservice/" + NAME + "
	dsn := ds.PathPrefix("/" + NAME).Subrouter()
	dsn.HandleFunc(
		"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetServiceByName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	dsn.HandleFunc(
		"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteServiceByName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodDelete)
	dsn.HandleFunc(
		"/{"+NAME+"}/"+OPSTATE+"/{"+OPSTATE+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateServiceOpStateByName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)
	dsn.HandleFunc(
		"/{"+NAME+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}",
		func(w http.ResponseWriter,
			r *http.Request) {
			restUpdateServiceAdminStateByName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)

	dsn.HandleFunc(
		"/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}",
		func(w http.ResponseWriter,
			r *http.Request) {
			restUpdateServiceLastReportedByName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)
	dsn.HandleFunc(
		"/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateServiceLastConnectedByName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get), errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)

	// /api/v1/"  + DEVICESERVICE + ID + "
	ds.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetServiceById(w, r, container.DBClientFrom(dic.Get), errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	ds.HandleFunc(
		"/"+ID+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteServiceById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodDelete)
	ds.HandleFunc(
		"/"+ID+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteServiceById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.NotificationsClientFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodDelete)

	ds.HandleFunc(
		"/{"+ID+"}/"+OPSTATE+"/{"+OPSTATE+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateServiceOpStateById(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)

	ds.HandleFunc(
		"/{"+ID+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateServiceAdminStateById(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)

	ds.HandleFunc(
		"/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}",
		func(w http.ResponseWriter,
			r *http.Request) {
			restUpdateServiceLastReportedById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)
	ds.HandleFunc(
		"/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}",
		func(w http.ResponseWriter,
			r *http.Request) {
			restUpdateServiceLastConnectedById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)
}

func loadProvisionWatcherRoutes(b *mux.Router, dic *di.Container) {
	b.HandleFunc(
		"/"+PROVISIONWATCHER,
		func(w http.ResponseWriter, r *http.Request) {
			restAddProvisionWatcher(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPost)
	b.HandleFunc(
		"/"+PROVISIONWATCHER,
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateProvisionWatcher(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)
	b.HandleFunc(
		"/"+PROVISIONWATCHER,
		func(w http.ResponseWriter, r *http.Request) {
			restGetProvisionWatchers(
				w,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)

	pw := b.PathPrefix("/" + PROVISIONWATCHER).Subrouter()
	// /api/v1/provisionwatcher
	pw.HandleFunc(
		"/"+ID+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteProvisionWatcherById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodDelete)

	pw.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProvisionWatcherById(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	pw.HandleFunc(
		"/"+NAME+"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteProvisionWatcherByName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodDelete)

	pw.HandleFunc(
		"/"+NAME+"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProvisionWatcherByName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	pw.HandleFunc(
		"/"+PROFILENAME+"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProvisionWatchersByProfileName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	pw.HandleFunc(
		"/"+PROFILE+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProvisionWatchersByProfileId(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	pw.HandleFunc(
		"/"+SERVICE+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProvisionWatchersByServiceId(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	pw.HandleFunc(
		"/"+SERVICENAME+"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProvisionWatchersByServiceName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	pw.HandleFunc(
		"/"+IDENTIFIER+"/{"+KEY+"}/{"+VALUE+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetProvisionWatchersByIdentifier(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
}

func loadAddressableRoutes(b *mux.Router, dic *di.Container) {
	// /api/v1/" + ADDRESSABLE + "
	b.HandleFunc(
		"/"+ADDRESSABLE,
		func(w http.ResponseWriter, r *http.Request) {
			restGetAllAddressables(
				w,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)

	b.HandleFunc(
		"/"+ADDRESSABLE,
		func(w http.ResponseWriter, r *http.Request) {
			restAddAddressable(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPost)

	b.HandleFunc(
		"/"+ADDRESSABLE,
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateAddressable(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodPut)

	a := b.PathPrefix("/" + ADDRESSABLE).Subrouter()
	a.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetAddressableById(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	a.HandleFunc(
		"/"+ID+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteAddressableById(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodDelete)
	a.HandleFunc(
		"/"+NAME+"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetAddressableByName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	a.HandleFunc(
		"/"+NAME+"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteAddressableByName(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodDelete)
	a.HandleFunc(
		"/"+TOPIC+"/{"+TOPIC+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetAddressableByTopic(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	a.HandleFunc(
		"/"+PORT+"/{"+PORT+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetAddressableByPort(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	a.HandleFunc(
		"/"+PUBLISHER+"/{"+PUBLISHER+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetAddressableByPublisher(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	a.HandleFunc(
		"/"+ADDRESS+"/{"+ADDRESS+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetAddressableByAddress(
				w,
				r,
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
}

func loadCommandRoutes(b *mux.Router, dic *di.Container) {
	// /api/v1/command
	b.HandleFunc(
		"/"+COMMAND,
		func(w http.ResponseWriter, r *http.Request) {
			restGetAllCommands(
				w,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)

	b.HandleFunc(
		"/"+COMMAND,
		func(w http.ResponseWriter, r *http.Request) {
			restGetAllCommands(
				w,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get),
				metadataContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)

	c := b.PathPrefix("/" + COMMAND).Subrouter()
	c.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetCommandById(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
	c.HandleFunc(
		"/"+NAME+"/{"+NAME+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetCommandsByName(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)

	d := c.PathPrefix("/" + DEVICE).Subrouter()
	d.HandleFunc(
		"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetCommandsByDeviceId(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				errorContainer.ErrorHandlerFrom(dic.Get))
		}).Methods(http.MethodGet)
}
