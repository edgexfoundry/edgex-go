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
	"fmt"
	"github.com/globalsign/mgo"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo/models"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
)

/* -----------------------Schedule Event ------------------------*/
func (m MongoClient) UpdateScheduleEvent(se contract.ScheduleEvent) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.ScheduleEvent)

	se.Modified = db.MakeTimestamp()

	// Handle DBRefs
	mse := mongoScheduleEvent{ScheduleEvent: se}

	return col.UpdateId(se.Id, mse)
}

func (m MongoClient) AddScheduleEvent(se *contract.ScheduleEvent) error {
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

func (m MongoClient) GetAllScheduleEvents(se *[]contract.ScheduleEvent) error {
	return m.GetScheduleEvents(se, bson.M{})
}

func (m MongoClient) GetScheduleEventByName(se *contract.ScheduleEvent, n string) error {
	return m.GetScheduleEvent(se, bson.M{"name": n})
}

func (m MongoClient) GetScheduleEventById(se *contract.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetScheduleEvent(se, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleEventById Invalid Object ID " + id)
		return err
	}
}

func (m MongoClient) GetScheduleEventsByScheduleName(se *[]contract.ScheduleEvent, n string) error {
	return m.GetScheduleEvents(se, bson.M{"schedule": n})
}

func (m MongoClient) GetScheduleEventsByAddressableId(se *[]contract.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetScheduleEvents(se, bson.M{"addressable" + ".$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleEventsByAddressableId Invalid Object ID" + id)
		return err
	}
}

func (m MongoClient) GetScheduleEventsByServiceName(se *[]contract.ScheduleEvent, n string) error {
	return m.GetScheduleEvents(se, bson.M{"service": n})
}

func (m MongoClient) GetScheduleEvent(se *contract.ScheduleEvent, q bson.M) error {
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

func (m MongoClient) GetScheduleEvents(se *[]contract.ScheduleEvent, q bson.M) error {
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

func (m MongoClient) DeleteScheduleEventById(id string) error {
	return m.deleteById(db.ScheduleEvent, id)
}

//  --------------------------Schedule ---------------------------*/
func (m MongoClient) GetAllSchedules(s *[]contract.Schedule) error {
	return m.GetSchedules(s, bson.M{})
}

func (m MongoClient) GetScheduleByName(s *contract.Schedule, n string) error {
	return m.GetSchedule(s, bson.M{"name": n})
}

func (m MongoClient) GetScheduleById(s *contract.Schedule, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetSchedule(s, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleById Invalid Object ID " + id)
		return err
	}
}

func (m MongoClient) AddSchedule(sch *contract.Schedule) error {
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

func (m MongoClient) UpdateSchedule(sch contract.Schedule) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Schedule)

	sch.Modified = db.MakeTimestamp()

	if err := col.UpdateId(sch.Id, sch); err != nil {
		return err
	}

	return nil
}

func (m MongoClient) DeleteScheduleById(id string) error {
	return m.deleteById(db.Schedule, id)
}

func (m MongoClient) GetSchedule(sch *contract.Schedule, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Schedule)
	err := col.Find(q).One(sch)
	return errorMap(err)
}

func (m MongoClient) GetSchedules(sch *[]contract.Schedule, q bson.M) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Schedule)
	return col.Find(q).Sort("queryts").All(sch)
}

/* ----------------------Device Report --------------------------*/
func (m MongoClient) GetAllDeviceReports() ([]contract.DeviceReport, error) {
	return m.getDeviceReports(bson.M{})
}

func (m MongoClient) GetDeviceReportByName(n string) (contract.DeviceReport, error) {
	return m.getDeviceReport(bson.M{"name": n})
}

func (m MongoClient) GetDeviceReportByDeviceName(n string) ([]contract.DeviceReport, error) {
	return m.getDeviceReports(bson.M{"device": n})
}

func (m MongoClient) GetDeviceReportById(id string) (contract.DeviceReport, error) {
	var query bson.M
	if bson.IsObjectIdHex(id) {
		query = bson.M{"_id": bson.ObjectIdHex(id)}
	} else {
		_, err := uuid.Parse(id)
		if err == nil {
			query = bson.M{"uuid": id}
		} else {
			return contract.DeviceReport{}, errors.New("mgoGetDeviceReportById Invalid Object ID " + id)
		}
	}
	return m.getDeviceReport(query)
}

