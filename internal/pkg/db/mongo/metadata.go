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
package mongo

import (
	"errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo/models"
	"github.com/google/uuid"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
)

/* -----------------------Schedule Event ------------------------*/
func (m *MongoClient) UpdateScheduleEvent(se contract.ScheduleEvent) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.ScheduleEvent)

	se.Modified = db.MakeTimestamp()

	// Handle DBRefs
	mse := mongoScheduleEvent{ScheduleEvent: se}

	return col.UpdateId(se.Id, mse)
}

func (m *MongoClient) AddScheduleEvent(se *contract.ScheduleEvent) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.ScheduleEvent)
	count, err := col.Find(bson.M{"name": se.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return db.ErrNotUnique
	}
	ts := db.MakeTimestamp()
	se.Created = ts
	se.Modified = ts
	se.Id = bson.NewObjectId()

	// Handle DBRefs
	mse := mongoScheduleEvent{ScheduleEvent: *se}

	return col.Insert(mse)
}

func (m *MongoClient) GetAllScheduleEvents(se *[]contract.ScheduleEvent) error {
	return m.GetScheduleEvents(se, bson.M{})
}

func (m *MongoClient) GetScheduleEventByName(se *contract.ScheduleEvent, n string) error {
	return m.GetScheduleEvent(se, bson.M{"name": n})
}

func (m *MongoClient) GetScheduleEventById(se *contract.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetScheduleEvent(se, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleEventById Invalid Object ID " + id)
		return err
	}
}

func (m *MongoClient) GetScheduleEventsByScheduleName(se *[]contract.ScheduleEvent, n string) error {
	return m.GetScheduleEvents(se, bson.M{"schedule": n})
}

func (m *MongoClient) GetScheduleEventsByAddressableId(se *[]contract.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetScheduleEvents(se, bson.M{"addressable" + ".$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleEventsByAddressableId Invalid Object ID" + id)
		return err
	}
}

func (m *MongoClient) GetScheduleEventsByServiceName(se *[]contract.ScheduleEvent, n string) error {
	return m.GetScheduleEvents(se, bson.M{"service": n})
}

func (m *MongoClient) GetScheduleEvent(se *contract.ScheduleEvent, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.ScheduleEvent)

	// Handle DBRef
	var mse mongoScheduleEvent

	err := col.Find(q).One(&mse)
	if err != nil {
		return errorMap(err)
	}
	*se = mse.ScheduleEvent
	return nil
}

func (m *MongoClient) GetScheduleEvents(se *[]contract.ScheduleEvent, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.ScheduleEvent)

	// Handle the DBRef
	var mses []mongoScheduleEvent

	err := col.Find(q).Sort("queryts").All(&mses)
	if err != nil {
		return err
	}

	*se = []contract.ScheduleEvent{}
	for _, mse := range mses {
		*se = append(*se, mse.ScheduleEvent)
	}

	return nil
}

func (m *MongoClient) DeleteScheduleEventById(id string) error {
	return m.deleteById(db.ScheduleEvent, id)
}

//  --------------------------Schedule ---------------------------*/
func (m *MongoClient) GetAllSchedules(s *[]contract.Schedule) error {
	return m.GetSchedules(s, bson.M{})
}

func (m *MongoClient) GetScheduleByName(s *contract.Schedule, n string) error {
	return m.GetSchedule(s, bson.M{"name": n})
}

func (m *MongoClient) GetScheduleById(s *contract.Schedule, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetSchedule(s, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleById Invalid Object ID " + id)
		return err
	}
}

func (m *MongoClient) AddSchedule(sch *contract.Schedule) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Schedule)
	count, err := col.Find(bson.M{"name": sch.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return db.ErrNotUnique
	}

	ts := db.MakeTimestamp()
	sch.Created = ts
	sch.Modified = ts
	sch.Id = bson.NewObjectId()
	return col.Insert(sch)
}

func (m *MongoClient) UpdateSchedule(sch contract.Schedule) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Schedule)

	sch.Modified = db.MakeTimestamp()

	if err := col.UpdateId(sch.Id, sch); err != nil {
		return err
	}

	return nil
}

