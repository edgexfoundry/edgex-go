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
	se.Modified = db.MakeTimestamp()

	// Handle DBRefs
	mse := mongoScheduleEvent{ScheduleEvent: se}

	return m.updateId(db.ScheduleEvent, se.Id, mse)
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
	sch.Modified = db.MakeTimestamp()

	if err := m.updateId(db.Schedule, sch.Id, sch); err != nil {
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
	query, err := idToBsonM(id)
	if err != nil {
		return contract.DeviceReport{}, err
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

	var mapped models.DeviceReport
	if err := mapped.FromContract(dr); err != nil {
		return errors.New("FromContract failed")
	}

	query, err := idToBsonM(dr.Id)
	if err != nil {
		return err
	}
	return errorMap(s.DB(m.database.Name).C(db.Addressable).Update(query, mapped))
}

func (m MongoClient) DeleteDeviceReportById(id string) error {
	return m.deleteById(db.DeviceReport, id)
}

/* ----------------------------- Device ---------------------------------- */
func (m MongoClient) AddDevice(d contract.Device) (string, error) {
	s := m.session.Copy()
	defer s.Close()

	col := s.DB(m.database.Name).C(db.Device)

	// Check if the name exist (Device names must be unique)
	count, err := col.Find(bson.M{"name": d.Name}).Count()
	if err != nil {
		return "", err
	}
	if count > 0 {
		return "", db.ErrNotUnique
	}

	var device models.Device
	if err := device.FromContract(d, m, m, m, m); err != nil {
		return "", errors.New("FromContract failed")
	}

	ts := db.MakeTimestamp()
	device.Created = ts
	device.Modified = ts

	return device.Uuid, col.Insert(device)
}

func (m MongoClient) UpdateDevice(d contract.Device) error {
	s := m.session.Copy()
	defer s.Close()

	var err error
	var mapped models.Device
	if err = mapped.FromContract(d, m, m, m, m); err != nil {
		return errors.New("FromContract failed")
	}

	var query bson.M
	if query, err = idToBsonM(d.Id); err != nil {
		return err
	}
	return errorMap(s.DB(m.database.Name).C(db.Device).Update(query, mapped))
}

func (m MongoClient) DeleteDeviceById(id string) error {
	return m.deleteById(db.Device, id)
}

func (m MongoClient) GetAllDevices() ([]contract.Device, error) {
	return m.getDevices(nil)
}

func (m MongoClient) GetDevicesByProfileId(id string) ([]contract.Device, error) {
	//Incoming profile ID could be either BSON or JSON.
	//If UUID, load the Mongo model to obtain the BSON Id because the contract won't have that.
	if bson.IsObjectIdHex(id) {
		return m.getDevices(bson.M{"profile.$id": bson.ObjectIdHex(id)})
	}

	if _, err := uuid.Parse(id); err == nil {
		model, err := m.getDeviceProfileById(id)
		if err != nil {
			return []contract.Device{}, err
		}
		return m.getDevices(bson.M{"profile.$id": model.Id})
	}

	return []contract.Device{}, errors.New("mgoGetDevicesByProfileId Invalid ID " + id)
}

func (m MongoClient) GetDeviceById(id string) (contract.Device, error) {
	query, err := idToBsonM(id)
	if err != nil {
		return contract.Device{}, err
	}
	return m.getDevice(query)
}

func (m MongoClient) GetDeviceByName(n string) (contract.Device, error) {
	return m.getDevice(bson.M{"name": n})
}

func (m MongoClient) GetDevicesByServiceId(id string) ([]contract.Device, error) {
	//Incoming device service ID could be either BSON or JSON.
	//If UUID, load the Mongo model to obtain the BSON Id because the contract won't have that.
	if bson.IsObjectIdHex(id) {
		return m.getDevices(bson.M{"profile.$id": bson.ObjectIdHex(id)})
	}

	if _, err := uuid.Parse(id); err == nil {
		model, err := m.getDeviceServiceById(id)
		if err != nil {
			return []contract.Device{}, err
		}
		return m.getDevices(bson.M{"profile.$id": model.Id})
	}

	return []contract.Device{}, errors.New("mgoGetDevicesByServiceId Invalid ID " + id)
}

func (m MongoClient) GetDevicesByAddressableId(id string) ([]contract.Device, error) {
	//Incoming addressable ID could be either BSON or JSON.
	//If UUID, load the Mongo model to obtain the BSON Id because the contract won't have that.
	if bson.IsObjectIdHex(id) {
		return m.getDevices(bson.M{"addressable.$id": bson.ObjectIdHex(id)})
	}

	if _, err := uuid.Parse(id); err == nil {
		model, err := m.getAddressable(bson.M{"uuid": id})
		if err != nil {
			return []contract.Device{}, err
		}
		return m.getDevices(bson.M{"addressable.$id": model.Id})
	}

	return []contract.Device{}, errors.New("mgoGetDevicesByAddressableId Invalid ID " + id)
}

func (m MongoClient) GetDevicesWithLabel(l string) ([]contract.Device, error) {
	return m.getDevices(bson.M{"labels": bson.M{"$in": []string{l}}})
}

func (m MongoClient) getDevices(q bson.M) ([]contract.Device, error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Device)
	mds := []models.Device{}

	err := col.Find(q).Sort("queryts").All(&mds)
	if err != nil {
		return []contract.Device{}, err
	}

	res := []contract.Device{}
	for _, md := range mds {
		d, err := md.ToContract(m, m, m, m)
		if err != nil {
			return []contract.Device{}, err
		}
		res = append(res, d)
	}

	return res, nil
}

