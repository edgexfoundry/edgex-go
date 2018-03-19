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
 * @original author: Spencer Bull & Ryan Comer, Dell
 * @version: 0.5.0
 * @updated by:  Jim White, Dell Technologies, Feb 27, 2108
 * Added func makeTimestamp and import of time to support it (Fede C. initiated during mono repo work)
 *******************************************************************************/
package metadata

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

/* -----------------------Schedule Event ------------------------*/
func mgoUpdateScheduleEvent(se models.ScheduleEvent) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(SECOL)

	se.Modified = makeTimestamp()

	// Handle DBRefs
	mse := MongoScheduleEvent{ScheduleEvent: se}

	return col.UpdateId(se.Id, mse)
}
func mgoAddScheduleEvent(se *models.ScheduleEvent) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(SECOL)
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

	if err := col.Insert(mse); err != nil {
		return err
	}
	return nil
}
func mgoGetAllScheduleEvents(se *[]models.ScheduleEvent) error {
	return mgoGetScheduleEvents(se, bson.M{})
}
func mgoGetScheduleEventByName(se *models.ScheduleEvent, n string) error {
	return mgoGetScheduleEvent(se, bson.M{NAME: n})
}
func mgoGetScheduleEventById(se *models.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return mgoGetScheduleEvent(se, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleEventById Invalid Object ID " + id)
		return err
	}
}
func mgoGetScheduleEventsByScheduleName(se *[]models.ScheduleEvent, n string) error {
	return mgoGetScheduleEvents(se, bson.M{SCHEDULE: n})
}
func mgoGetScheduleEventsByAddressableId(se *[]models.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return mgoGetScheduleEvents(se, bson.M{ADDRESSABLE + ".$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleEventsByAddressableId Invalid Object ID" + id)
		return err
	}
}
func mgoGetScheduleEventsByServiceName(se *[]models.ScheduleEvent, n string) error {
	return mgoGetScheduleEvents(se, bson.M{SERVICE: n})
}
func mgoGetScheduleEvent(se *models.ScheduleEvent, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(SECOL)

	// Handle DBRef
	var mse MongoScheduleEvent

	err := col.Find(q).One(&mse)
	if err != nil {
		return err
	}

	*se = mse.ScheduleEvent

	return err
}
func mgoGetScheduleEvents(se *[]models.ScheduleEvent, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(SECOL)

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

/* --------------------------Schedule ---------------------------*/
func mgoGetAllSchedules(s *[]models.Schedule) error {
	return mgoGetSchedules(s, bson.M{})
}
func mgoGetScheduleByName(s *models.Schedule, n string) error {
	return mgoGetSchedule(s, bson.M{NAME: n})
}
func mgoGetScheduleById(s *models.Schedule, id string) error {
	if bson.IsObjectIdHex(id) {
		return mgoGetSchedule(s, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetScheduleById Invalid Object ID " + id)
		return err
	}
}
func mgoAddSchedule(s *models.Schedule) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(SCOL)
	count, err := col.Find(bson.M{NAME: s.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		err := errors.New("Schedule already exist")
		return err
	}

	ts := makeTimestamp()
	s.Created = ts
	s.Modified = ts
	s.Id = bson.NewObjectId()
	if err := col.Insert(s); err != nil {
		return err
	}
	return nil
}
func mgoUpdateSchedule(s models.Schedule) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(SCOL)

	s.Modified = makeTimestamp()

	if err := col.UpdateId(s.Id, s); err != nil {
		return err
	}

	return nil
}
func mgoGetSchedule(s *models.Schedule, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(SCOL)
	err := col.Find(q).One(s)
	if err != nil {
		return err
	}

	return nil
}
func mgoGetSchedules(s *[]models.Schedule, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(SCOL)
	err := col.Find(q).Sort(QUERYTS).All(s)
	if err != nil {
		return err
	}

	return nil
}

/* ----------------------Device Report --------------------------*/
func mgoGetAllDeviceReports(d *[]models.DeviceReport) error {
	return mgoGetDeviceReports(d, bson.M{})
}
func mgoGetDeviceReportByName(d *models.DeviceReport, n string) error {
	return mgoGetDeviceReport(d, bson.M{NAME: n})
}
func mgoGetDeviceReportByDeviceName(d *[]models.DeviceReport, n string) error {
	return mgoGetDeviceReports(d, bson.M{DEVICE: n})
}
func mgoGetDeviceReportById(d *models.DeviceReport, id string) error {
	if bson.IsObjectIdHex(id) {
		return mgoGetDeviceReport(d, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceReportById Invalid Object ID " + id)
		return err
	}
}
func mgoGetDeviceReportsByScheduleEventName(d *[]models.DeviceReport, n string) error {
	return mgoGetDeviceReports(d, bson.M{"event": n})
}
func mgoGetDeviceReports(d *[]models.DeviceReport, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DRCOL)
	err := col.Find(q).Sort(QUERYTS).All(d)
	if err != nil {
		return err
	}

	return nil
}
func mgoGetDeviceReport(d *models.DeviceReport, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DRCOL)
	err := col.Find(q).One(d)
	if err != nil {
		return err
	}

	return nil
}
func mgoAddDeviceReport(d *models.DeviceReport) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DRCOL)
	count, err := col.Find(bson.M{NAME: d.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return ErrDuplicateName
	}
	ts := makeTimestamp()
	d.Created = ts
	d.Id = bson.NewObjectId()
	if err := col.Insert(d); err != nil {
		return err
	}

	return nil
}
func mgoUpdateDeviceReport(dr *models.DeviceReport) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DRCOL)

	return col.UpdateId(dr.Id, dr)
}
func mgoUpdateByIdInt(c string, did string, pv2 string, p2 int64) error {
	if bson.IsObjectIdHex(did) {
		ds := DS.dataStore()
		defer ds.s.Close()
		col := ds.s.DB(DB).C(c)
		err := col.UpdateId(bson.ObjectIdHex(did), bson.M{"$set": bson.M{pv2: p2, "modified": makeTimestamp()}})
		if err != nil {
			return err
		}

		return nil
	} else {
		err := errors.New("mgoUpdateByIdInt Invalid Object ID " + did)
		return err
	}
}
func mgoUpdateById(c string, did string, pv2 string, p2 string) error {
	if bson.IsObjectIdHex(did) {
		ds := DS.dataStore()
		defer ds.s.Close()
		col := ds.s.DB(DB).C(c)
		err := col.UpdateId(bson.ObjectIdHex(did), bson.M{"$set": bson.M{pv2: p2, "modified": makeTimestamp()}})
		if err != nil {
			return err
		}

		return nil
	} else {
		err := errors.New("mgoUpdateById Invalid Object ID " + did)
		return err
	}
}
func mgoUpdateByNameInt(c string, n string, pv2 string, p2 int64) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(c)
	err := col.Update(bson.M{"name": n}, bson.M{"$set": bson.M{pv2: p2, "modified": makeTimestamp()}})
	if err != nil {
		return err
	}

	return nil
}
func mgoUpdateByName(c string, n string, pv2 string, p2 string) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(c)
	err := col.Update(bson.M{"name": n}, bson.M{"$set": bson.M{pv2: p2, "modified": makeTimestamp()}})
	if err != nil {
		return err
	}

	return nil
}
func mgoDeleteById(c string, did string) error {
	if bson.IsObjectIdHex(did) {
		ds := DS.dataStore()
		defer ds.s.Close()
		col := ds.s.DB(DB).C(c)
		err := col.RemoveId(bson.ObjectIdHex(did))
		if err != nil {
			return err
		}

		return nil
	} else {
		err := errors.New("mgoDeleteById Invalid Object ID " + did)
		return err
	}
}
func mgoDeleteByName(c string, n string) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(c)
	err := col.Remove(bson.M{NAME: n})
	if err != nil {
		return err
	}

	return nil
}

