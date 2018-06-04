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
	"errors"
	"time"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

type mongoDB struct {
	s *mgo.Session
}

func getMongoSessionCopy() *mgo.Session {
	m, ok := db.(*mongoDB)
	if ok {
		return m.s.Copy()
	} else {
		return nil
	}
}

func (m *mongoDB) Connect() error {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{DOCKERMONGO},
		Timeout:  time.Duration(configuration.MongoDBConnectTimeout) * time.Millisecond,
		Database: MONGODATABASE,
		Username: DBUSER,
		Password: DBPASS,
	}
	var err error
	m.s, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return err
	}

	// Set timeout based on configuration
	m.s.SetSocketTimeout(time.Duration(configuration.MongoDBConnectTimeout) * time.Millisecond)
	return nil
}

func (m *mongoDB) CloseSession() {
	if m.s != nil {
		m.s.Close()
		m.s = nil
	}
}

/* -----------------------Schedule Event ------------------------*/
func (m *mongoDB) updateScheduleEvent(se models.ScheduleEvent) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(SECOL)

	se.Modified = makeTimestamp()

	// Handle DBRefs
	mse := MongoScheduleEvent{ScheduleEvent: se}

	return col.UpdateId(se.Id, mse)
}

func (m *mongoDB) addScheduleEvent(se *models.ScheduleEvent) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(SECOL)
	count, err := col.Find(bson.M{NAME: se.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return ErrDuplicateName
	}
	ts := makeTimestamp()
	se.Created = ts
	se.Modified = ts
	se.Id = bson.NewObjectId()

	// Handle DBRefs
	mse := MongoScheduleEvent{ScheduleEvent: *se}

	return col.Insert(mse)
}

func (m *mongoDB) getAllScheduleEvents(se *[]models.ScheduleEvent) error {
	return m.getScheduleEvents(se, bson.M{})
}

func (m *mongoDB) getScheduleEventByName(se *models.ScheduleEvent, n string) error {
	return m.getScheduleEvent(se, bson.M{NAME: n})
}

func (m *mongoDB) getScheduleEventById(se *models.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getScheduleEvent(se, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleEventById Invalid Object ID " + id)
		return err
	}
}

func (m *mongoDB) getScheduleEventsByScheduleName(se *[]models.ScheduleEvent, n string) error {
	return m.getScheduleEvents(se, bson.M{SCHEDULE: n})
}

func (m *mongoDB) getScheduleEventsByAddressableId(se *[]models.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getScheduleEvents(se, bson.M{ADDRESSABLE + ".$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleEventsByAddressableId Invalid Object ID" + id)
		return err
	}
}

func (m *mongoDB) getScheduleEventsByServiceName(se *[]models.ScheduleEvent, n string) error {
	return m.getScheduleEvents(se, bson.M{SERVICE: n})
}

func (m *mongoDB) getScheduleEvent(se *models.ScheduleEvent, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(SECOL)

	// Handle DBRef
	var mse MongoScheduleEvent

	err := col.Find(q).One(&mse)
	if err != nil {
		return err
	}

	*se = mse.ScheduleEvent

	return err
}

func (m *mongoDB) getScheduleEvents(se *[]models.ScheduleEvent, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(SECOL)

	// Handle the DBRef
	var mses []MongoScheduleEvent

	err := col.Find(q).Sort(QUERYTS).All(&mses)
	if err != nil {
		return err
	}

	for _, mse := range mses {
		*se = append(*se, mse.ScheduleEvent)
	}

	return nil
}

func (m *mongoDB) deleteScheduleEvent(se models.ScheduleEvent) error {
	return m.deleteById(SECOL, se.Id.Hex())
}

//  --------------------------Schedule ---------------------------*/
func (m *mongoDB) getAllSchedules(s *[]models.Schedule) error {
	return m.getSchedules(s, bson.M{})
}

func (m *mongoDB) getScheduleByName(s *models.Schedule, n string) error {
	return m.getSchedule(s, bson.M{NAME: n})
}

func (m *mongoDB) getScheduleById(s *models.Schedule, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getSchedule(s, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleById Invalid Object ID " + id)
		return err
	}
}

func (m *mongoDB) addSchedule(sch *models.Schedule) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(SCOL)
	count, err := col.Find(bson.M{NAME: sch.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		err := errors.New("Schedule already exist")
		return err
	}

	ts := makeTimestamp()
	sch.Created = ts
	sch.Modified = ts
	sch.Id = bson.NewObjectId()
	return col.Insert(s)
}