func (m MongoClient) GetDeviceReportsByScheduleEventName(n string) ([]contract.DeviceReport, error) {
	return m.getDeviceReports(bson.M{"event": n})
}

func (m MongoClient) getDeviceReports(q bson.M) ([]contract.DeviceReport, error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceReport)

	d := make([]models.DeviceReport, 0)
	err := col.Find(q).Sort("queryts").All(&d)
	return mapDeviceReports(d, err)
}

func mapDeviceReports(drs []models.DeviceReport, err error) ([]contract.DeviceReport, error) {
	if err != nil {
		return []contract.DeviceReport{}, err
	}

	var mapped []contract.DeviceReport
	for _, dr := range drs {
		mapped = append(mapped, dr.ToContract())
	}
	return mapped, nil
}

func (m MongoClient) getDeviceReport(q bson.M) (contract.DeviceReport, error) {
	s := m.session.Copy()
	defer s.Close()

	var d models.DeviceReport
	col := s.DB(m.database.Name).C(db.DeviceReport)
	err := col.Find(q).One(&d)
	return d.ToContract(), errorMap(err)
}

func (m MongoClient) AddDeviceReport(d contract.DeviceReport) (string, error) {
	s := m.session.Copy()
	defer s.Close()

	col := s.DB(m.database.Name).C(db.DeviceReport)
	count, err := col.Find(bson.M{"name": d.Name}).Count()
	if err != nil {
		return "", err
	} else if count > 0 {
		return "", db.ErrNotUnique
	}

	deviceReport := &models.DeviceReport{}
	if err := deviceReport.FromContract(d); err != nil {
		return "", errors.New("FromContract failed")
	}
	d = deviceReport.ToContract()

	err = col.Insert(deviceReport)
	return d.Id, err
}

func (m MongoClient) UpdateDeviceReport(dr contract.DeviceReport) error {
	s := m.session.Copy()
	defer s.Close()

	model := &models.DeviceReport{}
	if err := model.FromContract(dr); err != nil {
		return errors.New("FromContract failed")
	}

	var err error
	if model.Id.Valid() {
		err = s.DB(m.database.Name).C(db.DeviceReport).UpdateId(dr.Id, dr)
	} else {
		err = s.DB(m.database.Name).C(db.DeviceReport).Update(bson.M{"uuid": model.Uuid}, model)
	}
	return errorMap(err)
}

func (m MongoClient) DeleteDeviceReportById(id string) error {
	return m.deleteById(db.DeviceReport, id)
}

/* ----------------------------- Device ---------------------------------- */
func (m MongoClient) AddDevice(d *contract.Device) error {
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

	addr, err := m.getAddressableByName(d.Addressable.Name)
	if err != nil {
		return err
	}
	d.Addressable.Id = addr.Id.Hex()

	// Wrap the device in MongoDevice (For DBRefs)
	md := mongoDevice{Device: *d}

	return col.Insert(md)
}

func (m MongoClient) UpdateDevice(rd contract.Device) error {
	s := m.session.Copy()
	defer s.Close()
	c := s.DB(m.database.Name).C(db.Device)

	// Copy over the DBRefs
	md := mongoDevice{Device: rd}

	return c.UpdateId(rd.Id, md)
}

func (m MongoClient) DeleteDeviceById(id string) error {
	return m.deleteById(db.Device, id)
}

func (m MongoClient) GetAllDevices(d *[]contract.Device) error {
	return m.GetDevices(d, nil)
}

func (m MongoClient) GetDevicesByProfileId(d *[]contract.Device, pid string) error {
	if bson.IsObjectIdHex(pid) {
		return m.GetDevices(d, bson.M{"profile.$id": bson.ObjectIdHex(pid)})
	} else {
		err := errors.New("mgoGetDevicesByProfileId Invalid Object ID " + pid)
		return err
	}
}

func (m MongoClient) GetDeviceById(d *contract.Device, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetDevice(d, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceById Invalid Object ID " + id)
		return err
	}
}

func (m MongoClient) GetDeviceByName(d *contract.Device, n string) error {
	return m.GetDevice(d, bson.M{"name": n})
}

