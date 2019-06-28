/*******************************************************************************
 * Copyright 2018 Redis Labs Inc.
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
package redis

import (
	"errors"
	"fmt"
	"strconv"

	types "github.com/edgexfoundry/edgex-go/internal/core/metadata/errors"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

// /* ----------------------Device Report --------------------------*/
func (c *Client) GetAllDeviceReports() ([]contract.DeviceReport, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.DeviceReport, 0, -1)
	if err != nil {
		return []contract.DeviceReport{}, err
	}

	d := make([]contract.DeviceReport, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &d[i])
		if err != nil {
			return []contract.DeviceReport{}, err
		}
	}

	return d, nil
}

func (c *Client) GetDeviceReportByName(n string) (contract.DeviceReport, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	d := contract.DeviceReport{}
	err := getObjectByHash(conn, db.DeviceReport+":name", n, unmarshalObject, &d)

	return d, err
}

func (c *Client) GetDeviceReportByDeviceName(n string) ([]contract.DeviceReport, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, db.DeviceReport+":device:"+n)
	if err != nil {
		return []contract.DeviceReport{}, err
	}

	d := make([]contract.DeviceReport, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &d[i])
		if err != nil {
			return []contract.DeviceReport{}, err
		}
	}

	return d, nil
}

func (c *Client) GetDeviceReportById(id string) (contract.DeviceReport, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var d contract.DeviceReport
	err := getObjectById(conn, id, unmarshalObject, &d)
	return d, err
}

func (c *Client) GetDeviceReportsByScheduleEventName(n string) ([]contract.DeviceReport, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, db.DeviceReport+":scheduleevent:"+n)
	if err != nil {
		return []contract.DeviceReport{}, err
	}

	d := make([]contract.DeviceReport, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &d[i])
		if err != nil {
			return []contract.DeviceReport{}, err
		}
	}

	return d, nil
}

func (c *Client) GetDeviceReportsByAction(n string) ([]contract.DeviceReport, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, db.DeviceReport+":action:"+n)
	if err != nil {
		return []contract.DeviceReport{}, err
	}

	d := make([]contract.DeviceReport, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &d[i])
		if err != nil {
			return []contract.DeviceReport{}, err
		}
	}

	return d, nil
}

func (c *Client) AddDeviceReport(d contract.DeviceReport) (string, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	return addDeviceReport(conn, d)
}

func (c *Client) UpdateDeviceReport(dr contract.DeviceReport) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteDeviceReport(conn, dr.Id)
	if err != nil {
		return err
	}

	_, err = addDeviceReport(conn, dr)
	return err
}

func (c *Client) DeleteDeviceReportById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteDeviceReport(conn, id)
}

func addDeviceReport(conn redis.Conn, dr contract.DeviceReport) (string, error) {
	exists, err := redis.Bool(conn.Do("HEXISTS", db.DeviceReport+":name", dr.Name))
	if err != nil {
		return "", err
	} else if exists {
		return "", db.ErrNotUnique
	}

	_, err = uuid.Parse(dr.Id)
	if err != nil {
		dr.Id = uuid.New().String()
	}
	id := dr.Id

	ts := db.MakeTimestamp()
	if dr.Created == 0 {
		dr.Created = ts
	}
	dr.Modified = ts

	m, err := marshalObject(dr)
	if err != nil {
		return "", err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", id, m)
	_ = conn.Send("ZADD", db.DeviceReport, 0, id)
	_ = conn.Send("SADD", db.DeviceReport+":device:"+dr.Device, id)
	_ = conn.Send("SADD", db.DeviceReport+":action:"+dr.Action, id)
	_ = conn.Send("HSET", db.DeviceReport+":name", dr.Name, id)
	_, err = conn.Do("EXEC")
	return id, err
}

func deleteDeviceReport(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	dr := contract.DeviceReport{}
	_ = unmarshalObject(object, &dr)

	_ = conn.Send("MULTI")
	_ = conn.Send("DEL", id)
	_ = conn.Send("ZREM", db.DeviceReport, id)
	_ = conn.Send("SREM", db.DeviceReport+":device:"+dr.Device, id)
	_ = conn.Send("SREM", db.DeviceReport+":action:"+dr.Action, id)
	_ = conn.Send("HDEL", db.DeviceReport+":name", dr.Name)
	_, err = conn.Do("EXEC")
	return err
}

// /* ----------------------------- Device ---------------------------------- */
func (c *Client) AddDevice(d contract.Device, commands []contract.Command) (string, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	return addDevice(conn, d, commands)
}

func (c *Client) UpdateDevice(d contract.Device) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteDevice(conn, d.Id)
	if err != nil {
		return err
	}

	_, err = addDevice(conn, d, d.Profile.CoreCommands)
	return err
}