func (m MongoClient) getDevice(q bson.M) (contract.Device, error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Device)
	var d models.Device

	if err := col.Find(q).One(&d); err != nil {
		return contract.Device{}, errorMap(err)
	}
	return d.ToContract(m, m, m, m)
}

/* -----------------------------Device Profile -----------------------------*/

// DBRefToDeviceProfile converts DBRef to internal Mongo struct
func (m MongoClient) DBRefToDeviceProfile(dbRef mgo.DBRef) (a models.DeviceProfile, err error) {
	s := m.session.Copy()
	defer s.Close()

	if err = s.DB(m.database.Name).C(db.DeviceProfile).Find(bson.M{"_id": dbRef.Id}).One(&a); err != nil {
		return models.DeviceProfile{}, err
	}
	return
}

// DeviceProfileToDBRef converts internal Mongo struct to DBRef
func (m MongoClient) DeviceProfileToDBRef(model models.DeviceProfile) (dbRef mgo.DBRef, err error) {
	s := m.session.Copy()
	defer s.Close()

	// validate model with identity provided in contract actually exists
	model, err = m.getDeviceProfileById(model.Id.Hex())
	if err != nil {
		return
	}

	dbRef = mgo.DBRef{Collection: db.DeviceProfile, Id: model.Id}
	return
}

func (m MongoClient) GetDeviceProfileById(id string) (contract.DeviceProfile, error) {
	model, err := m.getDeviceProfileById(id)
	if err != nil {
		return contract.DeviceProfile{}, err
	}
	return model.ToContract(m)
}

func (m MongoClient) getDeviceProfileById(id string) (dp models.DeviceProfile, err error) {
	var query bson.M
	if query, err = idToBsonM(id); err != nil {
		return models.DeviceProfile{}, err
	}
	return m.getDeviceProfile(query)
}

func (m MongoClient) GetAllDeviceProfiles() ([]contract.DeviceProfile, error) {
	return m.getDeviceProfiles(nil)
}

func (m MongoClient) GetDeviceProfilesByModel(model string) ([]contract.DeviceProfile, error) {
	return m.getDeviceProfiles(bson.M{"model": model})
}

func (m MongoClient) GetDeviceProfilesWithLabel(l string) ([]contract.DeviceProfile, error) {
	var ls []string
	ls = append(ls, l)

	return m.getDeviceProfiles(bson.M{"labels": bson.M{"$in": ls}})
}

func (m MongoClient) GetDeviceProfilesByManufacturerModel(man string, mod string) ([]contract.DeviceProfile, error) {
	return m.getDeviceProfiles(bson.M{"manufacturer": man, "model": mod})
}

func (m MongoClient) GetDeviceProfilesByManufacturer(man string) ([]contract.DeviceProfile, error) {
	return m.getDeviceProfiles(bson.M{"manufacturer": man})
}