func (m *mongoDB) updateSchedule(sch models.Schedule) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(SCOL)

	sch.Modified = makeTimestamp()

	if err := col.UpdateId(sch.Id, sch); err != nil {
		return err
	}

	return nil
}

func (m *mongoDB) deleteSchedule(s models.Schedule) error {
	return m.deleteById(SCOL, s.Id.Hex())
}

func (m *mongoDB) getSchedule(sch *models.Schedule, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(SCOL)
	return col.Find(q).One(sch)
}

func (m *mongoDB) getSchedules(sch *[]models.Schedule, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(SCOL)
	return col.Find(q).Sort(QUERYTS).All(sch)
}

/* ----------------------Device Report --------------------------*/
func (m *mongoDB) getAllDeviceReports(d *[]models.DeviceReport) error {
	return m.getDeviceReports(d, bson.M{})
}

func (m *mongoDB) getDeviceReportByName(d *models.DeviceReport, n string) error {
	return m.getDeviceReport(d, bson.M{NAME: n})
}

func (m *mongoDB) getDeviceReportByDeviceName(d *[]models.DeviceReport, n string) error {
	return m.getDeviceReports(d, bson.M{DEVICE: n})
}

func (m *mongoDB) getDeviceReportById(d *models.DeviceReport, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getDeviceReport(d, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceReportById Invalid Object ID " + id)
		return err
	}
}

func (m *mongoDB) getDeviceReportsByScheduleEventName(d *[]models.DeviceReport, n string) error {
	return m.getDeviceReports(d, bson.M{"event": n})
}

func (m *mongoDB) getDeviceReports(d *[]models.DeviceReport, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DRCOL)
	return col.Find(q).Sort(QUERYTS).All(d)
}

func (m *mongoDB) getDeviceReport(d *models.DeviceReport, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DRCOL)
	return col.Find(q).One(d)
}

func (m *mongoDB) addDeviceReport(d *models.DeviceReport) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DRCOL)
	count, err := col.Find(bson.M{NAME: d.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return ErrDuplicateName
	}
	ts := makeTimestamp()
	d.Created = ts
	d.Id = bson.NewObjectId()
	return col.Insert(d)
}

func (m *mongoDB) updateDeviceReport(dr *models.DeviceReport) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DRCOL)

	return col.UpdateId(dr.Id, dr)
}

func (m *mongoDB) deleteDeviceReport(dr models.DeviceReport) error {
	return m.deleteById(DRCOL, dr.Id.Hex())
}

/* ----------------------------- Device ---------------------------------- */
func (m *mongoDB) addDevice(d *models.Device) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DEVICECOL)

	// Check if the name exist (Device names must be unique)
	count, err := col.Find(bson.M{NAME: d.Name}).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrDuplicateName
	}
	ts := makeTimestamp()
	d.Created = ts
	d.Modified = ts
	d.Id = bson.NewObjectId()

	// Wrap the device in MongoDevice (For DBRefs)
	md := MongoDevice{Device: *d}

	return col.Insert(md)
}

func (m *mongoDB) updateDevice(rd models.Device) error {
	s := m.s.Copy()
	defer s.Close()
	c := s.DB(DB).C(DEVICECOL)

	// Copy over the DBRefs
	md := MongoDevice{Device: rd}

	return c.UpdateId(rd.Id, md)
}

func (m *mongoDB) deleteDevice(d models.Device) error {
	return m.deleteById(DEVICECOL, d.Id.Hex())
}

func (m *mongoDB) getAllDevices(d *[]models.Device) error {
	return m.getDevices(d, nil)
}

func (m *mongoDB) getDevicesByProfileId(d *[]models.Device, pid string) error {
	if bson.IsObjectIdHex(pid) {
		return m.getDevices(d, bson.M{PROFILE + "." + "$id": bson.ObjectIdHex(pid)})
	} else {
		err := errors.New("mgoGetDevicesByProfileId Invalid Object ID " + pid)
		return err
	}
}

func (m *mongoDB) getDeviceById(d *models.Device, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getDevice(d, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceById Invalid Object ID " + id)
		return err
	}
}

func (m *mongoDB) getDeviceByName(d *models.Device, n string) error {
	return m.getDevice(d, bson.M{NAME: n})
}

func (m *mongoDB) getDevicesByServiceId(d *[]models.Device, sid string) error {
	if bson.IsObjectIdHex(sid) {
		return m.getDevices(d, bson.M{SERVICE + "." + "$id": bson.ObjectIdHex(sid)})
	} else {
		err := errors.New("mgoGetDevicesByServiceId Invalid Object ID " + sid)
		return err
	}
}

