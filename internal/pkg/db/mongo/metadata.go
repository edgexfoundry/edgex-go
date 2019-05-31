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

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

/* ----------------------Device Report --------------------------*/

func (mc MongoClient) GetAllDeviceReports() ([]contract.DeviceReport, error) {
	return mc.getDeviceReports(bson.M{})
}

func (mc MongoClient) GetDeviceReportByName(n string) (contract.DeviceReport, error) {
	return mc.getDeviceReport(bson.M{"name": n})
}

func (mc MongoClient) GetDeviceReportByDeviceName(n string) ([]contract.DeviceReport, error) {
	return mc.getDeviceReports(bson.M{"device": n})
}

func (mc MongoClient) GetDeviceReportById(id string) (contract.DeviceReport, error) {
	query, err := idToBsonM(id)
	if err != nil {
		return contract.DeviceReport{}, err
	}
	return mc.getDeviceReport(query)
}

func (mc MongoClient) GetDeviceReportsByAction(n string) ([]contract.DeviceReport, error) {
	return mc.getDeviceReports(bson.M{"action": n})
}

func (mc MongoClient) getDeviceReports(q bson.M) ([]contract.DeviceReport, error) {
	s := mc.session.Copy()
	defer s.Close()

	var drs []models.DeviceReport
	err := s.DB(mc.database.Name).C(db.DeviceReport).Find(q).Sort("queryts").All(&drs)
	if err != nil {
		return []contract.DeviceReport{}, errorMap(err)
	}

	mapped := make([]contract.DeviceReport, 0)
	for _, dr := range drs {
		mapped = append(mapped, dr.ToContract())
	}
	return mapped, nil
}

func (mc MongoClient) getDeviceReport(q bson.M) (contract.DeviceReport, error) {
	s := mc.session.Copy()
	defer s.Close()

	var d models.DeviceReport
	err := s.DB(mc.database.Name).C(db.DeviceReport).Find(q).One(&d)
	if err != nil {
		return contract.DeviceReport{}, errorMap(err)
	}
	return d.ToContract(), nil
}

func (mc MongoClient) AddDeviceReport(d contract.DeviceReport) (string, error) {
	s := mc.session.Copy()
	defer s.Close()

	col := s.DB(mc.database.Name).C(db.DeviceReport)
	count, err := col.Find(bson.M{"name": d.Name}).Count()
	if err != nil {
		return "", errorMap(err)
	}
	if count > 0 {
		return "", db.ErrNotUnique
	}

	var mapped models.DeviceReport
	id, err := mapped.FromContract(d)
	if err != nil {
		return "", errors.New("FromContract failed")
	}

	mapped.TimestampForAdd()

	if err = col.Insert(mapped); err != nil {
		return "", errorMap(err)
	}
	return id, nil
}

func (mc MongoClient) UpdateDeviceReport(dr contract.DeviceReport) error {
	var mapped models.DeviceReport
	id, err := mapped.FromContract(dr)
	if err != nil {
		return errors.New("FromContract failed")
	}

	mapped.TimestampForUpdate()

	return mc.updateId(db.DeviceReport, id, mapped)
}

func (mc MongoClient) DeleteDeviceReportById(id string) error {
	return mc.deleteById(db.DeviceReport, id)
}

/* ----------------------------- Device ---------------------------------- */
func (mc MongoClient) AddDevice(d contract.Device, commands []contract.Command) (string, error) {
	s := mc.session.Copy()
	defer s.Close()

	col := s.DB(mc.database.Name).C(db.Device)

	// Check if the name exist (Device names must be unique)
	count, err := col.Find(bson.M{"name": d.Name}).Count()
	if err != nil {
		return "", errorMap(err)
	}
	if count > 0 {
		return "", db.ErrNotUnique
	}

	var mapped models.Device
	id, err := mapped.FromContract(d, mc, mc, mc)
	if err != nil {
		return "", errors.New("FromContract failed")
	}

	mapped.TimestampForAdd()
	if err = errorMap(col.Insert(mapped)); err != nil {
		return id, err
	}

	//add commands based on DeviceProfile.CommandProfile
	err = mc.addCommands(commands, mapped.Uuid, mapped.Name)
	return id, errorMap(err)
}