func (m *MongoClient) DeleteScheduleById(id string) error {
	return m.deleteById(db.Schedule, id)
}

func (m *MongoClient) GetSchedule(sch *contract.Schedule, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Schedule)
	err := col.Find(q).One(sch)
	return errorMap(err)
}

func (m *MongoClient) GetSchedules(sch *[]contract.Schedule, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Schedule)
	return col.Find(q).Sort("queryts").All(sch)
}

/* ----------------------Device Report --------------------------*/
func (m *MongoClient) GetAllDeviceReports(d *[]contract.DeviceReport) error {
	return m.GetDeviceReports(d, bson.M{})
}

func (m *MongoClient) GetDeviceReportByName(d *contract.DeviceReport, n string) error {
	return m.GetDeviceReport(d, bson.M{"name": n})
}

func (m *MongoClient) GetDeviceReportByDeviceName(d *[]contract.DeviceReport, n string) error {
	return m.GetDeviceReports(d, bson.M{"device": n})
}

func (m *MongoClient) GetDeviceReportById(d *contract.DeviceReport, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetDeviceReport(d, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceReportById Invalid Object ID " + id)
		return err
	}
}

func (m *MongoClient) GetDeviceReportsByScheduleEventName(d *[]contract.DeviceReport, n string) error {
	return m.GetDeviceReports(d, bson.M{"event": n})
}

func (m *MongoClient) GetDeviceReports(d *[]contract.DeviceReport, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceReport)
	return col.Find(q).Sort("queryts").All(d)
}

func (m *MongoClient) GetDeviceReport(d *contract.DeviceReport, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceReport)
	err := col.Find(q).One(d)
	return errorMap(err)
}

func (m *MongoClient) AddDeviceReport(d *contract.DeviceReport) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceReport)
	count, err := col.Find(bson.M{"name": d.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return db.ErrNotUnique
	}
	ts := db.MakeTimestamp()
	d.Created = ts
	d.Id = bson.NewObjectId()
	return col.Insert(d)
}

func (m *MongoClient) UpdateDeviceReport(dr *contract.DeviceReport) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceReport)

	return col.UpdateId(dr.Id, dr)
}

func (m *MongoClient) DeleteDeviceReportById(id string) error {
	return m.deleteById(db.DeviceReport, id)
}

/* ----------------------------- Device ---------------------------------- */
func (m *MongoClient) AddDevice(d *contract.Device) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Device)

	// Check if the name exist (Device names must be unique)
	count, err := col.Find(bson.M{"name": d.Name}).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return db.ErrNotUnique
	}
	ts := db.MakeTimestamp()
	d.Created = ts
	d.Modified = ts
	d.Id = bson.NewObjectId()

	// Wrap the device in MongoDevice (For DBRefs)
	md := mongoDevice{Device: *d}

	return col.Insert(md)
}

func (m *MongoClient) UpdateDevice(rd contract.Device) error {
	s := m.session.Copy()
	defer s.Close()
	c := s.DB(m.database.Name).C(db.Device)

	// Copy over the DBRefs
	md := mongoDevice{Device: rd}

	return c.UpdateId(rd.Id, md)
}

func (m *MongoClient) DeleteDeviceById(id string) error {
	return m.deleteById(db.Device, id)
}

func (m *MongoClient) GetAllDevices(d *[]contract.Device) error {
	return m.GetDevices(d, nil)
}

func (m *MongoClient) GetDevicesByProfileId(d *[]contract.Device, pid string) error {
	if bson.IsObjectIdHex(pid) {
		return m.GetDevices(d, bson.M{"profile.$id": bson.ObjectIdHex(pid)})
	} else {
		err := errors.New("mgoGetDevicesByProfileId Invalid Object ID " + pid)
		return err
	}
}

func (m *MongoClient) GetDeviceById(d *contract.Device, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetDevice(d, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceById Invalid Object ID " + id)
		return err
	}
}

