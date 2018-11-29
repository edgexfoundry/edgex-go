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
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/mgo.v2/bson"
)

/* -----------------------Schedule Event ------------------------*/
func (c *Client) UpdateScheduleEvent(se models.ScheduleEvent) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteScheduleEvent(conn, se.Id.Hex())
	if err != nil {
		return err
	}

	return addScheduleEvent(conn, &se)
}

func (c *Client) AddScheduleEvent(se *models.ScheduleEvent) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return addScheduleEvent(conn, se)
}

func (c *Client) GetAllScheduleEvents(se *[]models.ScheduleEvent) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.ScheduleEvent, 0, -1)
	if err != nil {
		return err
	}

	*se = make([]models.ScheduleEvent, len(objects))
	for i, object := range objects {
		err = unmarshalScheduleEvent(object, &(*se)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetScheduleEventByName(se *models.ScheduleEvent, n string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectByHash(conn, db.ScheduleEvent+":name", n, unmarshalScheduleEvent, se)
}

func (c *Client) GetScheduleEventById(se *models.ScheduleEvent, id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectById(conn, id, unmarshalScheduleEvent, se)
}

func (c *Client) GetScheduleEventsByScheduleName(se *[]models.ScheduleEvent, n string) error {
	return c.getScheduleEventsByValue(se, db.ScheduleEvent+":schedule:"+n)
}

func (c *Client) GetScheduleEventsByAddressableId(se *[]models.ScheduleEvent, id string) error {
	return c.getScheduleEventsByValue(se, db.ScheduleEvent+":addressable:"+id)
}

func (c *Client) GetScheduleEventsByServiceName(se *[]models.ScheduleEvent, n string) error {
	return c.getScheduleEventsByValue(se, db.ScheduleEvent+":service:"+n)
}

func (c *Client) DeleteScheduleEventById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteScheduleEvent(conn, id)
}

func (c *Client) getScheduleEventsByValue(se *[]models.ScheduleEvent, v string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, v)
	if err != nil {
		return err
	}

	*se = make([]models.ScheduleEvent, len(objects))
	for i, object := range objects {
		err = unmarshalScheduleEvent(object, &(*se)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func addScheduleEvent(conn redis.Conn, se *models.ScheduleEvent) error {
	aid := se.Addressable.Id.Hex()
	conn.Send("MULTI")
	conn.Send("HEXISTS", db.ScheduleEvent+":name", se.Name)
	conn.Send("EXISTS", aid)
	conn.Send("HGET", db.Addressable+":name", se.Addressable.Name)
	rep, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return err
	}

	// Verify uniquness of name
	nex, err := redis.Bool(rep[0], nil)
	if err != nil {
		return err
	} else if nex {
		return db.ErrNotUnique
	}

	// Verify existence of an addressable by ID or name
	idex, err := redis.Bool(rep[1], nil)
	if err != nil {
		return err
	} else if !idex {
		aid, err = redis.String(rep[2], nil)
		if err == redis.ErrNil {
			return errors.New("Invalid addressable")
		} else if err != nil {
			return err
		}
		se.Addressable.Id = bson.ObjectIdHex(aid)
	}

	if !se.Id.Valid() {
		se.Id = bson.NewObjectId()
	}
	id := se.Id.Hex()

	ts := db.MakeTimestamp()
	if se.Created == 0 {
		se.Created = ts
	}
	se.Modified = ts

	m, err := marshalScheduleEvent(*se)
	if err != nil {
		return err
	}

	conn.Send("MULTI")
	conn.Send("SET", id, m)
	conn.Send("ZADD", db.ScheduleEvent, 0, id)
	conn.Send("HSET", db.ScheduleEvent+":name", se.Name, id)
	conn.Send("SADD", db.ScheduleEvent+":addressable:"+aid, id)
	conn.Send("SADD", db.ScheduleEvent+":schedule:"+se.Schedule, id)
	conn.Send("SADD", db.ScheduleEvent+":service:"+se.Service, id)
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

func deleteScheduleEvent(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	se := models.ScheduleEvent{}
	err = unmarshalScheduleEvent(object, &se)

	conn.Send("MULTI")
	conn.Send("DEL", id)
	conn.Send("ZREM", db.ScheduleEvent, id)
	conn.Send("HDEL", db.ScheduleEvent+":name", se.Name)
	conn.Send("SREM", db.ScheduleEvent+":addressable:"+se.Addressable.Id.Hex(), id)
	conn.Send("SREM", db.ScheduleEvent+":schedule:"+se.Schedule, id)
	conn.Send("SREM", db.ScheduleEvent+":service:"+se.Service, id)
	_, err = conn.Do("EXEC")
	return err
}

//  --------------------------Schedule ---------------------------*/
func (c *Client) GetAllSchedules(s *[]models.Schedule) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Schedule, 0, -1)
	if err != nil {
		return err
	}

	*s = make([]models.Schedule, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &(*s)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetScheduleByName(s *models.Schedule, n string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectByHash(conn, db.Schedule+":name", n, unmarshalObject, s)
}

func (c *Client) GetScheduleById(s *models.Schedule, id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectById(conn, id, unmarshalObject, s)
}

func (c *Client) AddSchedule(sch *models.Schedule) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return addSchedule(conn, sch)
}

func (c *Client) UpdateSchedule(sch models.Schedule) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteSchedule(conn, sch.Id.Hex())
	if err != nil {
		return err
	}

	return addSchedule(conn, &sch)
}

func (c *Client) DeleteScheduleById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteSchedule(conn, id)
}