func (mc MongoClient) UpdateDevice(d contract.Device) error {
	var mapped models.Device
	id, err := mapped.FromContract(d, mc, mc, mc)
	if err != nil {
		return errors.New("FromContract failed")
	}

	mapped.TimestampForUpdate()

	return mc.updateId(db.Device, id, mapped)
}

func (mc MongoClient) DeleteDeviceById(id string) error {
	err := mc.deleteById(db.Device, id)
	if err != nil {
		return err
	}
	return mc.deleteCommandByDeviceId(id)
}

func (mc MongoClient) GetAllDevices() ([]contract.Device, error) {
	return mc.getDevices(nil)
}

func (mc MongoClient) GetDevicesByProfileId(id string) ([]contract.Device, error) {
	dp, err := mc.getDeviceProfileById(id)
	if err != nil {
		return []contract.Device{}, err
	}
	return mc.getDevices(bson.M{"profile.$id": dp.Id})
}

func (mc MongoClient) GetDeviceById(id string) (contract.Device, error) {
	query, err := idToBsonM(id)
	if err != nil {
		return contract.Device{}, err
	}
	return mc.getDevice(query)
}

func (mc MongoClient) GetDeviceByName(n string) (contract.Device, error) {
	return mc.getDevice(bson.M{"name": n})
}

func (mc MongoClient) GetDevicesByServiceId(id string) ([]contract.Device, error) {
	ds, err := mc.getDeviceServiceById(id)
	if err != nil {
		return []contract.Device{}, err
	}
	return mc.getDevices(bson.M{"service.$id": ds.Id})
}

func (mc MongoClient) GetDevicesWithLabel(l string) ([]contract.Device, error) {
	return mc.getDevices(bson.M{"labels": bson.M{"$in": []string{l}}})
}

func (mc MongoClient) getDevices(q bson.M) ([]contract.Device, error) {
	s := mc.session.Copy()
	defer s.Close()

	var mds []models.Device
	err := s.DB(mc.database.Name).C(db.Device).Find(q).Sort("queryts").All(&mds)
	if err != nil {
		return []contract.Device{}, errorMap(err)
	}

	res := make([]contract.Device, 0)
	for _, md := range mds {
		d, err := md.ToContract(mc, mc, mc)
		if err != nil {
			return []contract.Device{}, err
		}
		res = append(res, d)
	}

	return res, nil
}

func (mc MongoClient) getDevice(q bson.M) (contract.Device, error) {
	s := mc.session.Copy()
	defer s.Close()

	var d models.Device
	if err := s.DB(mc.database.Name).C(db.Device).Find(q).One(&d); err != nil {
		return contract.Device{}, errorMap(err)
	}
	return d.ToContract(mc, mc, mc)
}

/* -----------------------------Device Profile -----------------------------*/

func (mc MongoClient) DBRefToDeviceProfile(dbRef mgo.DBRef) (a models.DeviceProfile, err error) {
	s := mc.session.Copy()
	defer s.Close()

	if err = s.DB(mc.database.Name).C(db.DeviceProfile).Find(bson.M{"_id": dbRef.Id}).One(&a); err != nil {
		return models.DeviceProfile{}, errorMap(err)
	}
	return
}

func (mc MongoClient) DeviceProfileToDBRef(model models.DeviceProfile) (dbRef mgo.DBRef, err error) {
	// validate model with identity provided in contract actually exists
	if model.Id.Valid() {
		model, err = mc.getDeviceProfileById(model.Id.Hex())
	} else {
		model, err = mc.getDeviceProfileById(model.Uuid)
	}
	if err != nil {
		return
	}

	dbRef = mgo.DBRef{Collection: db.DeviceProfile, Id: model.Id}
	return
}