func (m *MongoClient) GetDeviceByName(d *contract.Device, n string) error {
	return m.GetDevice(d, bson.M{"name": n})
}

func (m *MongoClient) GetDevicesByServiceId(d *[]contract.Device, sid string) error {
	if bson.IsObjectIdHex(sid) {
		return m.GetDevices(d, bson.M{"service.$id": bson.ObjectIdHex(sid)})
	} else {
		err := errors.New("mgoGetDevicesByServiceId Invalid Object ID " + sid)
		return err
	}
}

func (m *MongoClient) GetDevicesByAddressableId(d *[]contract.Device, aid string) error {
	if bson.IsObjectIdHex(aid) {
		return m.GetDevices(d, bson.M{"addressable.$id": bson.ObjectIdHex(aid)})
	} else {
		err := errors.New("mgoGetDevicesByAddressableId Invalid Object ID " + aid)
		return err
	}
}

func (m *MongoClient) GetDevicesWithLabel(d *[]contract.Device, l string) error {
	var ls []string
	ls = append(ls, l)
	return m.GetDevices(d, bson.M{"labels": bson.M{"$in": ls}})
}

func (m *MongoClient) GetDevices(d *[]contract.Device, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Device)
	mds := []mongoDevice{}

	err := col.Find(q).Sort("queryts").All(&mds)
	if err != nil {
		return err
	}

	*d = []contract.Device{}
	for _, md := range mds {
		*d = append(*d, md.Device)
	}

	return nil
}

func (m *MongoClient) GetDevice(d *contract.Device, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Device)
	md := mongoDevice{}

	err := col.Find(q).One(&md)
	if err != nil {
		return errorMap(err)
	}
	*d = md.Device
	return nil
}

/* -----------------------------Device Profile -----------------------------*/
func (m *MongoClient) GetDeviceProfileById(d *contract.DeviceProfile, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetDeviceProfile(d, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceProfileById Invalid Object ID " + id)
		return err
	}
}

func (m *MongoClient) GetAllDeviceProfiles(dp *[]contract.DeviceProfile) error {
	return m.GetDeviceProfiles(dp, nil)
}

func (m *MongoClient) GetDeviceProfilesByModel(dp *[]contract.DeviceProfile, model string) error {
	return m.GetDeviceProfiles(dp, bson.M{"model": model})
}

func (m *MongoClient) GetDeviceProfilesWithLabel(dp *[]contract.DeviceProfile, l string) error {
	var ls []string
	ls = append(ls, l)
	return m.GetDeviceProfiles(dp, bson.M{"labels": bson.M{"$in": ls}})
}
func (m *MongoClient) GetDeviceProfilesByManufacturerModel(dp *[]contract.DeviceProfile, man string, mod string) error {
	return m.GetDeviceProfiles(dp, bson.M{"manufacturer": man, "model": mod})
}
func (m *MongoClient) GetDeviceProfilesByManufacturer(dp *[]contract.DeviceProfile, man string) error {
	return m.GetDeviceProfiles(dp, bson.M{"manufacturer": man})
}
func (m *MongoClient) GetDeviceProfileByName(dp *contract.DeviceProfile, n string) error {
	return m.GetDeviceProfile(dp, bson.M{"name": n})
}

// Get device profiles with the passed query
func (m *MongoClient) GetDeviceProfiles(d *[]contract.DeviceProfile, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceProfile)

	// Handle the DBRefs
	var mdps []mongoDeviceProfile
	err := col.Find(q).Sort("queryts").All(&mdps)
	if err != nil {
		return err
	}

	*d = []contract.DeviceProfile{}
	for _, mdp := range mdps {
		*d = append(*d, mdp.DeviceProfile)
	}

	return err
}

// Get device profile with the passed query
func (m *MongoClient) GetDeviceProfile(d *contract.DeviceProfile, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceProfile)

	// Handle the DBRefs
	var mdp mongoDeviceProfile
	err := col.Find(q).One(&mdp)
	if err != nil {
		return errorMap(err)
	}
	*d = mdp.DeviceProfile
	return nil
}