func (c *Client) DeleteDeviceById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteDevice(conn, id)
}

func (c *Client) GetAllDevices() ([]contract.Device, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Device, 0, -1)
	if err != nil {
		return []contract.Device{}, err
	}

	d := make([]contract.Device, len(objects))
	for i, object := range objects {
		err = unmarshalDevice(object, &d[i])
		if err != nil {
			return []contract.Device{}, err
		}
	}

	return d, nil
}

func (c *Client) GetDevicesByProfileId(id string) ([]contract.Device, error) {
	return c.getDevicesByValue(db.Device + ":profile:" + id)
}

func (c *Client) GetDeviceById(id string) (contract.Device, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var d contract.Device
	err := getObjectById(conn, id, unmarshalDevice, &d)
	return d, err
}

func (c *Client) GetDeviceByName(n string) (contract.Device, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var d contract.Device
	err := getObjectByHash(conn, db.Device+":name", n, unmarshalDevice, &d)
	return d, err
}

func (c *Client) GetDevicesByServiceId(id string) ([]contract.Device, error) {
	return c.getDevicesByValue(db.Device + ":service:" + id)
}

func (c *Client) GetDevicesWithLabel(l string) ([]contract.Device, error) {
	return c.getDevicesByValue(db.Device + ":label:" + l)
}

func (c *Client) getDevicesByValue(v string) ([]contract.Device, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, v)
	if err != nil {
		return []contract.Device{}, err
	}

	d := make([]contract.Device, len(objects))
	for i, object := range objects {
		err = unmarshalDevice(object, &d[i])
		if err != nil {
			return []contract.Device{}, err
		}
	}

	return d, nil
}