func (m MongoClient) GetDeviceProfileByName(n string) (contract.DeviceProfile, error) {
	model, err := m.getDeviceProfile(bson.M{"name": n})
	if err != nil {
		return contract.DeviceProfile{}, err
	}
	return model.ToContract(m)
}

// Get device profiles with the passed query
func (m MongoClient) getDeviceProfiles(q bson.M) (cdps []contract.DeviceProfile, err error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceProfile)

	// Handle the DBRefs
	var dps []models.DeviceProfile
	err = col.Find(q).Sort("queryts").All(&dps)
	if err != nil {
		return []contract.DeviceProfile{}, err
	}

	cdps = make([]contract.DeviceProfile, 0)

	for _, dp := range dps {
		c, err := dp.ToContract(m)
		if err != nil {
			return []contract.DeviceProfile{}, err
		}
		cdps = append(cdps, c)
	}
	return
}

func (m MongoClient) getDeviceProfile(q bson.M) (d models.DeviceProfile, err error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceProfile)

	err = col.Find(q).One(&d)
	if err != nil {
		return models.DeviceProfile{}, errorMap(err)
	}

	return
}

func (m MongoClient) AddDeviceProfile(dp contract.DeviceProfile) (string, error) {
	s := m.session.Copy()
	defer s.Close()
	if len(dp.Name) == 0 {
		return "", db.ErrNameEmpty
	}
	col := s.DB(m.database.Name).C(db.DeviceProfile)
	count, err := col.Find(bson.M{"name": dp.Name}).Count()
	if err != nil {
		return "", err
	} else if count > 0 {
		return "", db.ErrNotUnique
	}
	for i := 0; i < len(dp.Commands); i++ {
		if newId, err := m.AddCommand(dp.Commands[i]); err != nil {
			return "", err
		} else {
			dp.Commands[i].Id = newId
		}
	}
	ts := db.MakeTimestamp()
	dp.Created = ts
	dp.Modified = ts

	var mdp models.DeviceProfile
	var id string
	id, err = mdp.FromContract(dp, m)
	if err != nil {
		return "", err
	}

	if err = col.Insert(mdp); err != nil {
		return "", err
	}

	return id, nil
}

func (m MongoClient) UpdateDeviceProfile(dp contract.DeviceProfile) error {
	dp.Modified = db.MakeTimestamp()

	mdp := models.DeviceProfile{}
	if _, err := mdp.FromContract(dp, m); err != nil {
		return err
	}

	return m.updateId(db.DeviceProfile, mdp.Id, mdp)
}