func (m *MongoClient) AddDeviceProfile(dp *contract.DeviceProfile) error {
	s := m.session.Copy()
	defer s.Close()
	if len(dp.Name) == 0 {
		return db.ErrNameEmpty
	}
	col := s.DB(m.database.Name).C(db.DeviceProfile)
	count, err := col.Find(bson.M{"name": dp.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return db.ErrNotUnique
	}
	for i := 0; i < len(dp.Commands); i++ {
		if err := m.AddCommand(&dp.Commands[i]); err != nil {
			return err
		}
	}
	ts := db.MakeTimestamp()
	dp.Created = ts
	dp.Modified = ts
	dp.Id = bson.NewObjectId()

	mdp := mongoDeviceProfile{DeviceProfile: *dp}

	return col.Insert(mdp)
}

func (m *MongoClient) UpdateDeviceProfile(dp *contract.DeviceProfile) error {
	s := m.session.Copy()
	defer s.Close()
	c := s.DB(m.database.Name).C(db.DeviceProfile)

	mdp := mongoDeviceProfile{DeviceProfile: *dp}
	mdp.Modified = db.MakeTimestamp()

	return c.UpdateId(mdp.Id, mdp)
}

// Get the device profiles that are currently using the command
func (m *MongoClient) GetDeviceProfilesUsingCommand(dp *[]contract.DeviceProfile, c contract.Command) error {
	var item bson.M
	if !bson.IsObjectIdHex(c.Id) {
		// EventID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(c.Id)
		if err != nil { // It is some unsupported type of string
			return db.ErrInvalidObjectId
		}
		item = bson.M{"uuid": c.Id}
	} else {
		item = bson.M{"$id": bson.ObjectIdHex(c.Id)}
	}
	query := bson.M{"commands": bson.M{"$elemMatch": item}}
	return m.GetDeviceProfiles(dp, query)
}

func (m *MongoClient) DeleteDeviceProfileById(id string) error {
	return m.deleteById(db.DeviceProfile, id)
}

//  -----------------------------------Addressable --------------------------*/
func (m *MongoClient) UpdateAddressable(ra *contract.Addressable, r *contract.Addressable) error {
	s := m.session.Copy()

	defer s.Close()
	c := s.DB(m.database.Name).C(db.Addressable)
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

func (m *MongoClient) GetAddressables(d *[]contract.Addressable) error {
	return m.GetAddressablesQuery(d, bson.M{})
}

func (m *MongoClient) GetAddressablesQuery(d *[]contract.Addressable, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Addressable)
	err := col.Find(q).Sort("queryts").All(d)
	if err != nil {
		return err
	}

	return nil
}

func (m *MongoClient) GetAddressableById(a *contract.Addressable, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetAddressable(a, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetAddressableById Invalid Object ID " + id)
		return err
	}
}

func (m *MongoClient) AddAddressable(a *contract.Addressable) (bson.ObjectId, error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Addressable)

	// check if the name exist
	count, err := col.Find(bson.M{"name": a.Name}).Count()
	if err != nil {
		return a.Id, err
	} else if count > 0 {
		return a.Id, db.ErrNotUnique
	}

	ts := db.MakeTimestamp()
	a.Created = ts
	a.Id = bson.NewObjectId()
	err = col.Insert(a)
	return a.Id, err
}

func (m *MongoClient) GetAddressableByName(a *contract.Addressable, n string) error {
	return m.GetAddressable(a, bson.M{"name": n})
}

func (m *MongoClient) GetAddressablesByTopic(a *[]contract.Addressable, t string) error {
	return m.GetAddressablesQuery(a, bson.M{"topic": t})
}

func (m *MongoClient) GetAddressablesByPort(a *[]contract.Addressable, p int) error {
	return m.GetAddressablesQuery(a, bson.M{"port": p})
}

func (m *MongoClient) GetAddressablesByPublisher(a *[]contract.Addressable, p string) error {
	return m.GetAddressablesQuery(a, bson.M{"publisher": p})
}