func addDevice(conn redis.Conn, d contract.Device, commands []contract.Command) (string, error) {
	exists, err := redis.Bool(conn.Do("HEXISTS", db.Device+":name", d.Name))
	if err != nil {
		return "", err
	} else if exists {
		return "", db.ErrNotUnique
	}

	_, err = uuid.Parse(d.Id)
	if err != nil {
		d.Id = uuid.New().String()
	}
	id := d.Id

	ts := db.MakeTimestamp()
	if d.Created == 0 {
		d.Created = ts
	}
	d.Modified = ts

	m, err := marshalDevice(d)
	if err != nil {
		return "", err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", id, m)
	_ = conn.Send("ZADD", db.Device, 0, id)
	_ = conn.Send("HSET", db.Device+":name", d.Name, id)
	_ = conn.Send("SADD", db.Device+":service:"+d.Service.Id, id)
	_ = conn.Send("SADD", db.Device+":profile:"+d.Profile.Id, id)
	for _, label := range d.Labels {
		_ = conn.Send("SADD", db.Device+":label:"+label, id)
	}
	//add commands
	for _, c := range commands {
		cid, err := addCommand(conn, false, c)
		if err != nil {
			return "", err
		}
		_ = conn.Send("SADD", db.Command+":device:"+id, cid)
	}
	_, err = conn.Do("EXEC")
	return id, err
}

func deleteDevice(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	d := contract.Device{}
	_ = unmarshalDevice(object, &d)

	cmds, err := getCommandsByDeviceId(conn, id)
	if err != nil {
		return err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("DEL", id)
	_ = conn.Send("ZREM", db.Device, id)
	_ = conn.Send("HDEL", db.Device+":name", d.Name)
	_ = conn.Send("SREM", db.Device+":service:"+d.Service.Id, id)
	_ = conn.Send("SREM", db.Device+":profile:"+d.Profile.Id, id)
	for _, label := range d.Labels {
		_ = conn.Send("SREM", db.Device+":label:"+label, id)
	}

	for _, c := range cmds {
		deleteCommand(conn, c)
		_ = conn.Send("SREM", db.Command+":device:"+id, c.Id)
	}

	_, err = conn.Do("EXEC")
	return err
}

// /* -----------------------------Device Profile -----------------------------*/
func (c *Client) GetDeviceProfileById(id string) (contract.DeviceProfile, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var d contract.DeviceProfile
	err := getObjectById(conn, id, unmarshalDeviceProfile, &d)
	return d, err
}

func (c *Client) GetAllDeviceProfiles() ([]contract.DeviceProfile, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.DeviceProfile, 0, -1)
	if err != nil {
		return []contract.DeviceProfile{}, err
	}

	dp := make([]contract.DeviceProfile, len(objects))
	for i, object := range objects {
		err = unmarshalDeviceProfile(object, &dp[i])
		if err != nil {
			return []contract.DeviceProfile{}, err
		}
	}

	return dp, nil
}

func (c *Client) GetDeviceProfilesByModel(model string) ([]contract.DeviceProfile, error) {
	return c.getDeviceProfilesByValues(db.DeviceProfile + ":model:" + model)
}

func (c *Client) GetDeviceProfilesWithLabel(l string) ([]contract.DeviceProfile, error) {
	return c.getDeviceProfilesByValues(db.DeviceProfile + ":label:" + l)
}

func (c *Client) GetDeviceProfilesByManufacturerModel(man string, mod string) ([]contract.DeviceProfile, error) {
	return c.getDeviceProfilesByValues(db.DeviceProfile+":manufacturer:"+man, db.DeviceProfile+":model:"+mod)
}

func (c *Client) GetDeviceProfilesByManufacturer(man string) ([]contract.DeviceProfile, error) {
	return c.getDeviceProfilesByValues(db.DeviceProfile + ":manufacturer:" + man)
}

func (c *Client) GetDeviceProfileByName(n string) (contract.DeviceProfile, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var dp contract.DeviceProfile
	err := getObjectByHash(conn, db.DeviceProfile+":name", n, unmarshalDeviceProfile, &dp)
	return dp, err
}

// Get device profiles with the passed query
func (c *Client) getDeviceProfilesByValues(vals ...string) ([]contract.DeviceProfile, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValues(conn, vals...)
	if err != nil {
		return []contract.DeviceProfile{}, err
	}

	dp := make([]contract.DeviceProfile, len(objects))
	for i, object := range objects {
		err = unmarshalDeviceProfile(object, &dp[i])
		if err != nil {
			return []contract.DeviceProfile{}, err
		}
	}

	return dp, nil
}

func (c *Client) AddDeviceProfile(dp contract.DeviceProfile) (string, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	return addDeviceProfile(conn, dp)
}

func (c *Client) UpdateDeviceProfile(dp contract.DeviceProfile) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteDeviceProfile(conn, dp.Id)
	if err != nil {
		return err
	}

	_, err = addDeviceProfile(conn, dp)
	return err
}

func (c *Client) DeleteDeviceProfileById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteDeviceProfile(conn, id)
}