// Get the device profiles that are currently using the command
func (m MongoClient) GetDeviceProfilesUsingCommand(c contract.Command) ([]contract.DeviceProfile, error) {
	var item bson.M
	if !bson.IsObjectIdHex(c.Id) {
		// EventID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(c.Id)
		if err != nil { // It is some unsupported type of string
			return []contract.DeviceProfile{}, db.ErrInvalidObjectId
		}
		item = bson.M{"uuid": c.Id}
	} else {
		item = bson.M{"$id": bson.ObjectIdHex(c.Id)}
	}

	query := bson.M{"commands": bson.M{"$elemMatch": item}}
	var dps []contract.DeviceProfile
	var err error
	if dps, err = m.getDeviceProfiles(query); err != nil {
		return []contract.DeviceProfile{}, err
	}
	return dps, nil
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

	var err error
	mapped := &models.Addressable{}
	if err = mapped.FromContract(a); err != nil {
		return err
	}

	mapped.Modified = db.MakeTimestamp()

	var query bson.M
	if query, err = idToBsonM(a.Id); err != nil {
		return err
	}
	return s.DB(m.database.Name).C(db.Addressable).Update(query, mapped)
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
	query, err := idToBsonM(id)
	if err != nil {
		return models.Addressable{}, err
	}
	return m.getAddressable(query)
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
func (m MongoClient) DBRefToDeviceService(dbRef mgo.DBRef) (ds models.DeviceService, err error) {
	s := m.session.Copy()
	defer s.Close()

	if err = s.DB(m.database.Name).C(db.DeviceService).Find(bson.M{"_id": dbRef.Id}).One(&ds); err != nil {
		return models.DeviceService{}, err
	}
	return
}

// DeviceServiceToDBRef converts internal Mongo struct to DBRef
func (m MongoClient) DeviceServiceToDBRef(model models.DeviceService) (dbRef mgo.DBRef, err error) {
	s := m.session.Copy()
	defer s.Close()

	// validate model with identity provided in contract actually exists
	model, err = m.getDeviceServiceById(model.Id.Hex())
	if err != nil {
		return
	}

	dbRef = mgo.DBRef{Collection: db.DeviceService, Id: model.Id}
	return
}

func (m MongoClient) GetDeviceServiceByName(n string) (contract.DeviceService, error) {
	ds, err := m.deviceService(bson.M{"name": n})
	if err != nil {
		return contract.DeviceService{}, err
	}
	return ds.ToContract(m)
}

func (m MongoClient) GetDeviceServiceById(id string) (contract.DeviceService, error) {
	ds, err := m.getDeviceServiceById(id)
	if err != nil {
		return contract.DeviceService{}, err
	}
	return ds.ToContract(m)
}

func (m MongoClient) getDeviceServiceById(id string) (models.DeviceService, error) {
	query, err := idToBsonM(id)
	if err != nil {
		return models.DeviceService{}, err
	}
	return m.deviceService(query)
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

func (m MongoClient) getDeviceServices(q bson.M) (dss []contract.DeviceService, err error) {
	dss = []contract.DeviceService{}
	mds, err := m.deviceServices(q)
	if err != nil {
		return
	}
	for _, ds := range mds {
		cds, err := ds.ToContract(m)
		if err != nil {
			return []contract.DeviceService{}, err
		}
		dss = append(dss, cds)
	}

	return dss, nil
}

func (m MongoClient) deviceServices(q bson.M) (dss []models.DeviceService, err error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.DeviceService)
	err = col.Find(q).Sort("queryts").All(&dss)
	if err != nil {
		return []models.DeviceService{}, err
	}

	return
}

func (m MongoClient) getDeviceService(q bson.M) (ds contract.DeviceService, err error) {
	mds, err := m.deviceService(q)
	if err != nil {
		return contract.DeviceService{}, err
	}
	ds, err = mds.ToContract(m)
	return
}

func (m MongoClient) deviceService(q bson.M) (models.DeviceService, error) {
	s := m.session.Copy()
	defer s.Close()

	var ds models.DeviceService
	err := s.DB(m.database.Name).C(db.DeviceService).Find(q).One(&ds)
	if err != nil {
		return models.DeviceService{}, errorMap(err)
	}

	return ds, nil
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

	var err error
	var mapped models.DeviceService
	if err = mapped.FromContract(ds, m); err != nil {
		return errors.New("FromContract failed")
	}
	mapped.Modified = db.MakeTimestamp()

	var query bson.M
	if query, err = idToBsonM(ds.Id); err != nil {
		return err
	}
	return s.DB(m.database.Name).C(db.DeviceService).Update(query, mapped)
}

func (m MongoClient) DeleteDeviceServiceById(id string) error {
	return m.deleteById(db.DeviceService, id)
}

//  ----------------------Provision Watcher -----------------------------*/
func (m MongoClient) GetAllProvisionWatchers() (pw []contract.ProvisionWatcher, err error) {
	return m.getProvisionWatchers(bson.M{})
}

func (m MongoClient) GetProvisionWatcherByName(n string) (pw contract.ProvisionWatcher, err error) {
	return m.GetProvisionWatcher(bson.M{"name": n})
}

func (m MongoClient) GetProvisionWatchersByIdentifier(k string, v string) (pw []contract.ProvisionWatcher, err error) {
	return m.getProvisionWatchers(bson.M{"identifiers." + k: v})
}

func (m MongoClient) GetProvisionWatchersByServiceId(id string) (pw []contract.ProvisionWatcher, err error) {
	var query bson.M
	if !bson.IsObjectIdHex(id) {
		// ID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return []contract.ProvisionWatcher{}, errors.New("mgoGetProvisionWatchersByServiceId Invalid Object ID " + id)
		}
		query = bson.M{"service.uuid": id}
	} else {
		query = bson.M{"service.$id": bson.ObjectIdHex(id)}
	}

	pw, err = m.getProvisionWatchers(query)
	if err != nil {
		return []contract.ProvisionWatcher{}, err
	}

	return pw, nil
}