func (m *MongoClient) GetAddressablesByAddress(a *[]contract.Addressable, add string) error {
	return m.GetAddressablesQuery(a, bson.M{"address": add})
}

func (m *MongoClient) GetAddressable(d *contract.Addressable, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Addressable)
	err := col.Find(q).One(d)
	return errorMap(err)
}

func (m *MongoClient) DeleteAddressableById(id string) error {
	return m.deleteById(db.Addressable, id)
}

/* ----------------------------- Device Service ----------------------------------*/
func (m *MongoClient) GetDeviceServiceByName(d *contract.DeviceService, n string) error {
	return m.GetDeviceService(d, bson.M{"name": n})
}

func (m *MongoClient) GetDeviceServiceById(d *contract.DeviceService, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetDeviceService(d, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceServiceByName Invalid Object ID " + id)
		return err
	}
}

func (m *MongoClient) GetAllDeviceServices(d *[]contract.DeviceService) error {
	return m.GetDeviceServices(d, bson.M{})
}

func (m *MongoClient) GetDeviceServicesByAddressableId(d *[]contract.DeviceService, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetDeviceServices(d, bson.M{"addressable.$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceServicesByAddressableId Invalid Object ID " + id)
		return err
	}
}

func (m *MongoClient) GetDeviceServicesWithLabel(d *[]contract.DeviceService, l string) error {
	var ls []string
	ls = append(ls, l)
	return m.GetDeviceServices(d, bson.M{"labels": bson.M{"$in": ls}})
}

func (m *MongoClient) GetDeviceServices(d *[]contract.DeviceService, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceService)
	mdss := []mongoDeviceService{}
	err := col.Find(q).Sort("queryts").All(&mdss)
	if err != nil {
		return err
	}
	*d = []contract.DeviceService{}
	for _, mds := range mdss {
		*d = append(*d, mds.DeviceService)
	}
	return nil
}

func (m *MongoClient) GetDeviceService(d *contract.DeviceService, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceService)
	mds := mongoDeviceService{}
	err := col.Find(q).One(&mds)
	if err != nil {
		return errorMap(err)
	}
	*d = mds.DeviceService
	return nil
}

func (m *MongoClient) AddDeviceService(d *contract.DeviceService) error {
	s := m.session.Copy()
	defer s.Close()

	col := s.DB(m.database.Name).C(db.DeviceService)

	ts := db.MakeTimestamp()
	d.Created = ts
	d.Modified = ts
	d.Id = bson.NewObjectId()

	// mongoDeviceService handles the DBRefs
	mds := mongoDeviceService{DeviceService: *d}
	return col.Insert(mds)
}

func (m *MongoClient) UpdateDeviceService(deviceService contract.DeviceService) error {
	s := m.session.Copy()
	defer s.Close()
	c := s.DB(m.database.Name).C(db.DeviceService)

	deviceService.Modified = db.MakeTimestamp()

	// Handle DBRefs
	mds := mongoDeviceService{DeviceService: deviceService}

	return c.UpdateId(deviceService.Id, mds)
}

func (m *MongoClient) DeleteDeviceServiceById(id string) error {
	return m.deleteById(db.DeviceService, id)
}

//  ----------------------Provision Watcher -----------------------------*/
func (m *MongoClient) GetAllProvisionWatchers(pw *[]contract.ProvisionWatcher) error {
	return m.GetProvisionWatchers(pw, bson.M{})
}

func (m *MongoClient) GetProvisionWatcherByName(pw *contract.ProvisionWatcher, n string) error {
	return m.GetProvisionWatcher(pw, bson.M{"name": n})
}

func (m *MongoClient) GetProvisionWatchersByIdentifier(pw *[]contract.ProvisionWatcher, k string, v string) error {
	return m.GetProvisionWatchers(pw, bson.M{"identifiers." + k: v})
}

func (m *MongoClient) GetProvisionWatchersByServiceId(pw *[]contract.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetProvisionWatchers(pw, bson.M{"service.$id": bson.ObjectIdHex(id)})
	} else {
		return errors.New("mgoGetProvisionWatchersByServiceId Invalid Object ID " + id)
	}
}