func addDeviceProfile(conn redis.Conn, dp contract.DeviceProfile) (string, error) {
	exists, err := redis.Bool(conn.Do("HEXISTS", db.DeviceProfile+":name", dp.Name))
	if err != nil {
		return "", err
	} else if exists {
		return "", db.ErrNotUnique
	}

	_, err = uuid.Parse(dp.Id)
	if err != nil {
		dp.Id = uuid.New().String()
	}
	id := dp.Id

	ts := db.MakeTimestamp()
	if dp.Created == 0 {
		dp.Created = ts
	}
	dp.Modified = ts

	m, err := marshalDeviceProfile(dp)
	if err != nil {
		return "", err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", id, m)
	_ = conn.Send("ZADD", db.DeviceProfile, 0, id)
	_ = conn.Send("HSET", db.DeviceProfile+":name", dp.Name, id)
	_ = conn.Send("SADD", db.DeviceProfile+":manufacturer:"+dp.Manufacturer, id)
	_ = conn.Send("SADD", db.DeviceProfile+":model:"+dp.Model, id)
	for _, label := range dp.Labels {
		_ = conn.Send("SADD", db.DeviceProfile+":label:"+label, id)
	}

	_, err = conn.Do("EXEC")
	return id, err
}

func deleteDeviceProfile(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	dp := contract.DeviceProfile{}
	_ = unmarshalDeviceProfile(object, &dp)

	_ = conn.Send("MULTI")
	_ = conn.Send("DEL", id)
	_ = conn.Send("ZREM", db.DeviceProfile, id)
	_ = conn.Send("HDEL", db.DeviceProfile+":name", dp.Name)
	_ = conn.Send("SREM", db.DeviceProfile+":manufacturer:"+dp.Manufacturer, id)
	_ = conn.Send("SREM", db.DeviceProfile+":model:"+dp.Model, id)
	for _, label := range dp.Labels {
		_ = conn.Send("SREM", db.DeviceProfile+":label:"+label, id)
	}

	_, err = conn.Do("EXEC")
	return err
}

/* -----------------------------------Addressable -------------------------- */
//func (c *Client) UpdateAddressable(updated *contract.Addressable, orig *contract.Addressable) error {
func (c *Client) UpdateAddressable(a contract.Addressable) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteAddressable(conn, a.Id)
	if err != nil {
		return err
	}

	_, err = addAddressable(conn, a)
	return err
}

func (c *Client) AddAddressable(a contract.Addressable) (string, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	return addAddressable(conn, a)
}

func (c *Client) GetAddressableById(id string) (contract.Addressable, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var a contract.Addressable
	err := getObjectById(conn, id, unmarshalObject, &a)
	return a, err
}

func (c *Client) GetAddressableByName(n string) (contract.Addressable, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var a contract.Addressable
	err := getObjectByHash(conn, db.Addressable+":name", n, unmarshalObject, &a)
	return a, err
}

func (c *Client) GetAddressablesByTopic(t string) ([]contract.Addressable, error) {
	return c.getAddressablesByValue(db.Addressable + ":topic:" + t)
}

func (c *Client) GetAddressablesByPort(p int) ([]contract.Addressable, error) {
	return c.getAddressablesByValue(db.Addressable + ":port:" + strconv.Itoa(p))
}

func (c *Client) GetAddressablesByPublisher(p string) ([]contract.Addressable, error) {
	return c.getAddressablesByValue(db.Addressable + ":publisher:" + p)
}

func (c *Client) GetAddressablesByAddress(add string) ([]contract.Addressable, error) {
	return c.getAddressablesByValue(db.Addressable + ":address:" + add)
}

func (c *Client) GetAddressables() ([]contract.Addressable, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Addressable, 0, -1)
	if err != nil {
		return []contract.Addressable{}, err
	}

	d := make([]contract.Addressable, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &d[i])
		if err != nil {
			return []contract.Addressable{}, err
		}
	}

	return d, nil
}

func (c *Client) DeleteAddressableById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteAddressable(conn, id)
}

func (c *Client) getAddressablesByValue(v string) ([]contract.Addressable, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, v)
	if err != nil {
		return []contract.Addressable{}, err
	}

	a := make([]contract.Addressable, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &a[i])
		if err != nil {
			return []contract.Addressable{}, err
		}
	}

	return a, nil
}

