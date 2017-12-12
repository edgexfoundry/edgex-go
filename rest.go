package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func loadRestRoutes() http.Handler {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()
	b.HandleFunc("/ping", ping)

	loadDeviceRoutes(b)
	//loadDeviceManagerRoutes(b)
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
	b.HandleFunc("/"+DEVICE, restAddNewDevice).Methods(POST)
	b.HandleFunc("/"+DEVICE, restUpdateDevice).Methods(PUT)
	b.HandleFunc("/"+DEVICE, restGetAllDevices).Methods(GET)

	d := b.PathPrefix("/" + DEVICE).Subrouter()

	d.HandleFunc("/"+LABEL+"/{"+LABEL+"}", restGetDevicesWithLabel).Methods(GET)
	d.HandleFunc("/"+PROFILE+"/{"+PROFILEID+"}", restGetDeviceByProfileId).Methods(GET)
	d.HandleFunc("/"+SERVICE+"/{"+SERVICEID+"}", restGetDeviceByServiceId).Methods(GET)
	d.HandleFunc("/"+SERVICENAME+"/{"+SERVICENAME+"}", restGetDeviceByServiceName).Methods(GET)
	d.HandleFunc("/"+ADDRESSABLENAME+"/{"+ADDRESSABLENAME+"}", restGetDeviceByAddressableName).Methods(GET)
	d.HandleFunc("/"+PROFILENAME+"/{"+PROFILENAME+"}", restGetDeviceByProfileName).Methods(GET)
	d.HandleFunc("/"+ADDRESSABLE+"/{"+ADDRESSABLEID+"}", restGetDeviceByAddressableId).Methods(GET)

	// /api/v1/" + DEVICE" + ID + "
	d.HandleFunc("/{"+ID+"}", restGetDeviceById).Methods(GET)
	d.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteDeviceById).Methods(DELETE)
	d.HandleFunc("/{"+ID+"}/"+OPSTATE+"/{"+OPSTATE+"}", restSetDeviceOpStateById).Methods(PUT)
	d.HandleFunc("/{"+ID+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restSetDeviceAdminStateById).Methods(PUT)
	d.HandleFunc("/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", restSetDeviceLastReportedById).Methods(PUT)
	d.HandleFunc("/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}/{"+LASTREPORTEDNOTIFY+"}", restSetDeviceLastReportedByIdNotify).Methods(PUT)
	d.HandleFunc("/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", restSetDeviceLastConnectedById).Methods(PUT)
	d.HandleFunc("/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}/{"+LASTCONNECTEDNOTIFY+"}", restSetLastConnectedByIdNotify).Methods(PUT)

	// /api/v1/" + DEVICE/" + NAME + "
	n := d.PathPrefix("/" + NAME).Subrouter()
	n.HandleFunc("/{"+NAME+"}", restGetDeviceByName).Methods(GET)
	n.HandleFunc("/{"+NAME+"}", restDeleteDeviceByName).Methods(DELETE)
	n.HandleFunc("/{"+NAME+"}/"+OPSTATE+"/{"+OPSTATE+"}", restSetDeviceOpStateByName).Methods(PUT)
	n.HandleFunc("/{"+NAME+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restSetDeviceAdminStateByName).Methods(PUT)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", restSetDeviceLastReportedByName).Methods(PUT)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}/{"+LASTREPORTEDNOTIFY+"}", restSetDeviceLastReportedByNameNotify).Methods(PUT)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", restSetDeviceLastConnectedByName).Methods(PUT)
	n.HandleFunc("/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}/{"+LASTCONNECTEDNOTIFY+"}", restSetDeviceLastConnectedByNameNotify).Methods(PUT)
}