func (m *MongoClient) GetProvisionWatchersByProfileId(pw *[]contract.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetProvisionWatchers(pw, bson.M{"profile.$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetProvisionWatcherByProfileId Invalid Object ID " + id)
		return err
	}
}

func (m *MongoClient) GetProvisionWatcherById(pw *contract.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetProvisionWatcher(pw, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetProvisionWatcherById Invalid Object ID " + id)
		return err
	}
}

func (m *MongoClient) GetProvisionWatcher(pw *contract.ProvisionWatcher, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.ProvisionWatcher)

	// Handle DBRefs
	var mpw mongoProvisionWatcher

	err := col.Find(q).One(&mpw)
	if err != nil {
		return errorMap(err)
	}
	*pw = mpw.ProvisionWatcher
	return nil
}

func (m *MongoClient) GetProvisionWatchers(pw *[]contract.ProvisionWatcher, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.ProvisionWatcher)

	// Handle DBRefs
	var mpws []mongoProvisionWatcher

	err := col.Find(q).Sort("queryts").All(&mpws)
	if err != nil {
		return err
	}

	*pw = []contract.ProvisionWatcher{}
	for _, mpw := range mpws {
		*pw = append(*pw, mpw.ProvisionWatcher)
	}

	return nil
}

func (m *MongoClient) AddProvisionWatcher(pw *contract.ProvisionWatcher) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.ProvisionWatcher)
	count, err := col.Find(bson.M{"name": pw.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return db.ErrNotUnique
	}

	// get Device Service
	var dev contract.DeviceService
	if pw.Service.Id.Hex() != "" {
		m.GetDeviceServiceById(&dev, pw.Service.Id.Hex())
	} else if pw.Service.Name != "" {
		m.GetDeviceServiceByName(&dev, pw.Service.Name)
	} else {
		return errors.New("Device Service ID or Name is required")
	}
	pw.Service = dev

	// get Device Profile
	var dp contract.DeviceProfile
	if pw.Profile.Id.Hex() != "" {
		m.GetDeviceProfileById(&dp, pw.Profile.Id.Hex())
	} else if pw.Profile.Name != "" {
		m.GetDeviceProfileByName(&dp, pw.Profile.Name)
	} else {
		return errors.New("Device Profile ID or Name is required")
	}
	pw.Profile = dp

	// Set data
	ts := db.MakeTimestamp()
	pw.Created = ts
	pw.Modified = ts
	pw.Id = bson.NewObjectId()

	// Handle DBRefs
	mpw := mongoProvisionWatcher{ProvisionWatcher: *pw}

	return col.Insert(mpw)
}

func (m *MongoClient) UpdateProvisionWatcher(pw contract.ProvisionWatcher) error {
	s := m.session.Copy()
	defer s.Close()
	c := s.DB(m.database.Name).C(db.ProvisionWatcher)

	pw.Modified = db.MakeTimestamp()

	// Handle DBRefs
	mpw := mongoProvisionWatcher{ProvisionWatcher: pw}

	return c.UpdateId(mpw.Id, mpw)
}

func (m *MongoClient) DeleteProvisionWatcherById(id string) error {
	return m.deleteById(db.ProvisionWatcher, id)
}

//  ------------------------Command -------------------------------------*/
func (m *MongoClient) GetAllCommands(d *[]contract.Command) error {
	return m.GetCommands(d, bson.M{})
}

func (m *MongoClient) GetCommands(d *[]contract.Command, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Command)
	return col.Find(q).Sort("queryts").All(d)
}

func (m *MongoClient) GetCommandById(c *contract.Command, id string) error {
	if bson.IsObjectIdHex(id) {
		s := m.session.Copy()
		defer s.Close()

		col := s.DB(m.database.Name).C(db.Command)

		var query bson.M
		if !bson.IsObjectIdHex(id) {
			// EventID is not a BSON ID. Is it a UUID?
			_, err := uuid.Parse(id)
			if err != nil { // It is some unsupported type of string
				return db.ErrInvalidObjectId
			}
			query = bson.M{"uuid": id}
		} else {
			query = bson.M{"_id": bson.ObjectIdHex(id)}
		}

		command := &models.Command{}
		err := col.Find(query).One(command)
		*c = command.ToContract()

		return err
	} else {
		return errors.New("mgoGetCommandById Invalid Object ID " + id)
	}
}