func (m *mongoDB) getDevicesByAddressableId(d *[]models.Device, aid string) error {
	if bson.IsObjectIdHex(aid) {
		// Check if the addressable exists
		var a *models.Addressable
		if m.getAddressableById(a, aid) == mgo.ErrNotFound {
			return mgo.ErrNotFound
		}
		return m.getDevices(d, bson.M{ADDRESSABLE + "." + "$id": bson.ObjectIdHex(aid)})
	} else {
		err := errors.New("mgoGetDevicesByAddressableId Invalid Object ID " + aid)
		return err
	}
}

func (m *mongoDB) getDevicesWithLabel(d *[]models.Device, l []string) error {
	return m.getDevices(d, bson.M{LABELS: bson.M{"$in": l}})
}

func (m *mongoDB) getDevices(d *[]models.Device, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DEVICECOL)
	mds := []MongoDevice{}

	err := col.Find(q).Sort(QUERYTS).All(&mds)
	if err != nil {
		return err
	}

	for _, md := range mds {
		*d = append(*d, md.Device)
	}

	return nil
}

func (m *mongoDB) getDevice(d *models.Device, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DEVICECOL)
	md := MongoDevice{}

	err := col.Find(q).One(&md)
	if err != nil {
		return err
	}

	*d = md.Device

	return nil
}

/* -----------------------------Device Profile -----------------------------*/
func (m *mongoDB) getDeviceProfileById(d *models.DeviceProfile, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getDeviceProfile(d, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceProfileById Invalid Object ID " + id)
		return err
	}
}

func (m *mongoDB) getAllDeviceProfiles(dp *[]models.DeviceProfile) error {
	return m.getDeviceProfiles(dp, nil)
}

func (m *mongoDB) getDeviceProfilesByModel(dp *[]models.DeviceProfile, model string) error {
	return m.getDeviceProfiles(dp, bson.M{MODEL: model})
}

func (m *mongoDB) getDeviceProfilesWithLabel(dp *[]models.DeviceProfile, l []string) error {
	return m.getDeviceProfiles(dp, bson.M{LABELS: bson.M{"$in": l}})
}
func (m *mongoDB) getDeviceProfilesByManufacturerModel(dp *[]models.DeviceProfile, man string, mod string) error {
	return m.getDeviceProfiles(dp, bson.M{MANUFACTURER: man, MODEL: mod})
}
func (m *mongoDB) getDeviceProfilesByManufacturer(dp *[]models.DeviceProfile, man string) error {
	return m.getDeviceProfiles(dp, bson.M{MANUFACTURER: man})
}
func (m *mongoDB) getDeviceProfileByName(dp *models.DeviceProfile, n string) error {
	return m.getDeviceProfile(dp, bson.M{NAME: n})
}

// Get device profiles with the passed query
func (m *mongoDB) getDeviceProfiles(d *[]models.DeviceProfile, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DPCOL)

	// Handle the DBRefs
	var mdps []MongoDeviceProfile
	err := col.Find(q).Sort(QUERYTS).All(&mdps)
	if err != nil {
		return err
	}

	for _, mdp := range mdps {
		*d = append(*d, mdp.DeviceProfile)
	}

	return err
}

// Get device profile with the passed query
func (m *mongoDB) getDeviceProfile(d *models.DeviceProfile, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DPCOL)

	// Handle the DBRefs
	var mdp MongoDeviceProfile
	err := col.Find(q).One(&mdp)
	if err != nil {
		return err
	}

	*d = mdp.DeviceProfile

	return err
}

func (m *mongoDB) addDeviceProfile(dp *models.DeviceProfile) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DPCOL)
	count, err := col.Find(bson.M{NAME: dp.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return ErrDuplicateName
	}
	for i := 0; i < len(dp.Commands); i++ {
		if err := m.addCommand(&dp.Commands[i]); err != nil {
			return err
		}
	}
	ts := makeTimestamp()
	dp.Created = ts
	dp.Modified = ts
	dp.Id = bson.NewObjectId()

	mdp := MongoDeviceProfile{DeviceProfile: *dp}

	return col.Insert(mdp)
}

func (m *mongoDB) updateDeviceProfile(dp *models.DeviceProfile) error {
	s := m.s.Copy()
	defer s.Close()
	c := s.DB(DB).C(DPCOL)

	mdp := MongoDeviceProfile{DeviceProfile: *dp}
	mdp.Modified = makeTimestamp()

	return c.UpdateId(mdp.Id, mdp)
}