func addAddressable(conn redis.Conn, a contract.Addressable) (string, error) {
	exists, err := redis.Bool(conn.Do("HEXISTS", db.Addressable+":name", a.Name))
	if err != nil {
		return a.Id, err
	} else if exists {
		return a.Id, db.ErrNotUnique
	}

	_, err = uuid.Parse(a.Id)
	if err != nil {
		a.Id = uuid.New().String()
	}
	id := a.Id

	ts := db.MakeTimestamp()
	if a.Created == 0 {
		a.Created = ts
	}
	a.Modified = ts

	m, err := marshalObject(a)
	if err != nil {
		return a.Id, err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", id, m)
	_ = conn.Send("ZADD", db.Addressable, 0, id)
	_ = conn.Send("SADD", db.Addressable+":topic:"+a.Topic, id)
	_ = conn.Send("SADD", db.Addressable+":port:"+strconv.Itoa(a.Port), id)
	_ = conn.Send("SADD", db.Addressable+":publisher:"+a.Publisher, id)
	_ = conn.Send("SADD", db.Addressable+":address:"+a.Address, id)
	_ = conn.Send("HSET", db.Addressable+":name", a.Name, id)
	_, err = conn.Do("EXEC")
	if err != nil {
		return a.Id, err
	}

	return a.Id, err
}

func deleteAddressable(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	// TODO: ??? check if addressable is used by DeviceServices

	a := contract.Addressable{}
	_ = unmarshalObject(object, &a)

	_ = conn.Send("MULTI")
	_ = conn.Send("DEL", id)
	_ = conn.Send("ZREM", db.Addressable, id)
	_ = conn.Send("SREM", db.Addressable+":topic:"+a.Topic, id)
	_ = conn.Send("SREM", db.Addressable+":port:"+strconv.Itoa(a.Port), id)
	_ = conn.Send("SREM", db.Addressable+":publisher:"+a.Publisher, id)
	_ = conn.Send("SREM", db.Addressable+":address:"+a.Address, id)
	_ = conn.Send("HDEL", db.Addressable+":name", a.Name)
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

// /* ----------------------------- Device Service ----------------------------------*/
func (c *Client) GetDeviceServiceByName(n string) (contract.DeviceService, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var d contract.DeviceService
	err := getObjectByHash(conn, db.DeviceService+":name", n, unmarshalDeviceService, &d)
	return d, err
}

func (c *Client) GetDeviceServiceById(id string) (contract.DeviceService, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var d contract.DeviceService
	err := getObjectById(conn, id, unmarshalDeviceService, &d)
	return d, err
}

func (c *Client) GetAllDeviceServices() ([]contract.DeviceService, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.DeviceService, 0, -1)
	if err != nil {
		return []contract.DeviceService{}, err
	}

	d := make([]contract.DeviceService, len(objects))
	for i, object := range objects {
		err = unmarshalDeviceService(object, &d[i])
		if err != nil {
			return []contract.DeviceService{}, err
		}
	}

	return d, nil
}

func (c *Client) GetDeviceServicesByAddressableId(id string) ([]contract.DeviceService, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, db.DeviceService+":addressable:"+id)
	if err != nil {
		return []contract.DeviceService{}, err
	}

	d := make([]contract.DeviceService, len(objects))
	for i, object := range objects {
		err = unmarshalDeviceService(object, &d[i])
		if err != nil {
			return []contract.DeviceService{}, err
		}
	}
	return d, nil
}

func (c *Client) GetDeviceServicesWithLabel(l string) ([]contract.DeviceService, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, db.DeviceService+":label:"+l)
	if err != nil {
		return []contract.DeviceService{}, err
	}

	d := make([]contract.DeviceService, len(objects))
	for i, object := range objects {
		err = unmarshalDeviceService(object, &d[i])
		if err != nil {
			return []contract.DeviceService{}, err
		}
	}
	return d, nil
}

func (c *Client) AddDeviceService(ds contract.DeviceService) (string, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	return addDeviceService(conn, ds)
}