func (m MongoClient) GetProvisionWatchersByProfileId(id string) (pw []contract.ProvisionWatcher, err error) {
	var query bson.M
	if !bson.IsObjectIdHex(id) {
		// ID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return []contract.ProvisionWatcher{}, errors.New("mgoGetProvisionWatchersByProfileId Invalid Object ID " + id)
		}
		query = bson.M{"profile.uuid": id}
	} else {
		query = bson.M{"profile.$id": bson.ObjectIdHex(id)}
	}

	pw, err = m.getProvisionWatchers(query)
	if err != nil {
		return []contract.ProvisionWatcher{}, err
	}

	return pw, nil
}

func (m MongoClient) GetProvisionWatcherById(id string) (pw contract.ProvisionWatcher, err error) {
	var query bson.M
	if !bson.IsObjectIdHex(id) {
		// ID is not a BSON ID. Is it a UUID?
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return contract.ProvisionWatcher{}, errors.New("mgoGetProvisionWatcherById Invalid Object ID " + id)
		}
		query = bson.M{"uuid": id}
	} else {
		query = bson.M{"_id": bson.ObjectIdHex(id)}
	}

	pw, err = m.GetProvisionWatcher(query)
	if err != nil {
		return contract.ProvisionWatcher{}, err
	}

	return
}

func (m MongoClient) GetProvisionWatcher(q bson.M) (pw contract.ProvisionWatcher, err error) {
	mpw, err := m.getProvisionWatcher(q)
	if err != nil {
		return
	}

	pw, err = mpw.ToContract(m, m, m, m)

	return
}

func (m MongoClient) getProvisionWatcher(q bson.M) (mpw models.ProvisionWatcher, err error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.ProvisionWatcher)

	err = col.Find(q).One(&mpw)
	if err != nil {
		return models.ProvisionWatcher{}, errorMap(err)
	}

	return mpw, nil
}

func (m MongoClient) getProvisionWatchers(q bson.M) (pws []contract.ProvisionWatcher, err error) {
	mpws, err := m.provisionWatchers(q)

	var cpws []contract.ProvisionWatcher
	for _, mpw := range mpws {
		cpw, err := mpw.ToContract(m, m, m, m)
		if err != nil {
			return []contract.ProvisionWatcher{}, err
		}
		cpws = append(cpws, cpw)
	}

	return cpws, nil
}

func (m MongoClient) provisionWatchers(q bson.M) (pws []models.ProvisionWatcher, err error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.ProvisionWatcher)

	var mpws []models.ProvisionWatcher

	err = col.Find(q).Sort("queryts").All(&mpws)
	if err != nil {
		return []models.ProvisionWatcher{}, err
	}

	return mpws, nil
}

func (m MongoClient) AddProvisionWatcher(pw contract.ProvisionWatcher) (string, error) {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.ProvisionWatcher)
	count, err := col.Find(bson.M{"name": pw.Name}).Count()
	if err != nil {
		return "", err
	} else if count > 0 {
		return "", db.ErrNotUnique
	}

	// get Device Service
	var dev contract.DeviceService
	if pw.Service.Id != "" {
		dev, err = m.GetDeviceServiceById(pw.Service.Id)
	} else if pw.Service.Name != "" {
		dev, err = m.GetDeviceServiceByName(pw.Service.Name)
	} else {
		return "", errors.New("Device Service ID or Name is required")
	}
	if err != nil {
		return "", err
	}
	pw.Service = dev

	// get Device Profile
	var dp contract.DeviceProfile
	if pw.Profile.Id != "" {
		dp, err = m.GetDeviceProfileById(pw.Profile.Id)
	} else if pw.Profile.Name != "" {
		dp, err = m.GetDeviceProfileByName(pw.Profile.Name)
	} else {
		return "", errors.New("Device Profile ID or Name is required")
	}
	if err != nil {
		return "", err
	}

	pw.Profile = dp

	var mpw models.ProvisionWatcher
	id, err := mpw.FromContract(pw, m, m, m, m)
	if err != nil {
		return "", errors.New("ProvisionWatcher FromContract() failed")
	}

	// Set timestamps
	ts := db.MakeTimestamp()
	mpw.Created = ts
	mpw.Modified = ts

	err = col.Insert(mpw)

	return id, err
}

