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
 *
 * @microservice: core-metadata-go service
 * @author: Spencer Bull & Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package metadata

import (
	"net/http"

	"github.com/gorilla/mux"
)

func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()
	b.HandleFunc("/ping", ping)

	loadDeviceRoutes(b)
	loadDeviceProfileRoutes(b)
	loadDeviceReportRoutes(b)
	loadDeviceServiceRoutes(b)
	loadScheduleEventRoutes(b)
	loadScheduleRoutes(b)
	loadProvisionWatcherRoutes(b)
	loadAddressableRoutes(b)
	loadCommandRoutes(b)
	return r
}
func loadDeviceRoutes(b *mux.Router) {
	// /api/v1/" + DEVICE
	b.HandleFunc("/"+DEVICE, restAddNewDevice).Methods(http.MethodPost)
	b.HandleFunc("/"+DEVICE, restUpdateDevice).Methods(http.MethodPut)
	b.HandleFunc("/"+DEVICE, restGetAllDevices).Methods(http.MethodGet)

	d := b.PathPrefix("/" + DEVICE).Subrouter()

	d.HandleFunc("/"+LABEL+"/{"+LABEL+"}", restGetDevicesWithLabel).Methods(http.MethodGet)
	d.HandleFunc("/"+PROFILE+"/{"+PROFILEID+"}", restGetDeviceByProfileId).Methods(http.MethodGet)
	d.HandleFunc("/"+SERVICE+"/{"+SERVICEID+"}", restGetDeviceByServiceId).Methods(http.MethodGet)
	d.HandleFunc("/"+SERVICENAME+"/{"+SERVICENAME+"}", restGetDeviceByServiceName).Methods(http.MethodGet)
	d.HandleFunc("/"+ADDRESSABLENAME+"/{"+ADDRESSABLENAME+"}", restGetDeviceByAddressableName).Methods(http.MethodGet)
	d.HandleFunc("/"+PROFILENAME+"/{"+PROFILENAME+"}", restGetDeviceByProfileName).Methods(http.MethodGet)
	d.HandleFunc("/"+ADDRESSABLE+"/{"+ADDRESSABLEID+"}", restGetDeviceByAddressableId).Methods(http.MethodGet)

	// /api/v1/" + DEVICE" + ID + "
	d.HandleFunc("/{"+ID+"}", restGetDeviceById).Methods(http.MethodGet)
	d.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteDeviceById).Methods(http.MethodDelete)
	d.HandleFunc("/{"+ID+"}/"+OPSTATE+"/{"+OPSTATE+"}", restSetDeviceOpStateById).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restSetDeviceAdminStateById).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", restSetDeviceLastReportedById).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}/{"+LASTREPORTEDNOTIFY+"}", restSetDeviceLastReportedByIdNotify).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", restSetDeviceLastConnectedById).Methods(http.MethodPut)
	d.HandleFunc("/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}/{"+LASTCONNECTEDNOTIFY+"}", restSetLastConnectedByIdNotify).Methods(http.MethodPut)

	// /api/v1/" + DEVICE/" + NAME + "
	n := d.PathPrefix("/" + NAME).Subrouter()
	n.HandleFunc("/{"+NAME+"}", restGetDeviceByName).Methods(http.MethodGet)
	n.HandleFunc("/{"+NAME+"}", restDeleteDeviceByName).Methods(http.MethodDelete)
	n.HandleFunc("/{"+NAME+"}/"+OPSTATE+"/{"+OPSTATE+"}", restSetDeviceOpStateByName).Methods(http.MethodPut)
	n.HandleFunc("/{"+NAME+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restSetDeviceAdminStateByName).Methods(http.MethodPut)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", restSetDeviceLastReportedByName).Methods(http.MethodPut)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}/{"+LASTREPORTEDNOTIFY+"}", restSetDeviceLastReportedByNameNotify).Methods(http.MethodPut)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", restSetDeviceLastConnectedByName).Methods(http.MethodPut)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}/{"+LASTCONNECTEDNOTIFY+"}", restSetDeviceLastConnectedByNameNotify).Methods(http.MethodPut)
}