/* ----------------------------- Device ---------------------------------- */
func mgoAddNewDevice(d *models.Device) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DEVICECOL)

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

	err = col.Insert(md)
	if err != nil {
		return err
	}

	return nil
}
func mgoUpdateDevice(rd models.Device) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	c := ds.s.DB(DB).C(DEVICECOL)

	// Copy over the DBRefs
	md := MongoDevice{Device: rd}

	return c.UpdateId(rd.Id, md)
}
func mgoGetAllDevices(d *[]models.Device) error {
	return mgoGetDevices(d, nil)
}
func mgoGetDevicesByProfileId(d *[]models.Device, pid string) error {
	if bson.IsObjectIdHex(pid) {
		return mgoGetDevices(d, bson.M{PROFILE + "." + "$id": bson.ObjectIdHex(pid)})
	} else {
		err := errors.New("mgoGetDevicesByProfileId Invalid Object ID " + pid)
		return err
	}
}
func mgoGetDeviceById(d *models.Device, id string) error {
	if bson.IsObjectIdHex(id) {
		return mgoGetDevice(d, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceById Invalid Object ID " + id)
		return err
	}
}
func mgoGetDeviceByName(d *models.Device, n string) error {
	return mgoGetDevice(d, bson.M{NAME: n})
}
func mgoGetDevicesByProfileName(d *[]models.Device, pn string) error {
	return mgoGetDevices(d, bson.M{PROFILE + "." + NAME: pn})
}
func mgoGetDevicesByServiceId(d *[]models.Device, sid string) error {
	if bson.IsObjectIdHex(sid) {
		return mgoGetDevices(d, bson.M{SERVICE + "." + "$id": bson.ObjectIdHex(sid)})
	} else {
		err := errors.New("mgoGetDevicesByServiceId Invalid Object ID " + sid)
		return err
	}
}
func mgoGetDevicesByServiceName(d *[]models.Device, sn string) error {
	return mgoGetDevices(d, bson.M{SERVICE + "." + NAME: sn})
}
func mgoGetDevicesByAddressableId(d *[]models.Device, aid string) error {
	if bson.IsObjectIdHex(aid) {
		// Check if the addressable exists
		var a *models.Addressable
		if mgoGetAddressableById(a, aid) == mgo.ErrNotFound {
			return mgo.ErrNotFound
		}
		return mgoGetDevices(d, bson.M{ADDRESSABLE + "." + "$id": bson.ObjectIdHex(aid)})
	} else {
		err := errors.New("mgoGetDevicesByAddressableId Invalid Object ID " + aid)
		return err
	}
}
func mgoGetDevicesByAddressableName(d *[]models.Device, an string) error {
	return mgoGetDevices(d, bson.M{ADDRESSABLE + "." + NAME: an})
}
func mgoGetDevicesWithLabel(d *[]models.Device, l []string) error {
	return mgoGetDevices(d, bson.M{LABELS: bson.M{"$in": l}})
}
func mgoGetDevices(d *[]models.Device, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DEVICECOL)
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
func mgoGetDevice(d *models.Device, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DEVICECOL)
	md := MongoDevice{}

	err := col.Find(q).One(&md)
	if err != nil {
		return err
	}

	*d = md.Device

	return nil
}