func (c *Client) UpdateDeviceService(ds contract.DeviceService) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteDeviceService(conn, ds.Id)
	if err != nil {
		return err
	}

	ds.Modified = db.MakeTimestamp()

	_, err = addDeviceService(conn, ds)

	return err
}

func (c *Client) DeleteDeviceServiceById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteDeviceService(conn, id)
}

func addDeviceService(conn redis.Conn, ds contract.DeviceService) (string, error) {
	aid := ds.Addressable.Id
	_ = conn.Send("MULTI")
	_ = conn.Send("HEXISTS", db.DeviceService+":name", ds.Name)
	_ = conn.Send("EXISTS", aid)
	_ = conn.Send("HGET", db.Addressable+":name", ds.Addressable.Name)
	rep, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return ds.Id, err
	}

	// Verify uniquness of name
	nex, err := redis.Bool(rep[0], nil)
	if err != nil {
		return "", err
	} else if nex {
		return "", db.ErrNotUnique
	}

	// Verify existence of an addressable by ID or name
	idex, err := redis.Bool(rep[1], nil)
	if err != nil {
		return ds.Id, err
	} else if !idex {
		aid, err = redis.String(rep[2], nil)
		if err == redis.ErrNil {
			return "", errors.New("Invalid addressable")
		} else if err != nil {
			return "", err
		}
		ds.Addressable.Id = aid // XXX This seems redundant
	}

	_, err = uuid.Parse(ds.Id)
	if err != nil {
		ds.Id = uuid.New().String()
	}
	id := ds.Id

	ts := db.MakeTimestamp()
	if ds.Created == 0 {
		ds.Created = ts
	}
	ds.Modified = ts

	m, err := marshalDeviceService(ds)
	if err != nil {
		return "", err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", id, m)
	_ = conn.Send("ZADD", db.DeviceService, 0, id)
	_ = conn.Send("HSET", db.DeviceService+":name", ds.Name, id)
	_ = conn.Send("SADD", db.DeviceService+":addressable:"+aid, id)
	for _, label := range ds.Labels {
		_ = conn.Send("SADD", db.DeviceService+":label:"+label, id)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		return "", err
	}

	return ds.Id, err
}

func deleteDeviceService(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	ds := contract.DeviceService{}
	_ = unmarshalDeviceService(object, &ds)

	_ = conn.Send("MULTI")
	_ = conn.Send("DEL", id)
	_ = conn.Send("ZREM", db.DeviceService, id)
	_ = conn.Send("HDEL", db.DeviceService+":name", ds.Name)
	_ = conn.Send("SREM", db.DeviceService+":addressable:"+ds.Addressable.Id, id)
	for _, label := range ds.Labels {
		_ = conn.Send("SREM", db.DeviceService+":label:"+label, id)
	}
	_, err = conn.Do("EXEC")
	return err
}

//  ----------------------Provision Watcher -----------------------------*/
func (c *Client) GetAllProvisionWatchers() ([]contract.ProvisionWatcher, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.ProvisionWatcher, 0, -1)
	if err != nil {
		return []contract.ProvisionWatcher{}, err
	}

	pw := make([]contract.ProvisionWatcher, len(objects))
	for i, object := range objects {
		err = unmarshalProvisionWatcher(object, &pw[i])
		if err != nil {
			return []contract.ProvisionWatcher{}, err
		}
	}

	return pw, nil
}

func (c *Client) GetProvisionWatcherByName(n string) (contract.ProvisionWatcher, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var pw contract.ProvisionWatcher
	err := getObjectByHash(conn, db.ProvisionWatcher+":name", n, unmarshalProvisionWatcher, &pw)
	return pw, err
}

func (c *Client) GetProvisionWatchersByIdentifier(k string, v string) (pw []contract.ProvisionWatcher, err error) {
	return c.getProvisionWatchersByValue(db.ProvisionWatcher + ":identifier:" + k + ":" + v)
}

func (c *Client) GetProvisionWatchersByServiceId(id string) ([]contract.ProvisionWatcher, error) {
	return c.getProvisionWatchersByValue(db.ProvisionWatcher + ":service:" + id)
}