// Get the device profiles that are currently using the command
func (m *mongoDB) getDeviceProfilesUsingCommand(dp *[]models.DeviceProfile, c models.Command) error {
	query := bson.M{"commands": bson.M{"$elemMatch": bson.M{"$id": c.Id}}}
	return m.getDeviceProfiles(dp, query)
}

func (m *mongoDB) deleteDeviceProfile(dp models.DeviceProfile) error {
	return m.deleteById(DPCOL, dp.Id.Hex())
}

//  -----------------------------------Addressable --------------------------*/
func (m *mongoDB) updateAddressable(ra *models.Addressable, r *models.Addressable) error {
	s := m.s.Copy()

	defer s.Close()
	c := s.DB(DB).C(ADDCOL)
	if ra == nil {
		return nil
	}
	if ra.Name != "" {
		r.Name = ra.Name
	}
	if ra.Protocol != "" {
		r.Protocol = ra.Protocol
	}
	if ra.Address != "" {
		r.Address = ra.Address
	}
	if ra.Port != int(0) {
		r.Port = ra.Port
	}
	if ra.Path != "" {
		r.Path = ra.Path
	}
	if ra.Publisher != "" {
		r.Publisher = ra.Publisher
	}
	if ra.User != "" {
		r.User = ra.User
	}
	if ra.Password != "" {
		r.Password = ra.Password
	}
	if ra.Topic != "" {
		r.Topic = ra.Topic
	}
	if err := c.UpdateId(r.Id, r); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) getAddressables(d *[]models.Addressable) error {
	return m.getAddressablesQuery(d, bson.M{})
}

func (m *mongoDB) getAddressablesQuery(d *[]models.Addressable, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(ADDCOL)
	err := col.Find(q).Sort(QUERYTS).All(d)
	if err != nil {
		return err
	}

	return nil
}

func (m *mongoDB) getAddressableById(a *models.Addressable, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getAddressable(a, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetAddressableById Invalid Object ID " + id)
		return err
	}
}

func (m *mongoDB) addAddressable(a *models.Addressable) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(ADDCOL)

	// check if the name exist
	count, err := col.Find(bson.M{NAME: a.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return ErrDuplicateName
	}

	ts := makeTimestamp()
	a.Created = ts
	a.Id = bson.NewObjectId()
	err = col.Insert(a)
	if err != nil {
		return err
	}

	return nil
}

func (m *mongoDB) getAddressableByName(a *models.Addressable, n string) error {
	return m.getAddressable(a, bson.M{NAME: n})
}

func (m *mongoDB) getAddressablesByTopic(a *[]models.Addressable, t string) error {
	return m.getAddressablesQuery(a, bson.M{TOPIC: t})
}

func (m *mongoDB) getAddressablesByPort(a *[]models.Addressable, p int) error {
	return m.getAddressablesQuery(a, bson.M{PORT: p})
}

func (m *mongoDB) getAddressablesByPublisher(a *[]models.Addressable, p string) error {
	return m.getAddressablesQuery(a, bson.M{PUBLISHER: p})
}

func (m *mongoDB) getAddressablesByAddress(a *[]models.Addressable, add string) error {
	return m.getAddressablesQuery(a, bson.M{ADDRESS: add})
}

func (m *mongoDB) getAddressable(d *models.Addressable, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(ADDCOL)
	err := col.Find(q).One(d)
	if err != nil {
		return err
	}

	return nil
}

func (m *mongoDB) deleteAddressable(a models.Addressable) error {
	return m.deleteById(ADDCOL, a.Id.Hex())
}

/* ----------------------------- Device Service ----------------------------------*/
func (m *mongoDB) getDeviceServiceByName(d *models.DeviceService, n string) error {
	return m.getDeviceService(d, bson.M{NAME: n})
}

func (m *mongoDB) getDeviceServiceById(d *models.DeviceService, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getDeviceService(d, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceServiceByName Invalid Object ID " + id)
		return err
	}
}

func (m *mongoDB) getAllDeviceServices(d *[]models.DeviceService) error {
	return m.getDeviceServices(d, bson.M{})
}

func (m *mongoDB) getDeviceServicesByAddressableId(d *[]models.DeviceService, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getDeviceServices(d, bson.M{ADDRESSABLE + ".$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceServicesByAddressableId Invalid Object ID " + id)
		return err
	}
}