func (mc MongoClient) GetDeviceProfileById(id string) (contract.DeviceProfile, error) {
	model, err := mc.getDeviceProfileById(id)
	if err != nil {
		return contract.DeviceProfile{}, err
	}
	return model.ToContract()
}

func (mc MongoClient) getDeviceProfileById(id string) (models.DeviceProfile, error) {
	query, err := idToBsonM(id)
	if err != nil {
		return models.DeviceProfile{}, err
	}
	return mc.getDeviceProfile(query)
}

func (mc MongoClient) GetAllDeviceProfiles() ([]contract.DeviceProfile, error) {
	return mc.getDeviceProfiles(nil)
}

func (mc MongoClient) GetDeviceProfilesByModel(model string) ([]contract.DeviceProfile, error) {
	return mc.getDeviceProfiles(bson.M{"model": model})
}

func (mc MongoClient) GetDeviceProfilesWithLabel(l string) ([]contract.DeviceProfile, error) {
	return mc.getDeviceProfiles(bson.M{"labels": bson.M{"$in": []string{l}}})
}

func (mc MongoClient) GetDeviceProfilesByManufacturerModel(man string, mod string) ([]contract.DeviceProfile, error) {
	return mc.getDeviceProfiles(bson.M{"manufacturer": man, "model": mod})
}

func (mc MongoClient) GetDeviceProfilesByManufacturer(man string) ([]contract.DeviceProfile, error) {
	return mc.getDeviceProfiles(bson.M{"manufacturer": man})
}

func (mc MongoClient) GetDeviceProfileByName(n string) (contract.DeviceProfile, error) {
	model, err := mc.getDeviceProfile(bson.M{"name": n})
	if err != nil {
		return contract.DeviceProfile{}, err
	}
	return model.ToContract()
}

// Get device profiles with the passed query
func (mc MongoClient) getDeviceProfiles(q bson.M) ([]contract.DeviceProfile, error) {
	s := mc.session.Copy()
	defer s.Close()

	var dps []models.DeviceProfile
	err := s.DB(mc.database.Name).C(db.DeviceProfile).Find(q).Sort("queryts").All(&dps)
	if err != nil {
		return []contract.DeviceProfile{}, errorMap(err)
	}

	cdps := make([]contract.DeviceProfile, 0)
	for _, dp := range dps {
		c, err := dp.ToContract()
		if err != nil {
			return []contract.DeviceProfile{}, err
		}
		cdps = append(cdps, c)
	}
	return cdps, nil
}

func (mc MongoClient) getDeviceProfile(q bson.M) (d models.DeviceProfile, err error) {
	s := mc.session.Copy()
	defer s.Close()

	err = s.DB(mc.database.Name).C(db.DeviceProfile).Find(q).One(&d)
	if err != nil {
		return models.DeviceProfile{}, errorMap(err)
	}

	return
}

func (mc MongoClient) AddDeviceProfile(dp contract.DeviceProfile) (string, error) {
	s := mc.session.Copy()
	defer s.Close()
	if len(dp.Name) == 0 {
		return "", db.ErrNameEmpty
	}
	col := s.DB(mc.database.Name).C(db.DeviceProfile)
	count, err := col.Find(bson.M{"name": dp.Name}).Count()
	if err != nil {
		return "", errorMap(err)
	}
	if count > 0 {
		return "", db.ErrNotUnique
	}

	var mapped models.DeviceProfile
	id, err := mapped.FromContract(dp)
	if err != nil {
		return "", err
	}

	mapped.TimestampForAdd()

	if err = col.Insert(mapped); err != nil {
		return "", errorMap(err)
	}

	return id, nil
}