func (m MongoClient) GetDevicesByServiceId(d *[]contract.Device, sid string) error {
	if bson.IsObjectIdHex(sid) {
		return m.GetDevices(d, bson.M{"service.$id": bson.ObjectIdHex(sid)})
	} else {
		err := errors.New("mgoGetDevicesByServiceId Invalid Object ID " + sid)
		return err
	}
}

func (m MongoClient) GetDevicesByAddressableId(d *[]contract.Device, aid string) error {
	//Incoming addressable ID could be either BSON or JSON.
	//Figure out which one it is. If UUID, load the Mongo addressable model to obtain the BSON Id
	//because the contract won't have that.
	var query bson.M
	if bson.IsObjectIdHex(aid) {
		query = bson.M{"addressable.$id": bson.ObjectIdHex(aid)}
	} else {
		_, err := uuid.Parse(aid)
		if err == nil {
			addr, err := m.getAddressable(bson.M{"uuid": aid})
			if err != nil {
				return err
			}
			query = bson.M{"addressable.$id": addr.Id}
		} else {
			return errors.New("mgoGetDevicesByAddressableId Invalid Object ID " + aid)
		}
	}
	return m.GetDevices(d, bson.M{"addressable.$id": query})
}

func (m MongoClient) GetDevicesWithLabel(d *[]contract.Device, l string) error {
	var ls []string
	ls = append(ls, l)
	return m.GetDevices(d, bson.M{"labels": bson.M{"$in": ls}})
}