// Query for the aux and de-reference the DBRefs
func query(colStr string, q bson.M, aux interface{}, model interface{}) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(colStr)

	err := col.Find(q).One(aux)
	if err != nil {
		return err
	}

	// Copy over fields and de-reference the references
	refAux := reflect.ValueOf(aux).Elem()
	refModel := reflect.ValueOf(model).Elem()
	for i := 0; i < refAux.NumField(); i++ {
		// Get the fields for Aux and Real
		name := refAux.Type().Field(i).Name
		fAux := refAux.FieldByName(name)
		fReal := refModel.FieldByName(name)

		// DBRef type - dereference
		if fAux.Type() == reflect.TypeOf(mgo.DBRef{}) {
			var aux2 interface{}
			model2 := reflect.Zero(reflect.TypeOf(fReal.Interface())).Interface()

			// Make a recursive call to de-reference
			query(fAux.Interface().(mgo.DBRef).Collection, bson.M{"_id": fAux.Interface().(mgo.DBRef).Id}, &aux2, &model2)

			// Set the returned value into the Real field
			fReal.Set(reflect.ValueOf(model2))
			continue
		}

		// Not a DBRef, just copy over the field
		fReal.Set(refAux.FieldByName(name))
	}

	return nil
}

/* -----------------------------Device Profile -----------------------------*/
func mgoGetDeviceProfileById(d *models.DeviceProfile, id string) error {
	if bson.IsObjectIdHex(id) {
		return mgoGetDeviceProfile(d, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceProfileById Invalid Object ID " + id)
		return err
	}
}
func mgoGetDeviceProfilesByModel(dp *[]models.DeviceProfile, m string) error {
	return mgoGetDeviceProfiles(dp, bson.M{MODEL: m})
}
func mgoGetDeviceProfilesWithLabel(dp *[]models.DeviceProfile, l []string) error {
	return mgoGetDeviceProfiles(dp, bson.M{LABELS: bson.M{"$in": l}})
}
func mgoGetDeviceProfilesByManufacturerModel(dp *[]models.DeviceProfile, man string, mod string) error {
	return mgoGetDeviceProfiles(dp, bson.M{MANUFACTURER: man, MODEL: mod})
}
func mgoGetDeviceProfilesByManufacturer(dp *[]models.DeviceProfile, man string) error {
	return mgoGetDeviceProfiles(dp, bson.M{MANUFACTURER: man})
}
func mgoGetDeviceProfileByName(dp *models.DeviceProfile, n string) error {
	return mgoGetDeviceProfile(dp, bson.M{NAME: n})
}

// Get device profiles with the passed query
func mgoGetDeviceProfiles(d *[]models.DeviceProfile, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DPCOL)

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
func mgoGetDeviceProfile(d *models.DeviceProfile, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DPCOL)

	// Handle the DBRefs
	var mdp MongoDeviceProfile
	err := col.Find(q).One(&mdp)
	if err != nil {
		return err
	}

	*d = mdp.DeviceProfile

	return err
}
func mgoAddNewDeviceProfile(dp *models.DeviceProfile) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DPCOL)
	count, err := col.Find(bson.M{NAME: dp.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return ErrDuplicateName
	}
	for i := 0; i < len(dp.Commands); i++ {
		if err := addCommand(&dp.Commands[i]); err != nil {
			return err
		}
	}
	ts := makeTimestamp()
	dp.Created = ts
	dp.Modified = ts
	dp.Id = bson.NewObjectId()

	mdp := MongoDeviceProfile{DeviceProfile: *dp}

	err = col.Insert(mdp)
	if err != nil {
		return err
	}

	return nil
}
func mgoUpdateDeviceProfile(dp *models.DeviceProfile) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	c := ds.s.DB(DB).C(DPCOL)

	mdp := MongoDeviceProfile{DeviceProfile: *dp}
	mdp.Modified = makeTimestamp()

	if err := c.UpdateId(mdp.Id, mdp); err != nil {
		return err
	}

	return nil
}