func (m *MongoClient) GetCommandByName(c *[]contract.Command, n string) error {
	s := m.session.Copy()
	defer s.Close()

	col := s.DB(m.database.Name).C(db.Command)
	commands := &[]models.Command{}
	err := col.Find(bson.M{"name": n}).All(commands)

	*c, err = mapCommands(*commands, err)
	return err
}

func mapCommands(commands []models.Command, err error) ([]contract.Command, error) {
	if err != nil {
		return []contract.Command{}, errorMap(err)
	}

	var mapped []contract.Command
	for _, cmd := range commands {
		mapped = append(mapped, cmd.ToContract())
	}
	return mapped, nil
}

func (m *MongoClient) AddCommand(c *contract.Command) error {
	s := m.session.Copy()
	defer s.Close()

	col := s.DB(m.database.Name).C(db.Command)

	ts := db.MakeTimestamp()
	c.Created = ts
	c.Id = uuid.New().String()
	return col.Insert(c)
}

// Update command uses the ID of the command for identification
func (m *MongoClient) UpdateCommand(c *contract.Command, r *contract.Command) error {
	s := m.session.Copy()
	defer s.Close()

	col := s.DB(m.database.Name).C(db.Command)
	if c == nil {
		return nil
	}

	var id interface{}
	var err error
	if bson.IsObjectIdHex(c.Id) {
		id = bson.ObjectIdHex(c.Id)
	} else {
		id, err = uuid.Parse(c.Id)
		if err != nil { // It is some unsupported type of string
			return errors.New("Id required for updating a command")
		}
	}

	// Update the fields
	if c.Name != "" {
		r.Name = c.Name
	}
	// TODO check for Get and Put Equality

	if (c.Get.String() != contract.Get{}.String()) {
		r.Get = c.Get
	}
	if (c.Put.String() != contract.Put{}.String()) {
		r.Put = c.Put
	}
	if c.Origin != 0 {
		r.Origin = c.Origin
	}

	return col.UpdateId(id, r)
}

// Delete the command by ID
// Check if the command is still in use by device profiles
func (m *MongoClient) DeleteCommandById(id string) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Command)

	if !bson.IsObjectIdHex(id) {
		return errors.New("Invalid ID")
	}

	// Check if the command is still in use
	query := bson.M{"commands": bson.M{"$elemMatch": bson.M{"_id": bson.ObjectIdHex(id)}}}
	count, err := s.DB(m.database.Name).C(db.DeviceProfile).Find(query).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return db.ErrCommandStillInUse
	}

	return col.RemoveId(bson.ObjectIdHex(id))
}

// Scrub all metadata
func (m *MongoClient) ScrubMetadata() error {
	s := m.session.Copy()
	defer s.Close()

	_, err := s.DB(m.database.Name).C(db.Addressable).RemoveAll(nil)
	if err != nil {
		return err
	}

	_, err = s.DB(m.database.Name).C(db.DeviceService).RemoveAll(nil)
	if err != nil {
		return err
	}
	_, err = s.DB(m.database.Name).C(db.DeviceProfile).RemoveAll(nil)
	if err != nil {
		return err
	}
	_, err = s.DB(m.database.Name).C(db.DeviceReport).RemoveAll(nil)
	if err != nil {
		return err
	}
	_, err = s.DB(m.database.Name).C(db.ScheduleEvent).RemoveAll(nil)
	if err != nil {
		return err
	}
	_, err = s.DB(m.database.Name).C(db.Device).RemoveAll(nil)
	if err != nil {
		return err
	}
	_, err = s.DB(m.database.Name).C(db.ProvisionWatcher).RemoveAll(nil)
	if err != nil {
		return err
	}

	return nil
}