func (m MongoClient) UpdateProvisionWatcher(pw contract.ProvisionWatcher) error {
	s := m.session.Copy()
	defer s.Close()

	pw.Modified = db.MakeTimestamp()

	var mpw models.ProvisionWatcher
	id, err := mpw.FromContract(pw, m, m, m, m)
	if err != nil {
		return err
	}

	return m.updateId(db.ProvisionWatcher, id, mpw)
}

func (m MongoClient) DeleteProvisionWatcherById(id string) error {
	return m.deleteById(db.ProvisionWatcher, id)
}

//  ------------------------Command -------------------------------------*/
func (m MongoClient) DBRefToCommand(dbRef mgo.DBRef) (c models.Command, err error) {
	s := m.session.Copy()
	defer s.Close()

	if err = s.DB(m.database.Name).C(db.Command).Find(bson.M{"_id": dbRef.Id}).One(&c); err != nil {
		return models.Command{}, err
	}
	return
}

func (m MongoClient) CommandToDBRef(c models.Command) (dbRef mgo.DBRef, err error) {
	s := m.session.Copy()
	defer s.Close()

	// validate command identity provided in contract actually exists and populate missing Id, Uuid field
	var command models.Command
	if c.Id.Valid() {
		command, err = m.getCommandById(c.Id.Hex())
	} else {
		command, err = m.getCommandById(c.Uuid)
	}
	if err != nil {
		return
	}

	dbRef = mgo.DBRef{Collection: db.Command, Id: command.Id}
	return
}

func (m MongoClient) GetAllCommands() ([]contract.Command, error) {
	s := m.session.Copy()
	defer s.Close()

	col := s.DB(m.database.Name).C(db.Command)
	commands := &[]models.Command{}
	err := col.Find(bson.M{}).Sort("queryts").All(commands)

	return mapCommands(*commands, err)
}

func (m MongoClient) GetCommandById(id string) (contract.Command, error) {
	command, err := m.getCommandById(id)

	if err != nil {
		return contract.Command{}, err
	}

	return command.ToContract(), nil
}

func (m MongoClient) getCommandById(id string) (models.Command, error) {
	s := m.session.Copy()
	defer s.Close()

	query, err := idToBsonM(id)
	if err != nil {
		return models.Command{}, err
	}

	var command models.Command
	if err = s.DB(m.database.Name).C(db.Command).Find(query).One(&command); err != nil {
		return models.Command{}, err
	}
	return command, nil
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

	var id string
	var err error
	command := &models.Command{}
	if id, err = command.FromContract(c); err != nil {
		return "", errors.New("FromContract failed")
	}

	err = s.DB(m.database.Name).C(db.Command).Insert(command)

	return id, err
}

// Update command uses the ID of the command for identification
func (m MongoClient) UpdateCommand(c contract.Command) error {
	s := m.session.Copy()
	defer s.Close()

	var err error
	var mapped models.Command
	if _, err = mapped.FromContract(c); err != nil {
		return errors.New("FromContract failed")
	}

	var query bson.M
	if query, err = idToBsonM(c.Id); err != nil {
		return err
	}
	return errorMap(s.DB(m.database.Name).C(db.Command).Update(query, mapped))
}

// Delete the command by ID
// Check if the command is still in use by device profiles
func (m MongoClient) DeleteCommandById(id string) error {
	s := m.session.Copy()
	defer s.Close()
	col := s.DB(m.database.Name).C(db.Command)

	// Check if the command is still in use
	findParameters, err := idToBsonM(id)
	if err != nil {
		return err
	}
	query := bson.M{"commands": bson.M{"$elemMatch": findParameters}}
	count, err := s.DB(m.database.Name).C(db.DeviceProfile).Find(query).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return db.ErrCommandStillInUse
	}

	// remove the command
	n, v, err := idToQueryParameters(id)
	if err != nil {
		return err
	}
	return col.Remove(bson.D{{Name: n, Value: v}})
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