// Get the device profiles that are currently using the command
func mgoGetDeviceProfilesUsingCommand(dp *[]models.DeviceProfile, c models.Command) error {
	query := bson.M{"commands": bson.M{"$elemMatch": bson.M{"$id": c.Id}}}
	return mgoGetDeviceProfiles(dp, query)
}

/* -----------------------------------Addressable --------------------------*/
func mgoUpdateAddressable(ra *models.Addressable, r *models.Addressable) error {
	ds := DS.dataStore()

	defer ds.s.Close()
	c := ds.s.DB(DB).C(ADDCOL)
	if ra == nil {
		return nil
	}
	if ra.Name != "" {
		r.Name = ra.Name
	}
	if strings.Compare(ra.Protocol, "HTTP") != 0 || strings.Compare(ra.Protocol, "TCP") != 0 { // TODO create ENUMS that can be unmarshalled by JSON
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
func mgoGetAddressables(d *[]models.Addressable, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(ADDCOL)
	err := col.Find(q).Sort(QUERYTS).All(d)
	if err != nil {
		return err
	}

	return nil
}
func mgoGetAddressableById(a *models.Addressable, id string) error {
	if bson.IsObjectIdHex(id) {
		return getAddressable(a, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetAddressableById Invalid Object ID " + id)
		return err
	}
}
func mgoAddNewAddressable(a *models.Addressable) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(ADDCOL)

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
func mgoGetAddressableByName(a *models.Addressable, n string) error {
	return mgoGetAddressable(a, bson.M{NAME: n})
}
func mgoGetAddressablesByTopic(a *[]models.Addressable, t string) error {
	return mgoGetAddressables(a, bson.M{TOPIC: t})
}
func mgoGetAddressablesByPort(a *[]models.Addressable, p int) error {
	return mgoGetAddressables(a, bson.M{PORT: p})
}
func mgoGetAddressablesByPublisher(a *[]models.Addressable, p string) error {
	return mgoGetAddressables(a, bson.M{PUBLISHER: p})
}
func mgoGetAddressablesByAddress(a *[]models.Addressable, add string) error {
	return mgoGetAddressables(a, bson.M{ADDRESS: add})
}
func mgoGetAddressable(d *models.Addressable, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(ADDCOL)
	err := col.Find(q).One(d)
	if err != nil {
		return err
	}

	return nil
}
func mgoIsAddressableAssociatedToDevice(a models.Addressable) (bool, error) {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DEVICECOL)
	query := bson.M{ADDRESSABLE + ".$id": a.Id}
	count, err := col.Find(query).Count()
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	} else {
		return false, nil
	}
}
func mgoIsAddressableAssociatedToDeviceService(a models.Addressable) (bool, error) {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DSCOL)
	query := bson.M{ADDRESSABLE + ".$id": a.Id}
	count, err := col.Find(query).Count()
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

/* ----------------------------- Device Service ----------------------------------*/
func mgoGetDeviceServiceByName(d *models.DeviceService, n string) error {
	return mgoGetDeviceService(d, bson.M{NAME: n})
}
func mgoGetDeviceServiceById(d *models.DeviceService, id string) error {
	if bson.IsObjectIdHex(id) {
		return mgoGetDeviceService(d, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceServiceByName Invalid Object ID " + id)
		return err
	}
}
func mgoGetAllDeviceServices(d *[]models.DeviceService) error {
	return mgoGetDeviceServices(d, bson.M{})
}
func mgoGetDeviceServicesByAddressableName(d *[]models.DeviceService, an string) error {
	return mgoGetDeviceServices(d, bson.M{ADDRESSABLE + "." + NAME: an})
}
func mgoGetDeviceServicesByAddressableId(d *[]models.DeviceService, id string) error {
	if bson.IsObjectIdHex(id) {
		return mgoGetDeviceServices(d, bson.M{ADDRESSABLE + ".$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetDeviceServicesByAddressableId Invalid Object ID " + id)
		return err
	}
}
func mgoGetDeviceServicesWithLabel(d *[]models.DeviceService, l []string) error {
	return mgoGetDeviceServices(d, bson.M{LABELS: bson.M{"$in": l}})
}
func mgoGetDeviceServices(d *[]models.DeviceService, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DSCOL)
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
func mgoGetDeviceService(d *models.DeviceService, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DSCOL)
	mds := MongoDeviceService{}
	err := col.Find(q).One(&mds)
	if err != nil {
		return err
	}
	*d = mds.DeviceService

	return nil
}
func mgoAddNewDeviceService(d *models.DeviceService) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(DSCOL)

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
	err = col.Insert(mds)
	if err != nil {
		return err
	}

	return nil
}
func mgoUpdateDeviceService(deviceService models.DeviceService) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	c := ds.s.DB(DB).C(DSCOL)

	deviceService.Service.Modified = makeTimestamp()

	// Handle DBRefs
	mds := MongoDeviceService{DeviceService: deviceService}

	return c.UpdateId(deviceService.Service.Id, mds)
}