func loadDeviceProfileRoutes(b *mux.Router) {
	///api/v1/" + DEVICEPROFILE + "
	b.HandleFunc("/"+DEVICEPROFILE+"", restGetAllDeviceProfiles).Methods(http.MethodGet)
	b.HandleFunc("/"+DEVICEPROFILE+"", restAddDeviceProfile).Methods(http.MethodPost)
	b.HandleFunc("/"+DEVICEPROFILE+"", restUpdateDeviceProfile).Methods(http.MethodPut)

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
	dpy.HandleFunc("/{"+ID+"}", restGetYamlProfileById).Methods(http.MethodGet)
}
func loadDeviceReportRoutes(b *mux.Router) {
	// /api/v1/devicereport
	b.HandleFunc("/"+DEVICEREPORT, restGetAllDeviceReports).Methods(http.MethodGet)
	b.HandleFunc("/"+DEVICEREPORT, restAddDeviceReport).Methods(http.MethodPost)
	b.HandleFunc("/"+DEVICEREPORT, restUpdateDeviceReport).Methods(http.MethodPut)

	dr := b.PathPrefix("/" + DEVICEREPORT).Subrouter()
	dr.HandleFunc("/{"+ID+"}", restGetReportById).Methods(http.MethodGet)
	dr.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteReportById).Methods(http.MethodDelete)
	dr.HandleFunc("/"+DEVICENAME+"/{"+DEVICENAME+"}", restGetDeviceReportByDeviceName).Methods(http.MethodGet)

	// /api/v1/devicereport/" + NAME + "
	drn := dr.PathPrefix("/" + NAME).Subrouter()
	drn.HandleFunc("/{"+NAME+"}", restGetReportByName).Methods(http.MethodGet)
	drn.HandleFunc("/{"+NAME+"}", restDeleteReportByName).Methods(http.MethodDelete)

	// /api/v1/devicereport/valueDescriptorsFor/devicename
	drvd := dr.PathPrefix("/" + VALUEDESCRIPTORSFOR).Subrouter()
	drvd.HandleFunc("/{"+DEVICENAME+"}", restGetValueDescriptorsForDeviceName).Methods(http.MethodGet)
}
func loadDeviceServiceRoutes(b *mux.Router) {
	// /api/v1/deviceservice
	b.HandleFunc("/"+DEVICESERVICE, restGetAllDeviceServices).Methods(http.MethodGet)
	b.HandleFunc("/"+DEVICESERVICE, restAddDeviceService).Methods(http.MethodPost)
	b.HandleFunc("/"+DEVICESERVICE, restUpdateDeviceService).Methods(http.MethodPut)

	ds := b.PathPrefix("/" + DEVICESERVICE).Subrouter()
	ds.HandleFunc("/"+ADDRESSABLENAME+"/{"+ADDRESSABLENAME+"}", restGetServiceByAddressableName).Methods(http.MethodGet)
	ds.HandleFunc("/"+ADDRESSABLE+"/{"+ADDRESSABLEID+"}", restGetServiceByAddressableId).Methods(http.MethodGet)
	ds.HandleFunc("/"+LABEL+"/{"+LABEL+"}", restGetServiceWithLabel).Methods(http.MethodGet)
	ds.HandleFunc("/"+DEVICEADDRESSABLES+"/{"+ID+"}", restGetAddressablesForAssociatedDevicesById).Methods(http.MethodGet)
	ds.HandleFunc("/"+DEVICEADDRESSABLESBYNAME+"/{"+NAME+"}", restGetAddressablesForAssociatedDevicesByName).Methods(http.MethodGet)

	// /api/v1/deviceservice/" + NAME + "
	dsn := ds.PathPrefix("/" + NAME).Subrouter()
	dsn.HandleFunc("/{"+NAME+"}", restGetServiceByName).Methods(http.MethodGet)
	dsn.HandleFunc("/{"+NAME+"}", restDeleteServiceByName).Methods(http.MethodDelete)
	dsn.HandleFunc("/{"+NAME+"}/"+OPSTATE+"/{"+OPSTATE+"}", restUpdateServiceOpStateByName).Methods(http.MethodPut)
	dsn.HandleFunc("/{"+NAME+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restUpdateServiceAdminStateByName).Methods(http.MethodPut)
	dsn.HandleFunc("/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", restUpdateServiceLastReportedByName).Methods(http.MethodPut)
	dsn.HandleFunc("/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", restUpdateServiceLastConnectedByName).Methods(http.MethodPut)

	// /api/v1/"  + DEVICESERVICE + ID + "
	ds.HandleFunc("/{"+ID+"}", restGetServiceById).Methods(http.MethodGet)
	ds.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteServiceById).Methods(http.MethodDelete)
	ds.HandleFunc("/{"+ID+"}/"+OPSTATE+"/{"+OPSTATE+"}", restUpdateServiceOpStateById).Methods(http.MethodPut)
	ds.HandleFunc("/{"+ID+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restUpdateServiceAdminStateById).Methods(http.MethodPut)
	ds.HandleFunc("/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", restUpdateServiceLastReportedById).Methods(http.MethodPut)
	ds.HandleFunc("/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", restUpdateServiceLastConnectedById).Methods(http.MethodPut)
}
func loadScheduleEventRoutes(b *mux.Router) {
	// /api/v1/scheduleevent
	b.HandleFunc("/"+SCHEDULEEVENT, restGetAllScheduleEvents).Methods(http.MethodGet)
	b.HandleFunc("/"+SCHEDULEEVENT, restAddScheduleEvent).Methods(http.MethodPost)
	b.HandleFunc("/"+SCHEDULEEVENT, restUpdateScheduleEvent).Methods(http.MethodPut)
	se := b.PathPrefix("/" + SCHEDULEEVENT).Subrouter()
	se.HandleFunc("/{"+ID+"}", restGetScheduleEventById).Methods(http.MethodGet)

	// /api/v1/scheduleevent/" + NAME + "
	sen := se.PathPrefix("/" + NAME + "").Subrouter()
	sen.HandleFunc("/{"+NAME+"}", restDeleteScheduleEventByName).Methods(http.MethodDelete)
	sen.HandleFunc("/{"+NAME+"}", restGetScheduleEventByName).Methods(http.MethodGet)

	// /api/v1/"  + SCHEDULEEVENT + ID + "
	seid := se.PathPrefix("/" + ID).Subrouter()
	seid.HandleFunc("/{"+ID+"}", restDeleteScheduleEventById).Methods(http.MethodDelete)

	// /api/v1/scheduleevent/addressable
	seaid := se.PathPrefix("/" + ADDRESSABLE).Subrouter()
	seaid.HandleFunc("/{"+ADDRESSABLEID+"}", restGetScheduleEventByAddressableId).Methods(http.MethodGet)

	sean := se.PathPrefix("/" + ADDRESSABLENAME).Subrouter()
	sean.HandleFunc("/{"+ADDRESSABLENAME+"}", restGetScheduleEventByAddressableName).Methods(http.MethodGet)

	// /api/v1/scheduleevent/servicename
	sesn := se.PathPrefix("/" + SERVICENAME).Subrouter()
	sesn.HandleFunc("/{"+SERVICENAME+"}", restGetScheduleEventsByServiceName).Methods(http.MethodGet)
}
func loadScheduleRoutes(b *mux.Router) {
	// /api/v1/schedule
	b.HandleFunc("/"+SCHEDULE, restGetAllSchedules).Methods(http.MethodGet)
	b.HandleFunc("/"+SCHEDULE, restAddSchedule).Methods(http.MethodPost)
	b.HandleFunc("/"+SCHEDULE, restUpdateSchedule).Methods(http.MethodPut)
	sch := b.PathPrefix("/" + SCHEDULE).Subrouter()
	sch.HandleFunc("/{"+ID+"}", restGetScheduleById).Methods(http.MethodGet)

	// /api/v1/schedule/" + NAME + "
	schn := sch.PathPrefix("/" + NAME + "").Subrouter()
	schn.HandleFunc("/{"+NAME+"}", restGetScheduleByName).Methods(http.MethodGet)
	schn.HandleFunc("/{"+NAME+"}", restDeleteScheduleByName).Methods(http.MethodDelete)

	// /api/v1/"  + SCHEDULE + ID + "
	schid := sch.PathPrefix("/" + ID).Subrouter()
	schid.HandleFunc("/{"+ID+"}", restDeleteScheduleById).Methods(http.MethodDelete)
}
func loadProvisionWatcherRoutes(b *mux.Router) {
	b.HandleFunc("/"+PROVISIONWATCHER, restAddProvisionWatcher).Methods(http.MethodPost)
	b.HandleFunc("/"+PROVISIONWATCHER, restUpdateProvisionWatcher).Methods(http.MethodPut)
	b.HandleFunc("/"+PROVISIONWATCHER, restGetProvisionWatchers).Methods(http.MethodGet)
	pw := b.PathPrefix("/" + PROVISIONWATCHER).Subrouter()
	// /api/v1/provisionwatcher
	pw.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteProvisionWatcherById).Methods(http.MethodDelete)
	pw.HandleFunc("/{"+ID+"}", restGetProvisionWatcherById).Methods(http.MethodGet)
	pw.HandleFunc("/"+NAME+"/{"+NAME+"}", restDeleteProvisionWatcherByName).Methods(http.MethodDelete)
	pw.HandleFunc("/"+NAME+"/{"+NAME+"}", restGetProvisionWatcherByName).Methods(http.MethodGet)
	pw.HandleFunc("/"+PROFILENAME+"/{"+NAME+"}", restGetProvisionWatchersByProfileName).Methods(http.MethodGet)
	pw.HandleFunc("/"+PROFILE+"/{"+ID+"}", restGetProvisionWatchersByProfileId).Methods(http.MethodGet)
	pw.HandleFunc("/"+SERVICE+"/{"+ID+"}", restGetProvisionWatchersByServiceId).Methods(http.MethodGet)
	pw.HandleFunc("/"+SERVICENAME+"/{"+NAME+"}", restGetProvisionWatchersByServiceName).Methods(http.MethodGet)
	pw.HandleFunc("/"+IDENTIFIER+"/{"+KEY+"}/{"+VALUE+"}", restGetProvisionWatchersByIdentifier).Methods(http.MethodGet)

}
func loadAddressableRoutes(b *mux.Router) {
	// /api/v1/" + ADDRESSABLE + "
	b.HandleFunc("/"+ADDRESSABLE, restGetAllAddressables).Methods(http.MethodGet)
	b.HandleFunc("/"+ADDRESSABLE, restAddAddressable).Methods(http.MethodPost)
	b.HandleFunc("/"+ADDRESSABLE, restUpdateAddressable).Methods(http.MethodPut)

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
func loadCommandRoutes(b *mux.Router) {

	// /api/v1/command
	b.HandleFunc("/"+COMMAND, restGetAllCommands).Methods(http.MethodGet)
	b.HandleFunc("/"+COMMAND, restAddCommand).Methods(http.MethodPost)
	b.HandleFunc("/"+COMMAND, restUpdateCommand).Methods(http.MethodPut)

	c := b.PathPrefix("/" + COMMAND).Subrouter()
	c.HandleFunc("/{"+ID+"}", restGetCommandById).Methods(http.MethodGet)
	c.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteCommandById).Methods(http.MethodDelete)
	c.HandleFunc("/"+NAME+"/{"+NAME+"}", restGetCommandByName).Methods(http.MethodGet)
	//c.HandleFunc("/" + NAME + "/{" + NAME + "}", restDeleteCommandByName).Methods(http.MethodDelete)
}
func ping(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}