func (mc MongoClient) UpdateDeviceProfile(dp contract.DeviceProfile) error {
	var mapped models.DeviceProfile
	id, err := mapped.FromContract(dp)
	if err != nil {
		return err
	}

	mapped.TimestampForUpdate()

	return mc.updateId(db.DeviceProfile, id, mapped)
}

func (mc MongoClient) DeleteDeviceProfileById(id string) error {
	return mc.deleteById(db.DeviceProfile, id)
}

//  -----------------------------------Addressable --------------------------*/

func (mc MongoClient) DBRefToAddressable(dbRef mgo.DBRef) (a models.Addressable, err error) {
	s := mc.session.Copy()
	defer s.Close()

	if err = s.DB(mc.database.Name).C(db.Addressable).Find(bson.M{"_id": dbRef.Id}).One(&a); err != nil {
		return models.Addressable{}, errorMap(err)
	}
	return
}

func (mc MongoClient) AddressableToDBRef(a models.Addressable) (dbRef mgo.DBRef, err error) {
	// validate addressable identity provided in contract actually exists and populate missing Id, Uuid field
	var addr models.Addressable
	if a.Id.Valid() {
		addr, err = mc.getAddressableById(a.Id.Hex())
	} else {
		addr, err = mc.getAddressableById(a.Uuid)
	}
	if err != nil {
		return
	}

	dbRef = mgo.DBRef{Collection: db.Addressable, Id: addr.Id}
	return
}

func (mc MongoClient) UpdateAddressable(a contract.Addressable) error {
	var mapped models.Addressable
	id, err := mapped.FromContract(a)
	if err != nil {
		return err
	}

	mapped.TimestampForUpdate()

	return mc.updateId(db.Addressable, id, mapped)
}

func (mc MongoClient) GetAddressables() ([]contract.Addressable, error) {
	return mapAddressables(mc.getAddressablesQuery(bson.M{}))
}

func (mc MongoClient) getAddressablesQuery(q bson.M) ([]models.Addressable, error) {
	s := mc.session.Copy()
	defer s.Close()

	items := make([]models.Addressable, 0)
	if err := s.DB(mc.database.Name).C(db.Addressable).Find(q).Sort("queryts").All(&items); err != nil {
		return []models.Addressable{}, errorMap(err)
	}

	return items, nil
}

func (mc MongoClient) GetAddressableById(id string) (contract.Addressable, error) {
	addr, err := mc.getAddressableById(id)
	if err != nil {
		return contract.Addressable{}, err
	}
	return addr.ToContract(), nil
}

func (mc MongoClient) getAddressableById(id string) (models.Addressable, error) {
	query, err := idToBsonM(id)
	if err != nil {
		return models.Addressable{}, err
	}
	return mc.getAddressable(query)
}

func (mc MongoClient) AddAddressable(a contract.Addressable) (string, error) {
	s := mc.session.Copy()
	defer s.Close()

	var mapped models.Addressable
	id, err := mapped.FromContract(a)
	if err != nil {
		return "", err
	}

	col := s.DB(mc.database.Name).C(db.Addressable)

	// check if the name exist
	count, err := col.Find(bson.M{"name": mapped.Name}).Count()
	if err != nil {
		return a.Id, errorMap(err)
	}
	if count > 0 {
		return a.Id, db.ErrNotUnique
	}

	mapped.TimestampForAdd()

	if err = col.Insert(mapped); err != nil {
		return "", errorMap(err)
	}
	return id, nil
}

func (mc MongoClient) GetAddressableByName(n string) (contract.Addressable, error) {
	addr, err := mc.getAddressableByName(n)
	if err != nil {
		return contract.Addressable{}, err
	}
	return addr.ToContract(), nil
}

func (mc MongoClient) getAddressableByName(n string) (models.Addressable, error) {
	addr, err := mc.getAddressable(bson.M{"name": n})
	if err != nil {
		return models.Addressable{}, err
	}
	return addr, nil
}

func (mc MongoClient) GetAddressablesByTopic(t string) ([]contract.Addressable, error) {
	return mapAddressables(mc.getAddressablesQuery(bson.M{"topic": t}))
}