/* ----------------------Provision Watcher -----------------------------*/
func mgoGetAllProvisionWatchers(pw *[]models.ProvisionWatcher) error {
	return mgoGetProvisionWatchers(pw, bson.M{})
}
func mgoGetProvisionWatcherByName(pw *models.ProvisionWatcher, n string) error {
	return mgoGetProvisionWatcher(pw, bson.M{NAME: n})
}
func mgoGetProvisionWatchersByProfileName(pw *[]models.ProvisionWatcher, n string) error {
	return mgoGetProvisionWatchers(pw, bson.M{PROFILE + "." + NAME: n})
}
func mgoGetProvisionWatchersByIdentifier(pw *[]models.ProvisionWatcher, k string, v string) error {
	return mgoGetProvisionWatchers(pw, bson.M{IDENTIFIERS + "." + k: v})
}
func mgoGetProvisionWatchersByServiceId(pw *[]models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return mgoGetProvisionWatchers(pw, bson.M{SERVICE + ".$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetProvisionWatchersByServiceId Invalid Object ID " + id)
		return err
	}
}
func mgoGetProvisionWatchersByServiceName(pw *[]models.ProvisionWatcher, n string) error {
	return mgoGetProvisionWatchers(pw, bson.M{SERVICE + "." + NAME: n})

}
func mgoGetProvisionWatcherByProfileId(pw *[]models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return mgoGetProvisionWatchers(pw, bson.M{PROFILE + ".$id": bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetProvisionWatcherByProfileId Invalid Object ID " + id)
		return err
	}
}
func mgoGetProvisionWatcherById(pw *models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return mgoGetProvisionWatcher(pw, bson.M{_ID: bson.ObjectIdHex(id)})
	} else {
		err := errors.New("mgoGetProvisionWatcherById Invalid Object ID " + id)
		return err
	}
}
func mgoGetProvisionWatcher(pw *models.ProvisionWatcher, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(PWCOL)

	// Handle DBRefs
	var mpw MongoProvisionWatcher

	err := col.Find(q).One(&mpw)
	if err != nil {
		return err
	}

	*pw = mpw.ProvisionWatcher

	return err
}
func mgoGetProvisionWatchers(pw *[]models.ProvisionWatcher, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(PWCOL)

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
func mgoAddProvisionWatcher(pw *models.ProvisionWatcher) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(PWCOL)
	count, err := col.Find(bson.M{NAME: pw.Name}).Count()
	if err != nil {
		return err
	} else if count > 0 {
		return ErrDuplicateName
	}

	// get Device Service
	var dev models.DeviceService
	if pw.Service.Service.Id.Hex() != "" {
		mgoGetDeviceServiceById(&dev, pw.Service.Service.Id.Hex())
	} else if pw.Service.Service.Name != "" {
		mgoGetDeviceServiceByName(&dev, pw.Service.Service.Name)
	} else {
		return errors.New("Device Service ID or Name is required")
	}
	pw.Service = dev

	// get Device Profile
	var dp models.DeviceProfile
	if pw.Profile.Id.Hex() != "" {
		mgoGetDeviceProfileById(&dp, pw.Profile.Id.Hex())
	} else if pw.Profile.Name != "" {
		mgoGetDeviceProfileByName(&dp, pw.Profile.Name)
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

	if err := col.Insert(mpw); err != nil {
		return err
	}

	return nil
}
func mgoUpdateProvisionWatcher(pw models.ProvisionWatcher) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	c := ds.s.DB(DB).C(PWCOL)

	pw.Modified = makeTimestamp()

	// Handle DBRefs
	mpw := MongoProvisionWatcher{ProvisionWatcher: pw}

	return c.UpdateId(mpw.Id, mpw)
}