func (m MongoClient) GetDevices(d *[]contract.Device, q bson.M) error {
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

func (m MongoClient) GetDevice(d *contract.Device, q bson.M) error {
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
func (m MongoClient) GetDeviceProfileById(d *contract.DeviceProfile, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetDeviceProfile(d, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceProfileById Invalid Object ID " + id)
		return err
	}
}

func (m MongoClient) GetAllDeviceProfiles(dp *[]contract.DeviceProfile) error {
	return m.GetDeviceProfiles(dp, nil)
}

func (m MongoClient) GetDeviceProfilesByModel(dp *[]contract.DeviceProfile, model string) error {
	return m.GetDeviceProfiles(dp, bson.M{"model": model})
}

func (m MongoClient) GetDeviceProfilesWithLabel(dp *[]contract.DeviceProfile, l string) error {
	var ls []string
	ls = append(ls, l)
	return m.GetDeviceProfiles(dp, bson.M{"labels": bson.M{"$in": ls}})
}
func (m MongoClient) GetDeviceProfilesByManufacturerModel(dp *[]contract.DeviceProfile, man string, mod string) error {
	return m.GetDeviceProfiles(dp, bson.M{"manufacturer": man, "model": mod})
}
func (m MongoClient) GetDeviceProfilesByManufacturer(dp *[]contract.DeviceProfile, man string) error {
	return m.GetDeviceProfiles(dp, bson.M{"manufacturer": man})
}
func (m MongoClient) GetDeviceProfileByName(dp *contract.DeviceProfile, n string) error {
	return m.GetDeviceProfile(dp, bson.M{"name": n})
}

// Get device profiles with the passed query
func (m MongoClient) GetDeviceProfiles(d *[]contract.DeviceProfile, q bson.M) error {
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
func (m MongoClient) GetDeviceProfile(d *contract.DeviceProfile, q bson.M) error {
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

func (m MongoClient) AddDeviceProfile(dp *contract.DeviceProfile) error {
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
		if newId, err := m.AddCommand(dp.Commands[i]); err != nil {
			return err
		} else {
			dp.Commands[i].Id = newId
		}
	}
	ts := db.MakeTimestamp()
	dp.Created = ts
	dp.Modified = ts
	dp.Id = bson.NewObjectId()

	mdp := mongoDeviceProfile{DeviceProfile: *dp}

	return col.Insert(mdp)
}

func (m MongoClient) UpdateDeviceProfile(dp *contract.DeviceProfile) error {
	s := m.session.Copy()
	defer s.Close()
	c := s.DB(m.database.Name).C(db.DeviceProfile)

	mdp := mongoDeviceProfile{DeviceProfile: *dp}
	mdp.Modified = db.MakeTimestamp()

	return c.UpdateId(mdp.Id, mdp)
}

// Get the device profiles that are currently using the command
func (m MongoClient) GetDeviceProfilesUsingCommand(dp *[]contract.DeviceProfile, c contract.Command) error {
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

func (m MongoClient) DeleteDeviceProfileById(id string) error {
	return m.deleteById(db.DeviceProfile, id)
}

//  -----------------------------------Addressable --------------------------*/

// DBRefToAddressable converts DBRef to internal Mongo struct
func (m MongoClient) DBRefToAddressable(dbRef mgo.DBRef) (a models.Addressable, err error) {
	s := m.session.Copy()
	defer s.Close()

	if err = s.DB(m.database.Name).C(db.Addressable).Find(bson.M{"_id": dbRef.Id}).One(&a); err != nil {
		return models.Addressable{}, err
	}
	return
}

// AddressableToDBRef converts internal Mongo struct to DBRef
func (m MongoClient) AddressableToDBRef(a models.Addressable) (dbRef mgo.DBRef, err error) {
	s := m.session.Copy()
	defer s.Close()

	// validate addressable identity provided in contract actually exists and populate missing Id, Uuid field
	var addr models.Addressable
	if a.Id.Valid() {
		addr, err = m.getAddressableById(a.Id.Hex())
	} else {
		addr, err = m.getAddressableById(a.Uuid)
	}
	if err != nil {
		return
	}

	dbRef = mgo.DBRef{Collection: db.Addressable, Id: addr.Id}
	return
}

func (m MongoClient) UpdateAddressable(a contract.Addressable) error {
	s := m.session.Copy()
	defer s.Close()

	mapped := &models.Addressable{}
	mapped.FromContract(a)
	mapped.Modified = db.MakeTimestamp()

	c := s.DB(m.database.Name).C(db.Addressable)

	var err error
	if mapped.Id.Valid() {
		err = c.UpdateId(mapped.Id, mapped)
	} else {
		query := bson.M{"uuid": mapped.Uuid}
		err = c.Update(query, mapped)
	}

	return err
}

func (m MongoClient) GetAddressables() ([]contract.Addressable, error) {
	list, err := m.getAddressablesQuery(bson.M{})
	return mapAddressables(list, err)
}

func (m MongoClient) getAddressablesQuery(q bson.M) ([]models.Addressable, error) {
	s := m.session.Copy()
	defer s.Close()

	items := []models.Addressable{}
	col := s.DB(m.database.Name).C(db.Addressable)
	err := col.Find(q).Sort("queryts").All(&items)
	if err != nil {
		return []models.Addressable{}, err
	}

	return items, nil
}

func (m MongoClient) GetAddressableById(id string) (contract.Addressable, error) {
	addr, err := m.getAddressableById(id)
	if err != nil {
		return contract.Addressable{}, err
	}
	return addr.ToContract(), nil
}

func (m MongoClient) getAddressableById(id string) (models.Addressable, error) {
	var query bson.M
	if !bson.IsObjectIdHex(id) {
		// AddressableID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return models.Addressable{}, db.ErrInvalidObjectId
		}
		query = bson.M{"uuid": id}
	} else {
		query = bson.M{"_id": bson.ObjectIdHex(id)}
	}

	addr, err := m.getAddressable(query)
	if err != nil {
		return models.Addressable{}, err
	}
	return addr, nil
}

func (m MongoClient) AddAddressable(a contract.Addressable) (string, error) {
	s := m.session.Copy()
	defer s.Close()

	addr := &models.Addressable{}
	err := addr.FromContract(a)
	if err != nil {
		return a.Id, err
	}
	col := s.DB(m.database.Name).C(db.Addressable)

	// check if the name exist
	count, err := col.Find(bson.M{"name": addr.Name}).Count()
	if err != nil {
		return a.Id, err
	} else if count > 0 {
		return a.Id, db.ErrNotUnique
	}

	err = col.Insert(addr)
	mapped := addr.ToContract()
	return mapped.Id, err
}

func (m MongoClient) GetAddressableByName(n string) (contract.Addressable, error) {
	addr, err := m.getAddressableByName(n)
	if err != nil {
		return contract.Addressable{}, err
	}
	return addr.ToContract(), nil
}

func (m MongoClient) getAddressableByName(n string) (models.Addressable, error) {
	addr, err := m.getAddressable(bson.M{"name": n})
	if err != nil {
		return models.Addressable{}, err
	}
	return addr, nil
}

func (m MongoClient) GetAddressablesByTopic(t string) ([]contract.Addressable, error) {
	list, err := m.getAddressablesQuery(bson.M{"topic": t})
	return mapAddressables(list, err)
}

func (m MongoClient) GetAddressablesByPort(p int) ([]contract.Addressable, error) {
	list, err := m.getAddressablesQuery(bson.M{"port": p})
	return mapAddressables(list, err)
}

func (m MongoClient) GetAddressablesByPublisher(p string) ([]contract.Addressable, error) {
	list, err := m.getAddressablesQuery(bson.M{"publisher": p})
	return mapAddressables(list, err)
}

func (m MongoClient) GetAddressablesByAddress(add string) ([]contract.Addressable, error) {
	list, err := m.getAddressablesQuery(bson.M{"address": add})
	return mapAddressables(list, err)
}

func (m MongoClient) getAddressable(q bson.M) (models.Addressable, error) {
	s := m.session.Copy()
	defer s.Close()

	a := models.Addressable{}
	col := s.DB(m.database.Name).C(db.Addressable)
	err := col.Find(q).One(&a)
	return a, errorMap(err)
}

func (m MongoClient) DeleteAddressableById(id string) error {
	return m.deleteById(db.Addressable, id)
}

func mapAddressables(addrs []models.Addressable, err error) ([]contract.Addressable, error) {
	if err != nil {
		return []contract.Addressable{}, err
	}

	var mapped []contract.Addressable
	for _, a := range addrs {
		mapped = append(mapped, a.ToContract())
	}
	return mapped, nil
}

/* ----------------------------- Device Service ----------------------------------*/
func (m MongoClient) GetDeviceServiceByName(n string) (contract.DeviceService, error) {
	return m.getDeviceService(bson.M{"name": n})
}

func (m MongoClient) GetDeviceServiceById(id string) (contract.DeviceService, error) {
	var query bson.M
	if !bson.IsObjectIdHex(id) {
		// EventID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return contract.DeviceService{}, errors.New("mgoGetDeviceServiceByName Invalid Object ID " + id)
		}
		query = bson.M{"uuid": id}
	} else {
		query = bson.M{"_id": bson.ObjectIdHex(id)}
	}
	return m.getDeviceService(query)
}

func (m MongoClient) GetAllDeviceServices() ([]contract.DeviceService, error) {
	return m.getDeviceServices(bson.M{})
}

func (m MongoClient) GetDeviceServicesByAddressableId(id string) ([]contract.DeviceService, error) {
	//Incoming addressable ID could be either BSON or JSON.
	//Figure out which one it is. If UUID, load the Mongo addressable model to obtain the BSON Id
	//because the contract won't have that.
	var query bson.M
	if bson.IsObjectIdHex(id) {
		query = bson.M{"addressable.$id": bson.ObjectIdHex(id)}
	} else {
		_, err := uuid.Parse(id)
		if err == nil {
			addr, err := m.getAddressable(bson.M{"uuid": id})
			if err != nil {
				return nil, err
			}
			query = bson.M{"addressable.$id": addr.Id}
		} else {
			return nil, errors.New("mgoGetDeviceServicesByAddressableId Invalid Object ID " + id)
		}
	}
	return m.getDeviceServices(bson.M{"addressable.$id": query})
}

func (m MongoClient) GetDeviceServicesWithLabel(l string) ([]contract.DeviceService, error) {
	var ls []string
	ls = append(ls, l)
	return m.getDeviceServices(bson.M{"labels": bson.M{"$in": ls}})
}

func (m MongoClient) getDeviceServices(q bson.M) ([]contract.DeviceService, error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceService)
	dss := []models.DeviceService{}
	err := col.Find(q).Sort("queryts").All(&dss)
	if err != nil {
		return nil, err
	}

	contractDeviceServices := []contract.DeviceService{}
	for _, deviceService := range dss {
		contractDeviceService, err := deviceService.ToContract(m)
		if err != nil {
			return []contract.DeviceService{}, err
		}
		contractDeviceServices = append(contractDeviceServices, contractDeviceService)
	}
	return contractDeviceServices, nil
}

func (m MongoClient) getDeviceService(q bson.M) (contract.DeviceService, error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceService)
	ds := models.DeviceService{}
	err := col.Find(q).One(&ds)
	if err != nil {
		return contract.DeviceService{}, errorMap(err)
	}
	return ds.ToContract(m)
}