func (mc MongoClient) GetAddressablesByPort(p int) ([]contract.Addressable, error) {
	return mapAddressables(mc.getAddressablesQuery(bson.M{"port": p}))
}

func (mc MongoClient) GetAddressablesByPublisher(p string) ([]contract.Addressable, error) {
	return mapAddressables(mc.getAddressablesQuery(bson.M{"publisher": p}))
}

func (mc MongoClient) GetAddressablesByAddress(add string) ([]contract.Addressable, error) {
	return mapAddressables(mc.getAddressablesQuery(bson.M{"address": add}))
}

func (mc MongoClient) getAddressable(q bson.M) (models.Addressable, error) {
	s := mc.session.Copy()
	defer s.Close()

	var a models.Addressable
	err := s.DB(mc.database.Name).C(db.Addressable).Find(q).One(&a)
	if err != nil {
		return models.Addressable{}, errorMap(err)
	}
	return a, nil
}

func (mc MongoClient) DeleteAddressableById(id string) error {
	return mc.deleteById(db.Addressable, id)
}

func mapAddressables(as []models.Addressable, err error) ([]contract.Addressable, error) {
	mapped := make([]contract.Addressable, 0)
	if err != nil {
		return mapped, errorMap(err)
	}
	for _, a := range as {
		mapped = append(mapped, a.ToContract())
	}
	return mapped, nil
}

/* ----------------------------- Device Service ----------------------------------*/

func (mc MongoClient) DBRefToDeviceService(dbRef mgo.DBRef) (ds models.DeviceService, err error) {
	s := mc.session.Copy()
	defer s.Close()

	if err = s.DB(mc.database.Name).C(db.DeviceService).Find(bson.M{"_id": dbRef.Id}).One(&ds); err != nil {
		return models.DeviceService{}, errorMap(err)
	}
	return
}

func (mc MongoClient) DeviceServiceToDBRef(model models.DeviceService) (dbRef mgo.DBRef, err error) {
	// validate model with identity provided in contract actually exists
	if model.Id.Valid() {
		model, err = mc.getDeviceServiceById(model.Id.Hex())
	} else {
		model, err = mc.getDeviceServiceById(model.Uuid)
	}
	if err != nil {
		return
	}

	dbRef = mgo.DBRef{Collection: db.DeviceService, Id: model.Id}
	return
}

func (mc MongoClient) GetDeviceServiceByName(n string) (contract.DeviceService, error) {
	ds, err := mc.deviceService(bson.M{"name": n})
	if err != nil {
		return contract.DeviceService{}, err
	}
	return ds.ToContract(mc)
}

func (mc MongoClient) GetDeviceServiceById(id string) (contract.DeviceService, error) {
	ds, err := mc.getDeviceServiceById(id)
	if err != nil {
		return contract.DeviceService{}, err
	}
	return ds.ToContract(mc)
}

func (mc MongoClient) getDeviceServiceById(id string) (models.DeviceService, error) {
	query, err := idToBsonM(id)
	if err != nil {
		return models.DeviceService{}, err
	}
	return mc.deviceService(query)
}

func (mc MongoClient) GetAllDeviceServices() ([]contract.DeviceService, error) {
	return mc.getDeviceServices(bson.M{})
}

func (mc MongoClient) GetDeviceServicesByAddressableId(id string) ([]contract.DeviceService, error) {
	addr, err := mc.getAddressableById(id)
	if err != nil {
		return []contract.DeviceService{}, err
	}
	return mc.getDeviceServices(bson.M{"addressable.$id": addr.Id})
}

func (mc MongoClient) GetDeviceServicesWithLabel(l string) ([]contract.DeviceService, error) {
	return mc.getDeviceServices(bson.M{"labels": bson.M{"$in": []string{l}}})
}