func (m *mongoDB) getDeviceServicesWithLabel(d *[]models.DeviceService, l []string) error {
	return m.getDeviceServices(d, bson.M{LABELS: bson.M{"$in": l}})
}

func (m *mongoDB) getDeviceServices(d *[]models.DeviceService, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DSCOL)
	mdss := []MongoDeviceService{}
	err := col.Find(q).Sort(QUERYTS).All(&mdss)
	if err != nil {
		return err
	}
	for _, mds := range mdss {
		*d = append(*d, mds.DeviceService)
	}

	return nil
}

func (m *mongoDB) getDeviceService(d *models.DeviceService, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DSCOL)
	mds := MongoDeviceService{}
	err := col.Find(q).One(&mds)
	if err != nil {
		return err
	}
	*d = mds.DeviceService

	return nil
}

func (m *mongoDB) addDeviceService(d *models.DeviceService) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(DSCOL)

	// check if the name exist
	count, err := col.Find(bson.M{NAME: d.Service.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return ErrDuplicateName
	}

	ts := makeTimestamp()
	d.Service.Created = ts
	d.Service.Modified = ts
	d.Service.Id = bson.NewObjectId()

	// MongoDeviceService handles the DBRefs
	mds := MongoDeviceService{DeviceService: *d}
	return col.Insert(mds)
}

func (m *mongoDB) updateDeviceService(deviceService models.DeviceService) error {
	s := m.s.Copy()
	defer s.Close()
	c := s.DB(DB).C(DSCOL)

	deviceService.Service.Modified = makeTimestamp()

	// Handle DBRefs
	mds := MongoDeviceService{DeviceService: deviceService}

	return c.UpdateId(deviceService.Service.Id, mds)
}

func (m *mongoDB) deleteDeviceService(ds models.DeviceService) error {
	return m.deleteById(DSCOL, ds.Id.Hex())
}

//  ----------------------Provision Watcher -----------------------------*/
func (m *mongoDB) getAllProvisionWatchers(pw *[]models.ProvisionWatcher) error {
	return m.getProvisionWatchers(pw, bson.M{})
}

func (m *mongoDB) getProvisionWatcherByName(pw *models.ProvisionWatcher, n string) error {
	return m.getProvisionWatcher(pw, bson.M{NAME: n})
}

func (m *mongoDB) getProvisionWatchersByIdentifier(pw *[]models.ProvisionWatcher, k string, v string) error {
	return m.getProvisionWatchers(pw, bson.M{IDENTIFIERS + "." + k: v})
}

func (m *mongoDB) getProvisionWatchersByServiceId(pw *[]models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getProvisionWatchers(pw, bson.M{SERVICE + ".$id": bson.ObjectIdHex(id)})
	} else {
		return errors.New("mgoGetProvisionWatchersByServiceId Invalid Object ID " + id)
	}
}

func (m *mongoDB) getProvisionWatcherByProfileId(pw *[]models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getProvisionWatchers(pw, bson.M{PROFILE + ".$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetProvisionWatcherByProfileId Invalid Object ID " + id)
		return err
	}
}

func (m *mongoDB) getProvisionWatcherById(pw *models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.getProvisionWatcher(pw, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetProvisionWatcherById Invalid Object ID " + id)
		return err
	}
}

func (m *mongoDB) getProvisionWatcher(pw *models.ProvisionWatcher, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(PWCOL)

	// Handle DBRefs
	var mpw MongoProvisionWatcher

	err := col.Find(q).One(&mpw)
	if err != nil {
		return err
	}

	*pw = mpw.ProvisionWatcher

	return err
}

func (m *mongoDB) getProvisionWatchers(pw *[]models.ProvisionWatcher, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(PWCOL)

	// Handle DBRefs
	var mpws []MongoProvisionWatcher

	err := col.Find(q).Sort(QUERYTS).All(&mpws)
	if err != nil {
		return err
	}

	for _, mpw := range mpws {
		*pw = append(*pw, mpw.ProvisionWatcher)
	}

	return nil
}