func (m MongoClient) AddDeviceService(ds contract.DeviceService) (string, error) {
	s := m.session.Copy()
	defer s.Close()

	var deviceService models.DeviceService
	if err := deviceService.FromContract(ds, m); err != nil {
		return "", errors.New("FromContract failed")
	}
	ts := db.MakeTimestamp()
	deviceService.Created = ts
	deviceService.Modified = ts

	col := s.DB(m.database.Name).C(db.DeviceService)
	err := col.Insert(deviceService)
	return deviceService.Uuid, err
}

func (m MongoClient) UpdateDeviceService(ds contract.DeviceService) error {
	s := m.session.Copy()
	defer s.Close()

	var deviceService models.DeviceService
	if err := deviceService.FromContract(ds, m); err != nil {
		return errors.New("FromContract failed")
	}
	deviceService.Modified = db.MakeTimestamp()

	col := s.DB(m.database.Name).C(db.DeviceService)
	if deviceService.Id.Valid() {
		return col.UpdateId(deviceService.Id, deviceService)
	}
	return col.Update(bson.M{"uuid": deviceService.Uuid}, deviceService)
}

func (m MongoClient) DeleteDeviceServiceById(id string) error {
	return m.deleteById(db.DeviceService, id)
}

//  ----------------------Provision Watcher -----------------------------*/
func (m MongoClient) GetAllProvisionWatchers(pw *[]contract.ProvisionWatcher) error {
	return m.GetProvisionWatchers(pw, bson.M{})
}