func (mc MongoClient) getDeviceServices(q bson.M) (dss []contract.DeviceService, err error) {
	dss = []contract.DeviceService{}
	mds, err := mc.deviceServices(q)
	if err != nil {
		return
	}
	for _, ds := range mds {
		cds, err := ds.ToContract(mc)
		if err != nil {
			return []contract.DeviceService{}, err
		}
		dss = append(dss, cds)
	}

	return dss, nil
}

func (mc MongoClient) deviceServices(q bson.M) (dss []models.DeviceService, err error) {
	s := mc.session.Copy()
	defer s.Close()
	if err = s.DB(mc.database.Name).C(db.DeviceService).Find(q).Sort("queryts").All(&dss); err != nil {
		return []models.DeviceService{}, errorMap(err)
	}

	return
}

func (mc MongoClient) getDeviceService(q bson.M) (ds contract.DeviceService, err error) {
	mds, err := mc.deviceService(q)
	if err != nil {
		return contract.DeviceService{}, err
	}
	ds, err = mds.ToContract(mc)
	return
}

func (mc MongoClient) deviceService(q bson.M) (models.DeviceService, error) {
	s := mc.session.Copy()
	defer s.Close()

	var ds models.DeviceService
	err := s.DB(mc.database.Name).C(db.DeviceService).Find(q).One(&ds)
	if err != nil {
		return models.DeviceService{}, errorMap(err)
	}

	return ds, nil
}

func (mc MongoClient) AddDeviceService(ds contract.DeviceService) (string, error) {
	s := mc.session.Copy()
	defer s.Close()

	var mapped models.DeviceService
	id, err := mapped.FromContract(ds, mc)
	if err != nil {
		return "", errors.New("FromContract failed")
	}

	mapped.TimestampForAdd()

	if err = s.DB(mc.database.Name).C(db.DeviceService).Insert(mapped); err != nil {
		return "", errorMap(err)
	}
	return id, nil
}

func (mc MongoClient) UpdateDeviceService(ds contract.DeviceService) error {
	var mapped models.DeviceService
	id, err := mapped.FromContract(ds, mc)
	if err != nil {
		return errors.New("FromContract failed")
	}

	mapped.TimestampForUpdate()

	return mc.updateId(db.DeviceService, id, mapped)
}

func (mc MongoClient) DeleteDeviceServiceById(id string) error {
	return mc.deleteById(db.DeviceService, id)
}

//  ----------------------Provision Watcher -----------------------------*/
func (mc MongoClient) GetAllProvisionWatchers() (pw []contract.ProvisionWatcher, err error) {
	return mc.getProvisionWatchers(bson.M{})
}

func (mc MongoClient) GetProvisionWatcherByName(n string) (pw contract.ProvisionWatcher, err error) {
	return mc.GetProvisionWatcher(bson.M{"name": n})
}

func (mc MongoClient) GetProvisionWatchersByIdentifier(k string, v string) (pw []contract.ProvisionWatcher, err error) {
	return mc.getProvisionWatchers(bson.M{"identifiers." + k: v})
}

func (mc MongoClient) GetProvisionWatchersByServiceId(id string) (pw []contract.ProvisionWatcher, err error) {
	ds, err := mc.getDeviceServiceById(id)
	if err != nil {
		return []contract.ProvisionWatcher{}, err
	}
	return mc.getProvisionWatchers(bson.M{"service.$id": ds.Id})
}

func (mc MongoClient) GetProvisionWatchersByProfileId(id string) (pw []contract.ProvisionWatcher, err error) {
	dp, err := mc.getDeviceProfileById(id)
	if err != nil {
		return []contract.ProvisionWatcher{}, err
	}
	return mc.getProvisionWatchers(bson.M{"profile.$id": dp.Id})
}

func (mc MongoClient) GetProvisionWatcherById(id string) (pw contract.ProvisionWatcher, err error) {
	query, err := idToBsonM(id)
	if err != nil {
		return contract.ProvisionWatcher{}, err
	}

	pw, err = mc.GetProvisionWatcher(query)
	if err != nil {
		return contract.ProvisionWatcher{}, err
	}

	return
}