func (m *mongoDB) addProvisionWatcher(pw *models.ProvisionWatcher) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(PWCOL)
	count, err := col.Find(bson.M{NAME: pw.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return ErrDuplicateName
	}

	// get Device Service
	var dev models.DeviceService
	if pw.Service.Service.Id.Hex() != "" {
		m.getDeviceServiceById(&dev, pw.Service.Service.Id.Hex())
	} else if pw.Service.Service.Name != "" {
		m.getDeviceServiceByName(&dev, pw.Service.Service.Name)
	} else {
		return errors.New("Device Service ID or Name is required")
	}
	pw.Service = dev

	// get Device Profile
	var dp models.DeviceProfile
	if pw.Profile.Id.Hex() != "" {
		m.getDeviceProfileById(&dp, pw.Profile.Id.Hex())
	} else if pw.Profile.Name != "" {
		m.getDeviceProfileByName(&dp, pw.Profile.Name)
	} else {
		return errors.New("Device Profile ID or Name is required")
	}
	pw.Profile = dp

	// Set data
	ts := makeTimestamp()
	pw.Created = ts
	pw.Modified = ts
	pw.Id = bson.NewObjectId()

	// Handle DBRefs
	mpw := MongoProvisionWatcher{ProvisionWatcher: *pw}

	return col.Insert(mpw)
}

func (m *mongoDB) updateProvisionWatcher(pw models.ProvisionWatcher) error {
	s := m.s.Copy()
	defer s.Close()
	c := s.DB(DB).C(PWCOL)

	pw.Modified = makeTimestamp()

	// Handle DBRefs
	mpw := MongoProvisionWatcher{ProvisionWatcher: pw}

	return c.UpdateId(mpw.Id, mpw)
}

func (m *mongoDB) deleteProvisionWatcher(pw models.ProvisionWatcher) error {
	return m.deleteById(PWCOL, pw.Id.Hex())
}

//  ------------------------Command -------------------------------------*/
func (m *mongoDB) getAllCommands(d *[]models.Command) error {
	return m.getCommands(d, bson.M{})
}

func (m *mongoDB) getCommands(d *[]models.Command, q bson.M) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(COMCOL)
	return col.Find(q).Sort(QUERYTS).All(d)
}

func (m *mongoDB) getCommandById(d *models.Command, id string) error {
	if bson.IsObjectIdHex(id) {
		s := m.s.Copy()
		defer s.Close()
		col := s.DB(DB).C(COMCOL)
		return col.Find(bson.M{_ID: bson.ObjectIdHex(id)}).One(d)
	} else {
		return errors.New("mgoGetCommandById Invalid Object ID " + id)
	}
}

func (m *mongoDB) getCommandByName(c *[]models.Command, n string) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(COMCOL)
	return col.Find(bson.M{NAME: n}).All(c)
}

func (m *mongoDB) addCommand(c *models.Command) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(COMCOL)

	ts := makeTimestamp()
	c.Created = ts
	c.Id = bson.NewObjectId()
	return col.Insert(c)
}

// Update command uses the ID of the command for identification
func (m *mongoDB) updateCommand(c *models.Command, r *models.Command) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(COMCOL)
	if c == nil {
		return nil
	}

	// Check if the command has a valid ID
	if len(c.Id.Hex()) == 0 || !bson.IsObjectIdHex(c.Id.Hex()) {
		return errors.New("ID required for updating a command")
	}

	// Update the fields
	if c.Name != "" {
		r.Name = c.Name
	}
	// TODO check for Get and Put Equality

	if (c.Get.String() != models.Get{}.String()) {
		r.Get = c.Get
	}
	if (c.Put.String() != models.Put{}.String()) {
		r.Put = c.Put
	}
	if c.Origin != 0 {
		r.Origin = c.Origin
	}

	return col.UpdateId(r.Id, r)
}

// Delete the command by ID
// Check if the command is still in use by device profiles
func (m *mongoDB) deleteCommandById(id string) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(COMCOL)

	if !bson.IsObjectIdHex(id) {
		return errors.New("Invalid ID")
	}

	// Check if the command is still in use
	query := bson.M{"commands": bson.M{"$elemMatch": bson.M{"_id": bson.ObjectIdHex(id)}}}
	count, err := s.DB(DB).C(DPCOL).Find(query).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrCommandStillInUse
	}

	return col.RemoveId(bson.ObjectIdHex(id))
}

func (m *mongoDB) deleteById(c string, did string) error {
	if bson.IsObjectIdHex(did) {
		s := m.s.Copy()
		defer s.Close()
		col := s.DB(DB).C(c)
		err := col.RemoveId(bson.ObjectIdHex(did))
		if err != nil {
			return err
		}

		return nil
	} else {
		err := errors.New("Invalid object ID " + did)
		return err
	}
}

func (m *mongoDB) deleteByName(c string, n string) error {
	s := m.s.Copy()
	defer s.Close()
	col := s.DB(DB).C(c)
	err := col.Remove(bson.M{NAME: n})
	if err != nil {
		return err
	}

	return nil
}