func (c *Client) GetProvisionWatchersByProfileId(id string) ([]contract.ProvisionWatcher, error) {
	return c.getProvisionWatchersByValue(db.ProvisionWatcher + ":profile:" + id)
}

func (c *Client) GetProvisionWatcherById(id string) (contract.ProvisionWatcher, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var pw contract.ProvisionWatcher
	err := getObjectById(conn, id, unmarshalProvisionWatcher, &pw)
	return pw, err
}

func (c *Client) getProvisionWatchersByValue(v string) ([]contract.ProvisionWatcher, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, v)
	if err != nil {
		return []contract.ProvisionWatcher{}, err
	}

	pw := make([]contract.ProvisionWatcher, len(objects))
	for i, object := range objects {
		err = unmarshalProvisionWatcher(object, &pw[i])
		if err != nil {
			return []contract.ProvisionWatcher{}, err
		}
	}

	return pw, nil
}

func (c *Client) AddProvisionWatcher(pw contract.ProvisionWatcher) (string, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	return addProvisionWatcher(conn, pw)
}

func (c *Client) UpdateProvisionWatcher(pw contract.ProvisionWatcher) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteProvisionWatcher(conn, pw.Id)
	if err != nil {
		return err
	}

	_, err = addProvisionWatcher(conn, pw)
	return err
}

func (c *Client) DeleteProvisionWatcherById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteProvisionWatcher(conn, id)
}

func addProvisionWatcher(conn redis.Conn, pw contract.ProvisionWatcher) (string, error) {
	pid := pw.Profile.Id
	sid := pw.Service.Id
	_ = conn.Send("MULTI")
	_ = conn.Send("HEXISTS", db.ProvisionWatcher+":name", pw.Name)
	_ = conn.Send("EXISTS", pid)
	_ = conn.Send("HGET", db.DeviceProfile+":name", pw.Profile.Name)
	_ = conn.Send("EXISTS", sid)
	_ = conn.Send("HGET", db.DeviceService+":name", pw.Service.Name)
	rep, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return "", err
	}

	// Verify uniquness of name
	nex, err := redis.Bool(rep[0], nil)
	if err != nil {
		return "", err
	} else if nex {
		return "", db.ErrNotUnique
	}

	// Verify existence of a device profile by ID or name
	idex, err := redis.Bool(rep[1], nil)
	if err != nil {
		return "", err
	} else if !idex {
		pid, err = redis.String(rep[2], nil)
		if err == redis.ErrNil {
			return "", errors.New("Invalid Device Profile")
		} else if err != nil {
			return "", err
		}
		pw.Profile.Id = pid
	}

	// Verify existence of a device profile by ID or name
	idex, err = redis.Bool(rep[3], nil)
	if err != nil {
		return "", err
	} else if !idex {
		sid, err = redis.String(rep[4], nil)
		if err == redis.ErrNil {
			return "", errors.New("Invalid Device Service")
		} else if err != nil {
			return "", err
		}
		pw.Service.Id = sid
	}

	_, err = uuid.Parse(pw.Id)
	if err != nil {
		pw.Id = uuid.New().String()
	}
	id := pw.Id

	ts := db.MakeTimestamp()
	if pw.Created == 0 {
		pw.Created = ts
	}
	pw.Modified = ts

	m, err := marshalProvisionWatcher(pw)
	if err != nil {
		return "", err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", id, m)
	_ = conn.Send("ZADD", db.ProvisionWatcher, 0, id)
	_ = conn.Send("HSET", db.ProvisionWatcher+":name", pw.Name, id)
	_ = conn.Send("SADD", db.ProvisionWatcher+":service:"+pw.Service.Id, id)
	_ = conn.Send("SADD", db.ProvisionWatcher+":profile:"+pw.Profile.Id, id)
	for k, v := range pw.Identifiers {
		_ = conn.Send("SADD", db.ProvisionWatcher+":identifier:"+k+":"+v, id)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		return "", err
	}

	return id, nil
}