func (mc MongoClient) GetProvisionWatcher(q bson.M) (pw contract.ProvisionWatcher, err error) {
	mpw, err := mc.getProvisionWatcher(q)
	if err != nil {
		return
	}

	pw, err = mpw.ToContract(mc, mc, mc)

	return
}

func (mc MongoClient) getProvisionWatcher(q bson.M) (mpw models.ProvisionWatcher, err error) {
	s := mc.session.Copy()
	defer s.Close()

	if err = s.DB(mc.database.Name).C(db.ProvisionWatcher).Find(q).One(&mpw); err != nil {
		return models.ProvisionWatcher{}, errorMap(err)
	}

	return mpw, nil
}

func (mc MongoClient) getProvisionWatchers(q bson.M) (pws []contract.ProvisionWatcher, err error) {
	mpws, err := mc.provisionWatchers(q)

	cpws := make([]contract.ProvisionWatcher, 0)
	for _, mpw := range mpws {
		cpw, err := mpw.ToContract(mc, mc, mc)
		if err != nil {
			return []contract.ProvisionWatcher{}, err
		}
		cpws = append(cpws, cpw)
	}

	return cpws, nil
}

func (mc MongoClient) provisionWatchers(q bson.M) (pws []models.ProvisionWatcher, err error) {
	s := mc.session.Copy()
	defer s.Close()

	var mpws []models.ProvisionWatcher
	if err = s.DB(mc.database.Name).C(db.ProvisionWatcher).Find(q).Sort("queryts").All(&mpws); err != nil {
		return []models.ProvisionWatcher{}, errorMap(err)
	}

	return mpws, nil
}

func (mc MongoClient) AddProvisionWatcher(pw contract.ProvisionWatcher) (string, error) {
	s := mc.session.Copy()
	defer s.Close()
	col := s.DB(mc.database.Name).C(db.ProvisionWatcher)
	count, err := col.Find(bson.M{"name": pw.Name}).Count()
	if err != nil {
		return "", errorMap(err)
	}
	if count > 0 {
		return "", db.ErrNotUnique
	}

	// get Device Service
	var dev contract.DeviceService
	switch {
	case pw.Service.Id != "":
		dev, err = mc.GetDeviceServiceById(pw.Service.Id)
	case pw.Service.Name != "":
		dev, err = mc.GetDeviceServiceByName(pw.Service.Name)
	default:
		return "", errors.New("Device Service ID or Name is required")
	}
	if err != nil {
		return "", err
	}
	pw.Service = dev

	// get Device Profile
	var dp contract.DeviceProfile
	switch {
	case pw.Profile.Id != "":
		dp, err = mc.GetDeviceProfileById(pw.Profile.Id)
	case pw.Profile.Name != "":
		dp, err = mc.GetDeviceProfileByName(pw.Profile.Name)
	default:
		return "", errors.New("Device Profile ID or Name is required")
	}
	if err != nil {
		return "", err
	}
	pw.Profile = dp

	var mapped models.ProvisionWatcher
	id, err := mapped.FromContract(pw, mc, mc, mc)
	if err != nil {
		return "", errors.New("ProvisionWatcher FromContract() failed")
	}

	mapped.TimestampForAdd()

	err = col.Insert(mapped)
	if err != nil {
		return "", errorMap(err)
	}
	return id, nil
}

func (mc MongoClient) UpdateProvisionWatcher(pw contract.ProvisionWatcher) error {
	var mapped models.ProvisionWatcher
	id, err := mapped.FromContract(pw, mc, mc, mc)
	if err != nil {
		return err
	}

	mapped.TimestampForUpdate()

	return mc.updateId(db.ProvisionWatcher, id, mapped)
}