func addSchedule(conn redis.Conn, sch *models.Schedule) error {
	exists, err := redis.Bool(conn.Do("HEXISTS", db.Schedule+":name", sch.Name))
	if err != nil {
		return err
	} else if exists {
		return db.ErrNotUnique
	}

	if !sch.Id.Valid() {
		sch.Id = bson.NewObjectId()
	}
	id := sch.Id.Hex()

	ts := db.MakeTimestamp()
	if sch.Created == 0 {
		sch.Created = ts
	}
	sch.Modified = ts

	m, err := marshalObject(sch)
	if err != nil {
		return err
	}

	conn.Send("MULTI")
	conn.Send("SET", id, m)
	conn.Send("ZADD", db.Schedule, 0, id)
	conn.Send("HSET", db.Schedule+":name", sch.Name, id)
	_, err = conn.Do("EXEC")
	return err
}

func deleteSchedule(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	s := models.Schedule{}
	err = unmarshalObject(object, &s)

	conn.Send("MULTI")
	conn.Send("DEL", id)
	conn.Send("ZREM", db.Schedule, id)
	conn.Send("HDEL", db.Schedule+":name", s.Name)
	_, err = conn.Do("EXEC")
	return err
}

// /* ----------------------Device Report --------------------------*/
func (c *Client) GetAllDeviceReports(d *[]models.DeviceReport) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.DeviceReport, 0, -1)
	if err != nil {
		return err
	}

	*d = make([]models.DeviceReport, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &(*d)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetDeviceReportByName(d *models.DeviceReport, n string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectByHash(conn, db.DeviceReport+":name", n, unmarshalObject, d)
}

func (c *Client) GetDeviceReportByDeviceName(d *[]models.DeviceReport, n string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, db.DeviceReport+":device:"+n)
	if err != nil {
		return err
	}

	*d = make([]models.DeviceReport, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &(*d)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetDeviceReportById(d *models.DeviceReport, id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectById(conn, id, unmarshalObject, d)
}

func (c *Client) GetDeviceReportsByScheduleEventName(d *[]models.DeviceReport, n string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, db.DeviceReport+":scheduleevent:"+n)
	if err != nil {
		return err
	}

	*d = make([]models.DeviceReport, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &(*d)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) AddDeviceReport(d *models.DeviceReport) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return addDeviceReport(conn, d)
}

func (c *Client) UpdateDeviceReport(dr *models.DeviceReport) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteDeviceReport(conn, dr.Id.Hex())
	if err != nil {
		return err
	}

	return addDeviceReport(conn, dr)
}

func (c *Client) DeleteDeviceReportById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteDeviceReport(conn, id)
}

func addDeviceReport(conn redis.Conn, dr *models.DeviceReport) error {
	exists, err := redis.Bool(conn.Do("HEXISTS", db.DeviceReport+":name", dr.Name))
	if err != nil {
		return err
	} else if exists {
		return db.ErrNotUnique
	}

	if !dr.Id.Valid() {
		dr.Id = bson.NewObjectId()
	}
	id := dr.Id.Hex()

	ts := db.MakeTimestamp()
	if dr.Created == 0 {
		dr.Created = ts
	}
	dr.Modified = ts

	m, err := marshalObject(dr)
	if err != nil {
		return err
	}

	conn.Send("MULTI")
	conn.Send("SET", id, m)
	conn.Send("ZADD", db.DeviceReport, 0, id)
	conn.Send("SADD", db.DeviceReport+":device:"+dr.Device, id)
	conn.Send("SADD", db.DeviceReport+":scheduleevent:"+dr.Event, id)
	conn.Send("HSET", db.DeviceReport+":name", dr.Name, id)
	_, err = conn.Do("EXEC")
	return err
}