/* ------------------------Command -------------------------------------*/
func mgoGetAllCommands(d *[]models.Command) error {
	return mgoGetCommands(d, bson.M{})
}

func mgoGetCommands(d *[]models.Command, q bson.M) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(COMCOL)
	err := col.Find(q).Sort(QUERYTS).All(d)
	if err != nil {
		return err
	}
	return nil
}
func mgoGetCommandById(d *models.Command, id string) error {
	if bson.IsObjectIdHex(id) {
		ds := DS.dataStore()
		defer ds.s.Close()
		col := ds.s.DB(DB).C(COMCOL)
		if err := col.Find(bson.M{_ID: bson.ObjectIdHex(id)}).One(d); err != nil {
			return err
		}

		return nil
	} else {
		err := errors.New("mgoGetCommandById Invalid Object ID " + id)
		return err
	}
}
func mgoGetCommandByName(c *[]models.Command, n string) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(COMCOL)
	if err := col.Find(bson.M{NAME: n}).All(c); err != nil {
		return err
	}

	return nil
}
func mgoAddCommand(c *models.Command) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(COMCOL)

	ts := makeTimestamp()
	c.Created = ts
	c.Id = bson.NewObjectId()
	if err := col.Insert(c); err != nil {
		return err
	}

	return nil
}

// Update command uses the ID of the command for identification
func mgoUpdateCommand(c *models.Command, r *models.Command) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(COMCOL)
	if c == nil {
		return nil
	}

	// Check if the command has a valid ID
	if len(c.Id.Hex()) == 0 || !bson.IsObjectIdHex(c.Id.Hex()) {
		err := errors.New("ID required for updating a command")
		return err
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

	if err := col.UpdateId(r.Id, r); err != nil {
		return err
	}
	return nil
}

// Delete the command by ID
// Check if the command is still in use by device profiles
func mgoDeleteCommandById(id string) error {
	ds := DS.dataStore()
	defer ds.s.Close()
	col := ds.s.DB(DB).C(COMCOL)

	if !bson.IsObjectIdHex(id) {
		return errors.New("Invalid ID")
	}

	// Check if the command is still in use
	query := bson.M{"commands": bson.M{"$elemMatch": bson.M{"_id": bson.ObjectIdHex(id)}}}
	count, err := ds.s.DB(DB).C(DPCOL).Find(query).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrCommandStillInUse
	}

	return col.RemoveId(bson.ObjectIdHex(id))
}