func (mc MongoClient) DeleteProvisionWatcherById(id string) error {
	return mc.deleteById(db.ProvisionWatcher, id)
}

//  ------------------------Command -------------------------------------*/

func (mc MongoClient) GetAllCommands() ([]contract.Command, error) {
	s := mc.session.Copy()
	defer s.Close()

	var commands []models.Command
	err := s.DB(mc.database.Name).C(db.Command).Find(bson.M{}).Sort("queryts").All(&commands)
	return mapCommands(commands, err)
}

func (mc MongoClient) GetCommandById(id string) (contract.Command, error) {
	command, err := mc.getCommandById(id)
	if err != nil {
		return contract.Command{}, err
	}
	return command.ToContract(), nil
}

func (mc MongoClient) getCommandById(id string) (models.Command, error) {
	s := mc.session.Copy()
	defer s.Close()

	query, err := idToBsonM(id)
	if err != nil {
		return models.Command{}, err
	}

	var command models.Command
	if err := s.DB(mc.database.Name).C(db.Command).Find(query).One(&command); err != nil {
		return models.Command{}, errorMap(err)
	}
	return command, nil
}

func (mc MongoClient) GetCommandByName(n string) ([]contract.Command, error) {
	s := mc.session.Copy()
	defer s.Close()

	var commands []models.Command
	err := s.DB(mc.database.Name).C(db.Command).Find(bson.M{"name": n}).All(&commands)

	return mapCommands(commands, err)
}

func (mc MongoClient) GetCommandsByDeviceId(did string) ([]contract.Command, error) {
	s := mc.session.Copy()
	defer s.Close()

	var commands []models.Command
	err := s.DB(mc.database.Name).C(db.Command).Find(bson.M{"deviceId": did}).All(&commands)

	return mapCommands(commands, err)
}

func (mc MongoClient) deleteCommandByDeviceId(did string) error {
	s := mc.session.Copy()
	defer s.Close()
	_, err := s.DB(mc.database.Name).C(db.Command).RemoveAll(bson.D{{Name: "deviceId", Value: did}})
	return errorMap(err)
}

func (mc MongoClient) addCommands(commands []contract.Command, did string, dname string) (err error) {
	s := mc.session.Copy()
	defer s.Close()
	for _, c := range commands {
		var cMapped models.Command
		if _, err = cMapped.FromContract(c, did, dname); err != nil {
			return errors.New("FromContract failed")
		}
		cMapped.TimestampForAdd()
		if err = s.DB(mc.database.Name).C(db.Command).Insert(cMapped); err != nil {
			return err
		}
	}
	return err
}

func mapCommands(commands []models.Command, err error) ([]contract.Command, error) {
	if err != nil {
		return nil, errorMap(err)
	}

	mapped := make([]contract.Command, 0)
	for _, cmd := range commands {
		mapped = append(mapped, cmd.ToContract())
	}

	return mapped, nil
}

// Scrub all metadata
func (mc MongoClient) ScrubMetadata() error {
	s := mc.session.Copy()
	defer s.Close()

	_, err := s.DB(mc.database.Name).C(db.Addressable).RemoveAll(nil)
	if err != nil {
		return errorMap(err)
	}
	_, err = s.DB(mc.database.Name).C(db.DeviceService).RemoveAll(nil)
	if err != nil {
		return errorMap(err)
	}
	_, err = s.DB(mc.database.Name).C(db.DeviceProfile).RemoveAll(nil)
	if err != nil {
		return errorMap(err)
	}
	_, err = s.DB(mc.database.Name).C(db.DeviceReport).RemoveAll(nil)
	if err != nil {
		return errorMap(err)
	}
	_, err = s.DB(mc.database.Name).C(db.Device).RemoveAll(nil)
	if err != nil {
		return errorMap(err)
	}
	_, err = s.DB(mc.database.Name).C(db.ProvisionWatcher).RemoveAll(nil)
	if err != nil {
		return errorMap(err)
	}

	return nil
}