/*func loadDeviceManagerRoutes(b *mux.Router) {
	// /api/v1/" + DEVICEMANAGER
	b.HandleFunc("/" + DEVICEMANAGER, restGetAllDeviceManagers).Methods(GET)
	b.HandleFunc("/" + DEVICEMANAGER, restAddDeviceManager).Methods(POST)
	b.HandleFunc("/" + DEVICEMANAGER, restUpdateDeviceManager).Methods(PUT)
	dm := b.PathPrefix("/" + DEVICEMANAGER).Subrouter()
	dm.HandleFunc("/" + ID + "/{" + ID + "}", restDeleteDeviceManagerById).Methods(DELETE)
	dm.HandleFunc("/" + ADDRESSABLENAME + "/{" + ADDRESSABLENAME + "}", restGetDeviceManagersByAddressableName).Methods(GET)
	dm.HandleFunc("/" + ADDRESSABLE + "/{" + ADDRESSABLEID + "}", restGetDeviceManagerByAddressableId).Methods(GET)
	dm.HandleFunc("/{" + ID + "}", restGetDeviceManagerById).Methods(GET)
	dm.HandleFunc("/{" + ID + "}/" + OPSTATE + "/{" + OPSTATE + "}", restUpdateDeviceManagerOpStateById).Methods(PUT)
	dm.HandleFunc("/{" + ID + "}/" + URLADMINSTATE + "/{" + ADMINSTATE + "}", restUpdateDeviceManagerAdminStateById).Methods(PUT)
	dm.HandleFunc("/{" + ID + "}/" + URLLASTREPORTED + "/{" + LASTREPORTED + "}", restUpdateDeviceManagerLastReportedById).Methods(PUT)
	dm.HandleFunc("/{" + ID + "}/" + URLLASTCONNECTED + "/{" + LASTCONNECTED + "}", restUpdateDeviceManagerLastConnectedById).Methods(PUT)
	dm.HandleFunc("/" + PROFILENAME + "/{" + PROFILENAME + "}", restGetDeviceManagersByProfileName).Methods(GET)
	dm.HandleFunc("/" + PROFILE + "/{" + PROFILEID + "}", restGetDeviceManagersByProfileId).Methods(GET)
	// /api/v1/" + DEVICEMANAGER/" + NAME + "
	dmn := dm.PathPrefix("/" + NAME).Subrouter()
	dmn.HandleFunc("/{" + NAME + "}", restDeleteDeviceManagerByName).Methods(DELETE)
	dmn.HandleFunc("/{" + NAME + "}", restGetDeviceManagerByName).Methods(GET)
	dmn.HandleFunc("/{" + NAME + "}/" + OPSTATE + "/{" + OPSTATE + "}", restUpdateDeviceManagerOpStateByName).Methods(PUT)
	dmn.HandleFunc("/{" + NAME + "}/" + URLADMINSTATE + "/{" + ADMINSTATE + "}", restUpdateDeviceManagerAdminStateByName).Methods(PUT)
	dmn.HandleFunc("/{" + NAME + "}/" + URLLASTREPORTED + "/{" + LASTREPORTED + "}", restUpdateDeviceManagerLastReportedByName).Methods(PUT)
	dmn.HandleFunc("/{" + NAME + "}/" + URLLASTCONNECTED + "/{" + LASTCONNECTED + "}", restUpdateDeviceManagerLastConnectedByName).Methods(PUT)

	// /api/v1/DEVICEMANAGER/SERVICENAME
	dm.HandleFunc("/" + SERVICENAME + "/{" + SERVICENAME + "}", restGetDeviceManagersByServiceName).Methods(GET)
	dm.HandleFunc("/" + SERVICE + "/{" + SERVICEID + "}", restGetDeviceManagersByServiceId).Methods(GET)

	// /api/v1/" + DEVICEMANAGER/" + SERVICE + "
	dm.HandleFunc("/" + LABEL+ "/{" + LABEL + "}", restGetDeviceManagerByLabel).Methods(GET)
}*/
func loadDeviceProfileRoutes(b *mux.Router) {
	///api/v1/" + DEVICEPROFILE + "
	b.HandleFunc("/"+DEVICEPROFILE+"", restGetAllDeviceProfiles).Methods(GET)
	b.HandleFunc("/"+DEVICEPROFILE+"", restAddDeviceProfile).Methods(POST)
	b.HandleFunc("/"+DEVICEPROFILE+"", restUpdateDeviceProfile).Methods(PUT)

	dp := b.PathPrefix("/" + DEVICEPROFILE).Subrouter()
	dp.HandleFunc("/{"+ID+"}", restGetProfileByProfileId).Methods(GET)
	dp.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteProfileByProfileId).Methods(DELETE)
	dp.HandleFunc("/"+UPLOADFILE, restAddProfileByYaml).Methods(POST)
	dp.HandleFunc("/"+UPLOAD, restAddProfileByYamlRaw).Methods(POST)
	dp.HandleFunc("/"+MODEL+"/{"+MODEL+"}", restGetProfileByModel).Methods(GET)
	dp.HandleFunc("/"+LABEL+"/{"+LABEL+"}", restGetProfileWithLabel).Methods(GET)

	// /api/v1/" + DEVICEPROFILE + "/"  + MANUFACTURER + "
	dpm := dp.PathPrefix("/" + MANUFACTURER).Subrouter()
	dpm.HandleFunc("/{"+MANUFACTURER+"}/"+MODEL+"/{"+MODEL+"}", restGetProfileByManufacturerModel).Methods(GET)
	dpm.HandleFunc("/{"+MANUFACTURER+"}", restGetProfileByManufacturer).Methods(GET)

	// /api/v1/" + DEVICEPROFILE + "/" + NAME + "
	dpn := dp.PathPrefix("/" + NAME).Subrouter()
	dpn.HandleFunc("/{"+NAME+"}", restGetProfileByName).Methods(GET)
	dpn.HandleFunc("/{"+NAME+"}", restDeleteProfileByName).Methods(DELETE)

	// /api/v1/" + DEVICEPROFILE + "/"  + YAML
	dpy := dp.PathPrefix("/" + YAML).Subrouter()
	// TODO add functionality
	dpy.HandleFunc("/"+NAME+"/{"+NAME+"}", restGetYamlProfileByName).Methods(GET)
	dpy.HandleFunc("/{"+ID+"}", restGetYamlProfileById).Methods(GET)
}
func loadDeviceReportRoutes(b *mux.Router) {
	// /api/v1/devicereport
	b.HandleFunc("/"+DEVICEREPORT, restGetAllDeviceReports).Methods(GET)
	b.HandleFunc("/"+DEVICEREPORT, restAddDeviceReport).Methods(POST)
	b.HandleFunc("/"+DEVICEREPORT, restUpdateDeviceReport).Methods(PUT)

	dr := b.PathPrefix("/" + DEVICEREPORT).Subrouter()
	dr.HandleFunc("/{"+ID+"}", restGetReportById).Methods(GET)
	dr.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteReportById).Methods(DELETE)
	dr.HandleFunc("/"+DEVICENAME+"/{"+DEVICENAME+"}", restGetDeviceReportByDeviceName).Methods(GET)

	// /api/v1/devicereport/" + NAME + "
	drn := dr.PathPrefix("/" + NAME).Subrouter()
	drn.HandleFunc("/{"+NAME+"}", restGetReportByName).Methods(GET)
	drn.HandleFunc("/{"+NAME+"}", restDeleteReportByName).Methods(DELETE)

	// /api/v1/devicereport/valueDescriptorsFor/devicename
	drvd := dr.PathPrefix("/" + VALUEDESCRIPTORSFOR).Subrouter()
	drvd.HandleFunc("/{"+DEVICENAME+"}", restGetValueDescriptorsForDeviceName).Methods(GET)
}
func loadDeviceServiceRoutes(b *mux.Router) {
	// /api/v1/deviceservice
	b.HandleFunc("/"+DEVICESERVICE, restGetAllDeviceServices).Methods(GET)
	b.HandleFunc("/"+DEVICESERVICE, restAddDeviceService).Methods(POST)
	b.HandleFunc("/"+DEVICESERVICE, restUpdateDeviceService).Methods(PUT)

	ds := b.PathPrefix("/" + DEVICESERVICE).Subrouter()
	ds.HandleFunc("/"+ADDRESSABLENAME+"/{"+ADDRESSABLENAME+"}", restGetServiceByAddressableName).Methods(GET)
	ds.HandleFunc("/"+ADDRESSABLE+"/{"+ADDRESSABLEID+"}", restGetServiceByAddressableId).Methods(GET)
	ds.HandleFunc("/"+LABEL+"/{"+LABEL+"}", restGetServiceWithLabel).Methods(GET)
	ds.HandleFunc("/"+DEVICEADDRESSABLES+"/{"+ID+"}", restGetAddressablesForAssociatedDevicesById).Methods(GET)
	ds.HandleFunc("/"+DEVICEADDRESSABLESBYNAME+"/{"+NAME+"}", restGetAddressablesForAssociatedDevicesByName).Methods(GET)

	// /api/v1/deviceservice/" + NAME + "
	dsn := ds.PathPrefix("/" + NAME).Subrouter()
	dsn.HandleFunc("/{"+NAME+"}", restGetServiceByName).Methods(GET)
	dsn.HandleFunc("/{"+NAME+"}", restDeleteServiceByName).Methods(DELETE)
	dsn.HandleFunc("/{"+NAME+"}/"+OPSTATE+"/{"+OPSTATE+"}", restUpdateServiceOpStateByName).Methods(PUT)
	dsn.HandleFunc("/{"+NAME+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restUpdateServiceAdminStateByName).Methods(PUT)
	dsn.HandleFunc("/{"+NAME+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", restUpdateServiceLastReportedByName).Methods(PUT)
	dsn.HandleFunc("/{"+NAME+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", restUpdateServiceLastConnectedByName).Methods(PUT)

	// /api/v1/"  + DEVICESERVICE + ID + "
	ds.HandleFunc("/{"+ID+"}", restGetServiceById).Methods(GET)
	ds.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteServiceById).Methods(DELETE)
	ds.HandleFunc("/{"+ID+"}/"+OPSTATE+"/{"+OPSTATE+"}", restUpdateServiceOpStateById).Methods(PUT)
	ds.HandleFunc("/{"+ID+"}/"+URLADMINSTATE+"/{"+ADMINSTATE+"}", restUpdateServiceAdminStateById).Methods(PUT)
	ds.HandleFunc("/{"+ID+"}/"+URLLASTREPORTED+"/{"+LASTREPORTED+"}", restUpdateServiceLastReportedById).Methods(PUT)
	ds.HandleFunc("/{"+ID+"}/"+URLLASTCONNECTED+"/{"+LASTCONNECTED+"}", restUpdateServiceLastConnectedById).Methods(PUT)
}
func loadScheduleEventRoutes(b *mux.Router) {
	// /api/v1/scheduleevent
	b.HandleFunc("/"+SCHEDULEEVENT, restGetAllScheduleEvents).Methods(GET)
	b.HandleFunc("/"+SCHEDULEEVENT, restAddScheduleEvent).Methods(POST)
	b.HandleFunc("/"+SCHEDULEEVENT, restUpdateScheduleEvent).Methods(PUT)
	se := b.PathPrefix("/" + SCHEDULEEVENT).Subrouter()
	se.HandleFunc("/{"+ID+"}", restGetScheduleEventById).Methods(GET)

	// /api/v1/scheduleevent/" + NAME + "
	sen := se.PathPrefix("/" + NAME + "").Subrouter()
	sen.HandleFunc("/{"+NAME+"}", restDeleteScheduleEventByName).Methods(DELETE)
	sen.HandleFunc("/{"+NAME+"}", restGetScheduleEventByName).Methods(GET)

	// /api/v1/"  + SCHEDULEEVENT + ID + "
	seid := se.PathPrefix("/" + ID).Subrouter()
	seid.HandleFunc("/{"+ID+"}", restDeleteScheduleEventById).Methods(DELETE)

	// /api/v1/scheduleevent/addressable
	seaid := se.PathPrefix("/" + ADDRESSABLE).Subrouter()
	seaid.HandleFunc("/{"+ADDRESSABLEID+"}", restGetScheduleEventByAddressableId).Methods(GET)

	sean := se.PathPrefix("/" + ADDRESSABLENAME).Subrouter()
	sean.HandleFunc("/{"+ADDRESSABLENAME+"}", restGetScheduleEventByAddressableName).Methods(GET)

	// /api/v1/scheduleevent/servicename
	sesn := se.PathPrefix("/" + SERVICENAME).Subrouter()
	sesn.HandleFunc("/{"+SERVICENAME+"}", restGetScheduleEventsByServiceName).Methods(GET)
}
func loadScheduleRoutes(b *mux.Router) {
	// /api/v1/schedule
	b.HandleFunc("/"+SCHEDULE, restGetAllSchedules).Methods(GET)
	b.HandleFunc("/"+SCHEDULE, restAddSchedule).Methods(POST)
	b.HandleFunc("/"+SCHEDULE, restUpdateSchedule).Methods(PUT)
	sch := b.PathPrefix("/" + SCHEDULE).Subrouter()
	sch.HandleFunc("/{"+ID+"}", restGetScheduleById).Methods(GET)

	// /api/v1/schedule/" + NAME + "
	schn := sch.PathPrefix("/" + NAME + "").Subrouter()
	schn.HandleFunc("/{"+NAME+"}", restGetScheduleByName).Methods(GET)
	schn.HandleFunc("/{"+NAME+"}", restDeleteScheduleByName).Methods(DELETE)

	// /api/v1/"  + SCHEDULE + ID + "
	schid := sch.PathPrefix("/" + ID).Subrouter()
	schid.HandleFunc("/{"+ID+"}", restDeleteScheduleById).Methods(DELETE)
}
func loadProvisionWatcherRoutes(b *mux.Router) {
	b.HandleFunc("/"+PROVISIONWATCHER, restAddProvisionWatcher).Methods(POST)
	b.HandleFunc("/"+PROVISIONWATCHER, restUpdateProvisionWatcher).Methods(PUT)
	b.HandleFunc("/"+PROVISIONWATCHER, restGetProvisionWatchers).Methods(GET)
	pw := b.PathPrefix("/" + PROVISIONWATCHER).Subrouter()
	// /api/v1/provisionwatcher
	pw.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteProvisionWatcherById).Methods(DELETE)
	pw.HandleFunc("/{"+ID+"}", restGetProvisionWatcherById).Methods(GET)
	pw.HandleFunc("/"+NAME+"/{"+NAME+"}", restDeleteProvisionWatcherByName).Methods(DELETE)
	pw.HandleFunc("/"+NAME+"/{"+NAME+"}", restGetProvisionWatcherByName).Methods(GET)
	pw.HandleFunc("/"+PROFILENAME+"/{"+NAME+"}", restGetProvisionWatchersByProfileName).Methods(GET)
	pw.HandleFunc("/"+PROFILE+"/{"+ID+"}", restGetProvisionWatchersByProfileId).Methods(GET)
	pw.HandleFunc("/"+SERVICE+"/{"+ID+"}", restGetProvisionWatchersByServiceId).Methods(GET)
	pw.HandleFunc("/"+SERVICENAME+"/{"+NAME+"}", restGetProvisionWatchersByServiceName).Methods(GET)
	pw.HandleFunc("/"+IDENTIFIER+"/{"+KEY+"}/{"+VALUE+"}", restGetProvisionWatchersByIdentifier).Methods(GET)

}
func loadAddressableRoutes(b *mux.Router) {
	// /api/v1/" + ADDRESSABLE + "
	b.HandleFunc("/"+ADDRESSABLE, restGetAllAddressables).Methods(GET)
	b.HandleFunc("/"+ADDRESSABLE, restAddAddressable).Methods(POST)
	b.HandleFunc("/"+ADDRESSABLE, restUpdateAddressable).Methods(PUT)

	a := b.PathPrefix("/" + ADDRESSABLE).Subrouter()
	a.HandleFunc("/{"+ID+"}", restGetAddressableById).Methods(GET)
	a.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteAddressableById).Methods(DELETE)
	a.HandleFunc("/"+NAME+"/{"+NAME+"}", restGetAddressableByName).Methods(GET)
	a.HandleFunc("/"+NAME+"/{"+NAME+"}", restDeleteAddressableByName).Methods(DELETE)
	a.HandleFunc("/"+TOPIC+"/{"+TOPIC+"}", restGetAddressableByTopic).Methods(GET)
	a.HandleFunc("/"+PORT+"/{"+PORT+"}", restGetAddressableByPort).Methods(GET)
	a.HandleFunc("/"+PUBLISHER+"/{"+PUBLISHER+"}", restGetAddressableByPublisher).Methods(GET)
	a.HandleFunc("/"+ADDRESS+"/{"+ADDRESS+"}", restGetAddressableByAddress).Methods(GET)
}
func loadCommandRoutes(b *mux.Router) {

	// /api/v1/command
	b.HandleFunc("/"+COMMAND, restGetAllCommands).Methods(GET)
	b.HandleFunc("/"+COMMAND, restAddCommand).Methods(POST)
	b.HandleFunc("/"+COMMAND, restUpdateCommand).Methods(PUT)

	c := b.PathPrefix("/" + COMMAND).Subrouter()
	c.HandleFunc("/{"+ID+"}", restGetCommandById).Methods(GET)
	c.HandleFunc("/"+ID+"/{"+ID+"}", restDeleteCommandById).Methods(DELETE)
	c.HandleFunc("/"+NAME+"/{"+NAME+"}", restGetCommandByName).Methods(GET)
	//c.HandleFunc("/" + NAME + "/{" + NAME + "}", restDeleteCommandByName).Methods(DELETE)
}
func ping(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}