func deleteDeviceReport(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	dr := models.DeviceReport{}
	err = unmarshalObject(object, &dr)

	conn.Send("MULTI")
	conn.Send("DEL", id)
	conn.Send("ZREM", db.DeviceReport, id)
	conn.Send("SREM", db.DeviceReport+":device:"+dr.Device, id)
	conn.Send("SREM", db.DeviceReport+":scheduleevent:"+dr.Event, id)
	conn.Send("HDEL", db.DeviceReport+":name", dr.Name)
	_, err = conn.Do("EXEC")
	return err
}

// /* ----------------------------- Device ---------------------------------- */
func (c *Client) AddDevice(d *models.Device) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return addDevice(conn, d)
}

func (c *Client) UpdateDevice(d models.Device) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteDevice(conn, d.Id.Hex())
	if err != nil {
		return err
	}

	return addDevice(conn, &d)
}

func (c *Client) DeleteDeviceById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteDevice(conn, id)
}

func (c *Client) GetAllDevices(d *[]models.Device) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Device, 0, -1)
	if err != nil {
		return err
	}

	*d = make([]models.Device, len(objects))
	for i, object := range objects {
		err = unmarshalDevice(object, &(*d)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetDevicesByProfileId(d *[]models.Device, pid string) error {
	return c.getDevicesByValue(d, db.Device+":profile:"+pid)
}

func (c *Client) GetDeviceById(d *models.Device, id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectById(conn, id, unmarshalDevice, d)
}

func (c *Client) GetDeviceByName(d *models.Device, n string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectByHash(conn, db.Device+":name", n, unmarshalDevice, d)
}

func (c *Client) GetDevicesByServiceId(d *[]models.Device, sid string) error {
	return c.getDevicesByValue(d, db.Device+":service:"+sid)
}

func (c *Client) GetDevicesByAddressableId(d *[]models.Device, aid string) error {
	return c.getDevicesByValue(d, db.Device+":addressable:"+aid)
}

func (c *Client) GetDevicesWithLabel(d *[]models.Device, l string) error {
	return c.getDevicesByValue(d, db.Device+":label:"+l)
}

func (c *Client) getDevicesByValue(d *[]models.Device, v string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, v)
	if err != nil {
		return err
	}

	*d = make([]models.Device, len(objects))
	for i, object := range objects {
		err = unmarshalDevice(object, &(*d)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func addDevice(conn redis.Conn, d *models.Device) error {
	aid := d.Addressable.Id.Hex()
	conn.Send("MULTI")
	conn.Send("HEXISTS", db.Device+":name", d.Name)
	conn.Send("EXISTS", aid)
	conn.Send("HGET", db.Addressable+":name", d.Addressable.Name)
	rep, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return err
	}

	// Verify uniquness of name
	nex, err := redis.Bool(rep[0], nil)
	if err != nil {
		return err
	} else if nex {
		return db.ErrNotUnique
	}

	// Verify existence of an addressable by ID or name
	idex, err := redis.Bool(rep[1], nil)
	if err != nil {
		return err
	} else if !idex {
		aid, err = redis.String(rep[2], nil)
		if err == redis.ErrNil {
			return errors.New("Invalid addressable")
		} else if err != nil {
			return err
		}
		d.Addressable.Id = bson.ObjectIdHex(aid)
	}

	if !d.Id.Valid() {
		d.Id = bson.NewObjectId()
	}
	id := d.Id.Hex()

	ts := db.MakeTimestamp()
	if d.Created == 0 {
		d.Created = ts
	}
	d.Modified = ts

	m, err := marshalDevice(*d)
	if err != nil {
		return err
	}

	conn.Send("MULTI")
	conn.Send("SET", id, m)
	conn.Send("ZADD", db.Device, 0, id)
	conn.Send("HSET", db.Device+":name", d.Name, id)
	conn.Send("SADD", db.Device+":addressable:"+d.Addressable.Id.Hex(), id)
	conn.Send("SADD", db.Device+":service:"+d.Service.Id.Hex(), id)
	conn.Send("SADD", db.Device+":profile:"+d.Profile.Id.Hex(), id)
	for _, label := range d.Labels {
		conn.Send("SADD", db.Device+":label:"+label, id)
	}
	_, err = conn.Do("EXEC")
	return err
}

func deleteDevice(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	d := models.Device{}
	err = unmarshalDevice(object, &d)

	conn.Send("MULTI")
	conn.Send("DEL", id)
	conn.Send("ZREM", db.Device, id)
	conn.Send("HDEL", db.Device+":name", d.Name)
	conn.Send("SREM", db.Device+":addressable:"+d.Addressable.Id.Hex(), id)
	conn.Send("SREM", db.Device+":service:"+d.Service.Id.Hex(), id)
	conn.Send("SREM", db.Device+":profile:"+d.Profile.Id.Hex(), id)
	for _, label := range d.Labels {
		conn.Send("SREM", db.Device+":label:"+label, id)
	}
	_, err = conn.Do("EXEC")
	return err
}

// /* -----------------------------Device Profile -----------------------------*/
func (c *Client) GetDeviceProfileById(d *models.DeviceProfile, id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectById(conn, id, unmarshalDeviceProfile, d)
}

func (c *Client) GetAllDeviceProfiles(dp *[]models.DeviceProfile) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.DeviceProfile, 0, -1)
	if err != nil {
		return err
	}

	*dp = make([]models.DeviceProfile, len(objects))
	for i, object := range objects {
		err = unmarshalDeviceProfile(object, &(*dp)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetDeviceProfilesByModel(dp *[]models.DeviceProfile, model string) error {
	return c.getDeviceProfilesByValues(dp, db.DeviceProfile+":model:"+model)
}

func (c *Client) GetDeviceProfilesWithLabel(dp *[]models.DeviceProfile, l string) error {
	return c.getDeviceProfilesByValues(dp, db.DeviceProfile+":label:"+l)
}

func (c *Client) GetDeviceProfilesByManufacturerModel(dp *[]models.DeviceProfile, man string, mod string) error {
	return c.getDeviceProfilesByValues(dp, db.DeviceProfile+":manufacturer:"+man, db.DeviceProfile+":model:"+mod)
}

func (c *Client) GetDeviceProfilesByManufacturer(dp *[]models.DeviceProfile, man string) error {
	return c.getDeviceProfilesByValues(dp, db.DeviceProfile+":manufacturer:"+man)
}

func (c *Client) GetDeviceProfileByName(dp *models.DeviceProfile, n string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectByHash(conn, db.DeviceProfile+":name", n, unmarshalDeviceProfile, dp)
}

// Get the device profiles that are currently using the command
func (c *Client) GetDeviceProfilesUsingCommand(dp *[]models.DeviceProfile, cmd models.Command) error {
	return c.getDeviceProfilesByValues(dp, db.DeviceProfile+":command:"+cmd.Id.Hex())
}

// Get device profiles with the passed query
func (c *Client) getDeviceProfilesByValues(dp *[]models.DeviceProfile, vals ...string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValues(conn, vals...)
	if err != nil {
		return err
	}

	*dp = make([]models.DeviceProfile, len(objects))
	for i, object := range objects {
		err = unmarshalDeviceProfile(object, &(*dp)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) AddDeviceProfile(dp *models.DeviceProfile) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return addDeviceProfile(conn, dp)
}

func (c *Client) UpdateDeviceProfile(dp *models.DeviceProfile) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteDeviceProfile(conn, dp.Id.Hex())
	if err != nil {
		return err
	}

	return addDeviceProfile(conn, dp)
}

func (c *Client) DeleteDeviceProfileById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteDeviceProfile(conn, id)
}

func addDeviceProfile(conn redis.Conn, dp *models.DeviceProfile) error {
	exists, err := redis.Bool(conn.Do("HEXISTS", db.DeviceProfile+":name", dp.Name))
	if err != nil {
		return err
	} else if exists {
		return db.ErrNotUnique
	}

	if !dp.Id.Valid() {
		dp.Id = bson.NewObjectId()
	}
	id := dp.Id.Hex()

	ts := db.MakeTimestamp()
	if dp.Created == 0 {
		dp.Created = ts
	}
	dp.Modified = ts

	m, err := marshalDeviceProfile(*dp)
	if err != nil {
		return err
	}

	conn.Send("MULTI")
	conn.Send("SET", id, m)
	conn.Send("ZADD", db.DeviceProfile, 0, id)
	conn.Send("HSET", db.DeviceProfile+":name", dp.Name, id)
	conn.Send("SADD", db.DeviceProfile+":manufacturer:"+dp.Manufacturer, id)
	conn.Send("SADD", db.DeviceProfile+":model:"+dp.Model, id)
	for _, label := range dp.Labels {
		conn.Send("SADD", db.DeviceProfile+":label:"+label, id)
	}
	if len(dp.Commands) > 0 {
		cids := redis.Args{}.Add(db.DeviceProfile + ":commands:" + id)
		for _, c := range dp.Commands {
			err = addCommand(conn, false, &c)
			if err != nil {
				return err
			}
			cid := c.Id.Hex()
			conn.Send("SADD", db.DeviceProfile+":command:"+cid, id)
			cids = cids.Add(cid)
		}
		conn.Send("SADD", cids...)
	}
	_, err = conn.Do("EXEC")
	return err
}

func deleteDeviceProfile(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	dp := models.DeviceProfile{}
	err = unmarshalObject(object, &dp)

	conn.Send("MULTI")
	conn.Send("DEL", id)
	conn.Send("ZREM", db.DeviceProfile, id)
	conn.Send("HDEL", db.DeviceProfile+":name", dp.Name)
	conn.Send("SREM", db.DeviceProfile+":manufacturer:"+dp.Manufacturer, id)
	conn.Send("SREM", db.DeviceProfile+":model:"+dp.Model, id)
	for _, label := range dp.Labels {
		conn.Send("SREM", db.DeviceProfile+":label:"+label, id)
	}
	// TODO: should commands be also removed?
	for _, c := range dp.Commands {
		conn.Send("SREM", db.DeviceProfile+":command:"+c.Id.Hex(), id)
	}
	conn.Send("DEL", db.DeviceProfile+":commands:"+id)
	_, err = conn.Do("EXEC")
	return err
}

/* -----------------------------------Addressable -------------------------- */
func (c *Client) UpdateAddressable(updated *models.Addressable, orig *models.Addressable) error {
	conn := c.Pool.Get()
	defer conn.Close()

	if updated == nil {
		return nil
	}

	err := deleteAddressable(conn, orig.Id.Hex())
	if err != nil {
		return err
	}

	if updated.Name != "" {
		orig.Name = updated.Name
	}
	if updated.Protocol != "" {
		orig.Protocol = updated.Protocol
	}
	if updated.Address != "" {
		orig.Address = updated.Address
	}
	if updated.Port != int(0) {
		orig.Port = updated.Port
	}
	if updated.Path != "" {
		orig.Path = updated.Path
	}
	if updated.Publisher != "" {
		orig.Publisher = updated.Publisher
	}
	if updated.User != "" {
		orig.User = updated.User
	}
	if updated.Password != "" {
		orig.Password = updated.Password
	}
	if updated.Topic != "" {
		orig.Topic = updated.Topic
	}

	_, err = addAddressable(conn, orig)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) AddAddressable(a *models.Addressable) (bson.ObjectId, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	return addAddressable(conn, a)
}

func (c *Client) GetAddressableById(a *models.Addressable, id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectById(conn, id, unmarshalObject, a)
}

func (c *Client) GetAddressableByName(a *models.Addressable, n string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectByHash(conn, db.Addressable+":name", n, unmarshalObject, a)
}

func (c *Client) GetAddressablesByTopic(a *[]models.Addressable, t string) error {
	return c.getAddressablesByValue(a, db.Addressable+":topic:"+t)
}

func (c *Client) GetAddressablesByPort(a *[]models.Addressable, p int) error {
	return c.getAddressablesByValue(a, db.Addressable+":port:"+strconv.Itoa(p))
}

func (c *Client) GetAddressablesByPublisher(a *[]models.Addressable, p string) error {
	return c.getAddressablesByValue(a, db.Addressable+":publisher:"+p)
}

func (c *Client) GetAddressablesByAddress(a *[]models.Addressable, add string) error {
	return c.getAddressablesByValue(a, db.Addressable+":address:"+add)
}

func (c *Client) GetAddressables(d *[]models.Addressable) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Addressable, 0, -1)
	if err != nil {
		return err
	}

	*d = make([]models.Addressable, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &(*d)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DeleteAddressableById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteAddressable(conn, id)
}

func (c *Client) getAddressablesByValue(a *[]models.Addressable, v string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, v)
	if err != nil {
		return err
	}

	*a = make([]models.Addressable, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &(*a)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func addAddressable(conn redis.Conn, a *models.Addressable) (bson.ObjectId, error) {
	exists, err := redis.Bool(conn.Do("HEXISTS", db.Addressable+":name", a.Name))
	if err != nil {
		return a.Id, err
	} else if exists {
		return a.Id, db.ErrNotUnique
	}

	if !a.Id.Valid() {
		a.Id = bson.NewObjectId()
	}
	id := a.Id.Hex()

	ts := db.MakeTimestamp()
	if a.Created == 0 {
		a.Created = ts
	}
	a.Modified = ts

	m, err := marshalObject(a)
	if err != nil {
		return a.Id, err
	}

	conn.Send("MULTI")
	conn.Send("SET", id, m)
	conn.Send("ZADD", db.Addressable, 0, id)
	conn.Send("SADD", db.Addressable+":topic:"+a.Topic, id)
	conn.Send("SADD", db.Addressable+":port:"+strconv.Itoa(a.Port), id)
	conn.Send("SADD", db.Addressable+":publisher:"+a.Publisher, id)
	conn.Send("SADD", db.Addressable+":address:"+a.Address, id)
	conn.Send("HSET", db.Addressable+":name", a.Name, id)
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

	a := models.Addressable{}
	err = unmarshalObject(object, &a)

	conn.Send("MULTI")
	conn.Send("DEL", id)
	conn.Send("ZREM", db.Addressable, id)
	conn.Send("SREM", db.Addressable+":topic:"+a.Topic, id)
	conn.Send("SREM", db.Addressable+":port:"+strconv.Itoa(a.Port), id)
	conn.Send("SREM", db.Addressable+":publisher:"+a.Publisher, id)
	conn.Send("SREM", db.Addressable+":address:"+a.Address, id)
	conn.Send("HDEL", db.Addressable+":name", a.Name)
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

// /* ----------------------------- Device Service ----------------------------------*/
func (c *Client) GetDeviceServiceByName(d *models.DeviceService, n string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectByHash(conn, db.DeviceService+":name", n, unmarshalDeviceService, d)
}

func (c *Client) GetDeviceServiceById(d *models.DeviceService, id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectById(conn, id, unmarshalDeviceService, d)
}

func (c *Client) GetAllDeviceServices(d *[]models.DeviceService) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.DeviceService, 0, -1)
	if err != nil {
		return err
	}

	*d = make([]models.DeviceService, len(objects))
	for i, object := range objects {
		err = unmarshalDeviceService(object, &(*d)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetDeviceServicesByAddressableId(d *[]models.DeviceService, id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, db.DeviceService+":addressable:"+id)
	if err != nil {
		return err
	}

	*d = make([]models.DeviceService, len(objects))
	for i, object := range objects {
		err = unmarshalDeviceService(object, &(*d)[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) GetDeviceServicesWithLabel(d *[]models.DeviceService, l string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, db.DeviceService+":label:"+l)
	if err != nil {
		return err
	}

	*d = make([]models.DeviceService, len(objects))
	for i, object := range objects {
		err = unmarshalDeviceService(object, &(*d)[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) AddDeviceService(d *models.DeviceService) error {
	conn := c.Pool.Get()
	defer conn.Close()

	_, err := addDeviceService(conn, d)
	return err
}

func (c *Client) UpdateDeviceService(deviceService models.DeviceService) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteDeviceService(conn, deviceService.Id.Hex())
	if err != nil {
		return err
	}

	deviceService.Modified = db.MakeTimestamp()

	_, err = addDeviceService(conn, &deviceService)

	return err
}

func (c *Client) DeleteDeviceServiceById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteDeviceService(conn, id)
}

func addDeviceService(conn redis.Conn, ds *models.DeviceService) (bson.ObjectId, error) {
	aid := ds.Addressable.Id.Hex()
	conn.Send("MULTI")
	conn.Send("HEXISTS", db.DeviceService+":name", ds.Name)
	conn.Send("EXISTS", aid)
	conn.Send("HGET", db.Addressable+":name", ds.Addressable.Name)
	rep, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return ds.Id, err
	}

	// Verify uniquness of name
	nex, err := redis.Bool(rep[0], nil)
	if err != nil {
		return ds.Id, err
	} else if nex {
		return ds.Id, db.ErrNotUnique
	}

	// Verify existence of an addressable by ID or name
	idex, err := redis.Bool(rep[1], nil)
	if err != nil {
		return ds.Id, err
	} else if !idex {
		aid, err = redis.String(rep[2], nil)
		if err == redis.ErrNil {
			return ds.Id, errors.New("Invalid addressable")
		} else if err != nil {
			return ds.Id, err
		}
		ds.Addressable.Id = bson.ObjectIdHex(aid)
	}

	if !ds.Id.Valid() {
		ds.Id = bson.NewObjectId()
	}
	id := ds.Id.Hex()

	ts := db.MakeTimestamp()
	if ds.Created == 0 {
		ds.Created = ts
	}
	ds.Modified = ts

	m, err := marshalDeviceService(*ds)
	if err != nil {
		return ds.Id, err
	}

	conn.Send("MULTI")
	conn.Send("SET", id, m)
	conn.Send("ZADD", db.DeviceService, 0, id)
	conn.Send("HSET", db.DeviceService+":name", ds.Name, id)
	conn.Send("SADD", db.DeviceService+":addressable:"+aid, id)
	for _, label := range ds.Labels {
		conn.Send("SADD", db.DeviceService+":label:"+label, id)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		return ds.Id, err
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

	ds := models.DeviceService{}
	err = unmarshalDeviceService(object, &ds)

	conn.Send("MULTI")
	conn.Send("DEL", id)
	conn.Send("ZREM", db.DeviceService, id)
	conn.Send("HDEL", db.DeviceService+":name", ds.Name)
	conn.Send("SREM", db.DeviceService+":addressable:"+ds.Addressable.Id.Hex(), id)
	for _, label := range ds.Labels {
		conn.Send("SREM", db.DeviceService+":label:"+label, id)
	}
	_, err = conn.Do("EXEC")
	return err
}

//  ----------------------Provision Watcher -----------------------------*/
func (c *Client) GetAllProvisionWatchers(pw *[]models.ProvisionWatcher) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.ProvisionWatcher, 0, -1)
	if err != nil {
		return err
	}

	*pw = make([]models.ProvisionWatcher, len(objects))
	for i, object := range objects {
		err = unmarshalProvisionWatcher(object, &(*pw)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetProvisionWatcherByName(pw *models.ProvisionWatcher, n string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectByHash(conn, db.ProvisionWatcher+":name", n, unmarshalProvisionWatcher, pw)
}

func (c *Client) GetProvisionWatchersByIdentifier(pw *[]models.ProvisionWatcher, k string, v string) error {
	return c.getProvisionWatchersByValue(pw, db.ProvisionWatcher+":identifier:"+k+":"+v)
}

func (c *Client) GetProvisionWatchersByServiceId(pw *[]models.ProvisionWatcher, id string) error {
	return c.getProvisionWatchersByValue(pw, db.ProvisionWatcher+":service:"+id)
}

func (c *Client) GetProvisionWatchersByProfileId(pw *[]models.ProvisionWatcher, id string) error {
	return c.getProvisionWatchersByValue(pw, db.ProvisionWatcher+":profile:"+id)
}

func (c *Client) GetProvisionWatcherById(pw *models.ProvisionWatcher, id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectById(conn, id, unmarshalProvisionWatcher, pw)
}

func (c *Client) getProvisionWatchersByValue(pw *[]models.ProvisionWatcher, v string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, v)
	if err != nil {
		return err
	}

	*pw = make([]models.ProvisionWatcher, len(objects))
	for i, object := range objects {
		err = unmarshalProvisionWatcher(object, &(*pw)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) AddProvisionWatcher(pw *models.ProvisionWatcher) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return addProvisionWatcher(conn, pw)
}

func (c *Client) UpdateProvisionWatcher(pw models.ProvisionWatcher) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteProvisionWatcher(conn, pw.Id.Hex())
	if err != nil {
		return err
	}

	return addProvisionWatcher(conn, &pw)
}

func (c *Client) DeleteProvisionWatcherById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return deleteProvisionWatcher(conn, id)
}

func addProvisionWatcher(conn redis.Conn, pw *models.ProvisionWatcher) error {
	pid := pw.Profile.Id.Hex()
	sid := pw.Service.Id.Hex()
	conn.Send("MULTI")
	conn.Send("HEXISTS", db.ProvisionWatcher+":name", pw.Name)
	conn.Send("EXISTS", pid)
	conn.Send("HGET", db.DeviceProfile+":name", pw.Profile.Name)
	conn.Send("EXISTS", sid)
	conn.Send("HGET", db.DeviceService+":name", pw.Service.Name)
	rep, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return err
	}

	// Verify uniquness of name
	nex, err := redis.Bool(rep[0], nil)
	if err != nil {
		return err
	} else if nex {
		return db.ErrNotUnique
	}

	// Verify existence of a device profile by ID or name
	idex, err := redis.Bool(rep[1], nil)
	if err != nil {
		return err
	} else if !idex {
		pid, err = redis.String(rep[2], nil)
		if err == redis.ErrNil {
			return errors.New("Invalid Device Profile")
		} else if err != nil {
			return err
		}
		pw.Profile.Id = bson.ObjectIdHex(pid)
	}

	// Verify existence of a device profile by ID or name
	idex, err = redis.Bool(rep[3], nil)
	if err != nil {
		return err
	} else if !idex {
		sid, err = redis.String(rep[4], nil)
		if err == redis.ErrNil {
			return errors.New("Invalid Device Service")
		} else if err != nil {
			return err
		}
		pw.Service.Id = bson.ObjectIdHex(sid)
	}

	if !pw.Id.Valid() {
		pw.Id = bson.NewObjectId()
	}
	id := pw.Id.Hex()

	ts := db.MakeTimestamp()
	if pw.Created == 0 {
		pw.Created = ts
	}
	pw.Modified = ts

	m, err := marshalProvisionWatcher(*pw)
	if err != nil {
		return err
	}

	conn.Send("MULTI")
	conn.Send("SET", id, m)
	conn.Send("ZADD", db.ProvisionWatcher, 0, id)
	conn.Send("HSET", db.ProvisionWatcher+":name", pw.Name, id)
	conn.Send("SADD", db.ProvisionWatcher+":service:"+pw.Service.Id.Hex(), id)
	conn.Send("SADD", db.ProvisionWatcher+":profile:"+pw.Profile.Id.Hex(), id)
	for k, v := range pw.Identifiers {
		conn.Send("SADD", db.ProvisionWatcher+":identifier:"+k+":"+v, id)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

func deleteProvisionWatcher(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	pw := models.ProvisionWatcher{}
	err = unmarshalProvisionWatcher(object, &pw)

	conn.Send("MULTI")
	conn.Send("DEL", id)
	conn.Send("ZREM", db.ProvisionWatcher, id)
	conn.Send("HDEL", db.ProvisionWatcher+":name", pw.Name)
	conn.Send("SREM", db.ProvisionWatcher+":service:"+pw.Service.Id.Hex(), id)
	conn.Send("SREM", db.ProvisionWatcher+":profile:"+pw.Profile.Id.Hex(), id)
	for k, v := range pw.Identifiers {
		conn.Send("SREM", db.ProvisionWatcher+":identifier:"+k+":"+v, id)
	}
	_, err = conn.Do("EXEC")
	return err
}

//  ------------------------Command -------------------------------------*/
func (c *Client) GetAllCommands(d *[]models.Command) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Command, 0, -1)
	if err != nil {
		return err
	}

	*d = make([]models.Command, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &(*d)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetCommandById(d *models.Command, id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return getObjectById(conn, id, unmarshalObject, d)
}

func (c *Client) GetCommandByName(cmd *[]models.Command, n string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, db.Command+":name:"+n)
	if err != nil {
		return err
	}

	*cmd = make([]models.Command, len(objects))
	for i, object := range objects {
		err = unmarshalObject(object, &(*cmd)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) AddCommand(cmd *models.Command) error {
	conn := c.Pool.Get()
	defer conn.Close()

	return addCommand(conn, true, cmd)
}

// Update command uses the ID of the command for identification
func (c *Client) UpdateCommand(updated *models.Command, orig *models.Command) error {
	conn := c.Pool.Get()
	defer conn.Close()

	if updated == nil {
		return nil
	}

	err := deleteCommand(conn, orig.Id.Hex())
	if err != nil {
		return err
	}

	if updated.Name != "" {
		orig.Name = updated.Name
	}
	if updated.Get != nil && (updated.Get.String() != models.Get{}.String()) {
		orig.Get = updated.Get
	}
	if updated.Put != nil && (updated.Put.String() != models.Put{}.String()) {
		orig.Put = updated.Put
	}
	if updated.Origin != 0 {
		orig.Origin = updated.Origin
	}

	return addCommand(conn, true, orig)
}

// Delete the command by ID
func (c *Client) DeleteCommandById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	// TODO: ??? Check if the command is still in use by device profiles

	return deleteCommand(conn, id)
}

func addCommand(conn redis.Conn, tx bool, cmd *models.Command) error {
	if !cmd.Id.Valid() {
		cmd.Id = bson.NewObjectId()
	}
	id := cmd.Id.Hex()

	ts := db.MakeTimestamp()
	if cmd.Created == 0 {
		cmd.Created = ts
	}
	cmd.Modified = ts

	m, err := marshalObject(cmd)
	if err != nil {
		return err
	}

	if tx {
		conn.Send("MULTI")
	}
	conn.Send("SET", id, m)
	conn.Send("ZADD", db.Command, 0, id)
	conn.Send("SADD", db.Command+":name:"+cmd.Name, id)
	if tx {
		_, err = conn.Do("EXEC")
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteCommand(conn redis.Conn, id string) error {
	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return db.ErrNotFound
	} else if err != nil {
		return err
	}

	cmd := models.Command{}
	err = unmarshalObject(object, &cmd)

	conn.Send("MULTI")
	conn.Send("DEL", id)
	conn.Send("ZREM", db.Command, id)
	conn.Send("SREM", db.Command+":name:"+cmd.Name, id)
	_, err = conn.Do("EXEC")
	return err
}

func (c *Client) ScrubMetadata() (err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	cols := []string{
		db.Addressable, db.Command, db.DeviceService,
		db.Schedule, db.DeviceReport, db.DeviceProfile,
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