func deleteProvisionWatcher(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	pw := contract.ProvisionWatcher{}
	_ = unmarshalProvisionWatcher(object, &pw)

	_ = conn.Send("MULTI")
	_ = conn.Send("DEL", id)
	_ = conn.Send("ZREM", db.ProvisionWatcher, id)
	_ = conn.Send("HDEL", db.ProvisionWatcher+":name", pw.Name)
	_ = conn.Send("SREM", db.ProvisionWatcher+":service:"+pw.Service.Id, id)
	_ = conn.Send("SREM", db.ProvisionWatcher+":profile:"+pw.Profile.Id, id)
	for k, v := range pw.Identifiers {
		_ = conn.Send("SREM", db.ProvisionWatcher+":identifier:"+k+":"+v, id)
	}
	_, err = conn.Do("EXEC")
	return err
}

//  ------------------------Command -------------------------------------*/
func (c *Client) GetAllCommands() ([]contract.Command, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Command, 0, -1)
	if err != nil {
		return []contract.Command{}, err
	}

	d := make([]contract.Command, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &d[i])
		if err != nil {
			return []contract.Command{}, err
		}
	}

	return d, nil
}

func (c *Client) GetCommandById(id string) (contract.Command, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var d contract.Command
	err := getObjectById(conn, id, unmarshalObject, &d)
	return d, err
}

func (c *Client) GetCommandByName(n string) ([]contract.Command, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, db.Command+":name:"+n)
	if err != nil {
		return []contract.Command{}, err
	}

	cmd := make([]contract.Command, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &cmd[i])
		if err != nil {
			return []contract.Command{}, err
		}
	}

	return cmd, nil
}

func addCommand(conn redis.Conn, tx bool, cmd contract.Command) (string, error) {
	cmd.Id = uuid.New().String()
	ts := db.MakeTimestamp()
	if cmd.Created == 0 {
		cmd.Created = ts
	}
	cmd.Modified = ts

	m, err := marshalObject(cmd)
	if err != nil {
		return "", err
	}

	if tx {
		_ = conn.Send("MULTI")
	}
	_ = conn.Send("SET", cmd.Id, m)
	_ = conn.Send("ZADD", db.Command, 0, cmd.Id)
	_ = conn.Send("SADD", db.Command+":name:"+cmd.Name, cmd.Id)
	if tx {
		_, err = conn.Do("EXEC")
		if err != nil {
			return "", err
		}
	}

	return cmd.Id, nil
}
func (c *Client) GetCommandsByDeviceId(did string) ([]contract.Command, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	return getCommandsByDeviceId(conn, did)
}

func getCommandsByDeviceId(conn redis.Conn, id string) (commands []contract.Command, err error) {
	err = validateKeyExists(conn, id)
	if err != nil {
		if err == db.ErrNotFound {
			err = types.NewErrItemNotFound(fmt.Sprintf("device with id %s not found", id))
		}
		return []contract.Command{}, err
	}

	objects, err := getObjectsByValue(conn, db.Command+":device:"+id)
	if err != nil {
		return []contract.Command{}, err
	}

	commands = make([]contract.Command, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &commands[i])
		if err != nil {
			return []contract.Command{}, err
		}
	}
	return commands, nil
}

// GetCommandsByDaviceName coming soon

func deleteCommand(conn redis.Conn, cmd contract.Command) {
	_ = conn.Send("DEL", cmd.Id)
	_ = conn.Send("ZREM", db.Command, cmd.Id)
	_ = conn.Send("SREM", db.Command+":name:"+cmd.Name, cmd.Id)
}

func (c *Client) ScrubMetadata() (err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	cols := []string{
		db.Addressable, db.Command, db.DeviceService, db.DeviceReport, db.DeviceProfile,
		db.Device, db.ProvisionWatcher,
	}

	for _, col := range cols {
		err = unlinkCollection(conn, col)
		if err != nil {
			return err
		}
	}

	return nil
}