func (m MongoClient) GetProvisionWatcherByName(pw *contract.ProvisionWatcher, n string) error {
	return m.GetProvisionWatcher(pw, bson.M{"name": n})
}

func (m MongoClient) GetProvisionWatchersByIdentifier(pw *[]contract.ProvisionWatcher, k string, v string) error {
	return m.GetProvisionWatchers(pw, bson.M{"identifiers." + k: v})
}

func (m MongoClient) GetProvisionWatchersByServiceId(pw *[]contract.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetProvisionWatchers(pw, bson.M{"service.$id": bson.ObjectIdHex(id)})
	} else {
		return errors.New("mgoGetProvisionWatchersByServiceId Invalid Object ID " + id)
	}
}

func (m MongoClient) GetProvisionWatchersByProfileId(pw *[]contract.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetProvisionWatchers(pw, bson.M{"profile.$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetProvisionWatcherByProfileId Invalid Object ID " + id)
		return err
	}
}

func (m MongoClient) GetProvisionWatcherById(pw *contract.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return m.GetProvisionWatcher(pw, bson.M{"_id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetProvisionWatcherById Invalid Object ID " + id)
		return err
	}
}

func (m MongoClient) GetProvisionWatcher(pw *contract.ProvisionWatcher, q bson.M) error {
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

func (m MongoClient) GetProvisionWatchers(pw *[]contract.ProvisionWatcher, q bson.M) error {
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

func (m MongoClient) AddProvisionWatcher(pw *contract.ProvisionWatcher) error {
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
	if pw.Service.Id != "" {
		dev, err = m.GetDeviceServiceById(pw.Service.Id)
	} else if pw.Service.Name != "" {
		dev, err = m.GetDeviceServiceByName(pw.Service.Name)
	} else {
		return errors.New("Device Service ID or Name is required")
	}
	if err != nil {
		return err
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

func (m MongoClient) UpdateProvisionWatcher(pw contract.ProvisionWatcher) error {
	s := m.session.Copy()
	defer s.Close()
	c := s.DB(m.database.Name).C(db.ProvisionWatcher)

	pw.Modified = db.MakeTimestamp()

	// Handle DBRefs
	mpw := mongoProvisionWatcher{ProvisionWatcher: pw}

	return c.UpdateId(mpw.Id, mpw)
}

func (m MongoClient) DeleteProvisionWatcherById(id string) error {
	return m.deleteById(db.ProvisionWatcher, id)
}

//  ------------------------Command -------------------------------------*/
func (m MongoClient) GetAllCommands() ([]contract.Command, error) {
	s := m.session.Copy()
	defer s.Close()

	col := s.DB(m.database.Name).C(db.Command)
	commands := &[]models.Command{}
	err := col.Find(bson.M{}).Sort("queryts").All(commands)

	return mapCommands(*commands, err)
}

func (m MongoClient) GetCommandById(id string) (contract.Command, error) {
	s := m.session.Copy()
	defer s.Close()

	col := s.DB(m.database.Name).C(db.Command)

	var query bson.M
	if !bson.IsObjectIdHex(id) {
		// EventID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return contract.Command{}, db.ErrNotFound
		}
		query = bson.M{"uuid": id}
	} else {
		query = bson.M{"_id": bson.ObjectIdHex(id)}
	}

	command := &models.Command{}
	err := col.Find(query).One(command)

	return command.ToContract(), err
}

func (m MongoClient) GetCommandByName(n string) ([]contract.Command, error) {
	s := m.session.Copy()
	defer s.Close()

	col := s.DB(m.database.Name).C(db.Command)
	commands := &[]models.Command{}
	err := col.Find(bson.M{"name": n}).All(commands)

	return mapCommands(*commands, err)
}

func mapCommands(commands []models.Command, err error) ([]contract.Command, error) {
	if err != nil {
		return nil, errorMap(err)
	}

	var mapped []contract.Command
	for _, cmd := range commands {
		mapped = append(mapped, cmd.ToContract())
	}

	return mapped, nil
}

func (m MongoClient) AddCommand(c contract.Command) (string, error) {
	s := m.session.Copy()
	defer s.Close()

	command := &models.Command{}
	if err := command.FromContract(c); err != nil {
		return "", errors.New("FromContract failed")
	}

	err := s.DB(m.database.Name).C(db.Command).Insert(command)
	return command.Uuid, err
}

// Update command uses the ID of the command for identification
func (m MongoClient) UpdateCommand(c *contract.Command) error {
	s := m.session.Copy()
	defer s.Close()

	if c == nil {
		return nil
	}

	model := &models.Command{}
	if err := model.FromContract(*c); err != nil {
		return errors.New("FromContract failed")
	}

	var err error
	if model.Id.Valid() {
		err = s.DB(m.database.Name).C(db.Command).UpdateId(model.Id, model)
	} else {
		err = s.DB(m.database.Name).C(db.Command).Update(bson.M{"uuid": model.Uuid}, model)
	}
	return errorMap(err)
}

// Delete the command by ID
// Check if the command is still in use by device profiles
func (m MongoClient) DeleteCommandById(id string) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Command)

	var findParameters bson.M
	var deleteParameters bson.D
	if !bson.IsObjectIdHex(id) {
		// EventID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return db.ErrInvalidObjectId
		}
		findParameters = bson.M{"uuid": id}
		deleteParameters = bson.D{{Name: "uuid", Value: id}}
	} else {
		var objectId = bson.ObjectIdHex(id)
		findParameters = bson.M{"_id": objectId}
		deleteParameters = bson.D{{Name: "_id", Value: objectId}}
	}

	// Check if the command is still in use
	query := bson.M{"commands": bson.M{"$elemMatch": findParameters}}
	count, err := s.DB(m.database.Name).C(db.DeviceProfile).Find(query).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return db.ErrCommandStillInUse
	}

	return col.Remove(deleteParameters)
}

func (m MongoClient) GetAndMapCommands(c []mgo.DBRef) ([]contract.Command, error) {
	s := m.session.Copy()
	defer s.Close()

	var commands []contract.Command
	for _, cRef := range c {
		var command contract.Command
		command, err := m.GetCommandById(fmt.Sprintf("%s", cRef.Id))
		if err != nil {
			return []contract.Command{}, errorMap(err)
		}
		commands = append(commands, command)
	}
	return commands, nil
}

// Scrub all metadata
func (m MongoClient) ScrubMetadata() error {
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
