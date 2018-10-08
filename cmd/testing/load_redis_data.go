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

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"gopkg.in/mgo.v2/bson"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/redis"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

type handler func(map[string]interface{})

func processEvent(jsonMap map[string]interface{}) {
	var err error

	// See go/src/github.com/edgexfoundry/edgex-go/internal/pkg/db/redis/event.go
	e := struct {
		ID       string
		Pushed   int64
		Device   string
		Created  int64
		Modified int64
		Origin   int64
	}{
		ID:       jsonMap["_id"].(map[string]interface{})["$oid"].(string),
		Pushed:   int64(jsonMap["pushed"].(float64)),
		Device:   jsonMap["device"].(string),
		Created:  int64(jsonMap["created"].(float64)),
		Modified: int64(jsonMap["modified"].(float64)),
		Origin:   int64(jsonMap["origin"].(float64)),
	}

	// See go/src/github.com/edgexfoundry/edgex-go/internal/pkg/db/redis/data.go:addEvent
	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(e)
	redisConn.Send("SET", e.ID, marshalled)
	redisConn.Send("ZADD", db.EventsCollection, 0, e.ID)
	redisConn.Send("ZADD", db.EventsCollection+":created", e.Created, e.ID)
	redisConn.Send("ZADD", db.EventsCollection+":pushed", e.Pushed, e.ID)
	redisConn.Send("ZADD", db.EventsCollection+":device:"+e.Device, e.Created, e.ID)

	if len(jsonMap["readings"].([]interface{})) > 0 {
		readingIds := make([]interface{}, len(jsonMap["readings"].([]interface{}))*2+1)
		readingIds[0] = db.EventsCollection + ":readings:" + e.ID

		for i, v := range jsonMap["readings"].([]interface{}) {
			readingIds[i*2+1] = 0
			value := v.(map[string]interface{})["$id"].(map[string]interface{})["$oid"].(string)
			readingIds[i*2+2] = value
			redisConn.Send("ZADD", db.ReadingsCollection, 0, value)
		}

		redisConn.Send("ZADD", readingIds...)
	}

	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func processReading(jsonMap map[string]interface{}) {
	var err error

	setId := jsonMap["_id"].(map[string]interface{})["$oid"].(string)
	r := models.Reading{
		Id:       bson.ObjectIdHex(setId),
		Pushed:   int64(jsonMap["pushed"].(float64)),
		Created:  int64(jsonMap["created"].(float64)),
		Origin:   int64(jsonMap["origin"].(float64)),
		Modified: int64(jsonMap["modified"].(float64)),
		Name:     jsonMap["name"].(string),
		Value:    jsonMap["value"].(string),
	}

	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(r)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", db.ReadingsCollection, 0, setId)
	redisConn.Send("ZADD", db.ReadingsCollection+":created", r.Created, setId)
	redisConn.Send("ZADD", db.ReadingsCollection+":device:"+r.Device, r.Created, setId)
	redisConn.Send("ZADD", db.ReadingsCollection+":name:"+r.Name, r.Created, setId)
	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func readOptionalStringArray(i interface{}) []string {
	if i == nil {
		return nil
	}

	a := make([]string, len(i.([]interface{})))
	for c, v := range i.([]interface{}) {
		a[c] = v.(string)
	}

	return a
}

func processValueDescriptors(jsonMap map[string]interface{}) {
	var err error

	setId := jsonMap["_id"].(map[string]interface{})["$oid"].(string)
	valueDesc := models.ValueDescriptor{
		Id:           bson.ObjectIdHex(setId),
		Created:      int64(jsonMap["created"].(float64)),
		Modified:     int64(jsonMap["modified"].(float64)),
		Origin:       int64(jsonMap["origin"].(float64)),
		Name:         jsonMap["name"].(string),
		Min:          jsonMap["min"].(string),
		Max:          jsonMap["max"].(string),
		DefaultValue: jsonMap["defaultValue"].(string),
		Type:         jsonMap["type"].(string),
		UomLabel:     jsonMap["uomLabel"].(string),
		Formatting:   jsonMap["formatting"].(string),
		Labels:       readOptionalStringArray(jsonMap["labels"]),
	}

	// valueDesc.Labels = make([]string, len(jsonMap["labels"].([]interface{})))
	// for i, v := range jsonMap["labels"].([]interface{}) {
	// 	valueDesc.Labels[i] = v.(string)
	// }

	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(valueDesc)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", db.ValueDescriptorCollection, 0, setId)
	redisConn.Send("HSET", db.ValueDescriptorCollection+":name", valueDesc.Name, setId)
	redisConn.Send("ZADD", db.ValueDescriptorCollection+":uomlabel:"+valueDesc.UomLabel, 0, setId)
	redisConn.Send("ZADD", db.ValueDescriptorCollection+":type:"+valueDesc.Type, 0, setId)
	for _, label := range valueDesc.Labels {
		redisConn.Send("ZADD", db.ValueDescriptorCollection+":label:"+label, 0, setId)
	}

	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func readOptionalBaseObject(i interface{}) models.BaseObject {
	if i == nil {
		return models.BaseObject{}
	}

	m := i.(map[string]interface{})
	return models.BaseObject{
		Created:  int64(m["created"].(float64)),
		Modified: int64(m["modified"].(float64)),
		Origin:   int64(m["origin"].(float64)),
	}
}

func readOptionalInt(i interface{}) int {
	if i == nil {
		return 0
	}

	return int(i.(float64))
}

func readOptionalAddressable(i interface{}) models.Addressable {
	if i == nil {
		return models.Addressable{}
	}

	m := i.(map[string]interface{})

	var a models.Addressable

	if m["$ref"] != nil {
		a = models.Addressable{
			Id: bson.ObjectIdHex(m["$id"].(map[string]interface{})["$oid"].(string)),
		}
	} else {
		a = models.Addressable{
			BaseObject: readOptionalBaseObject(m),
			Id:         bson.ObjectIdHex(m["_id"].(map[string]interface{})["$oid"].(string)),
			Name:       readOptionalString(m["name"]),
			Protocol:   readOptionalString(m["protocol"]),
			HTTPMethod: readOptionalString(m["method"]),
			Address:    readOptionalString(m["address"]),
			Port:       readOptionalInt(m["port"]),
			Path:       readOptionalString(m["path"]),
			Publisher:  readOptionalString(m["publisher"]),
			User:       readOptionalString(m["user"]),
			Password:   readOptionalString(m["password"]),
			Topic:      readOptionalString(m["topic"]),
		}
	}

	return a
}

func processAddressable(jsonMap map[string]interface{}) {
	var err error

	a := readOptionalAddressable(jsonMap)
	setId := a.Id.Hex()
	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(a)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", db.Addressable, 0, setId)
	redisConn.Send("SADD", db.Addressable+":topic:"+a.Topic, setId)
	redisConn.Send("SADD", db.Addressable+":port:"+strconv.Itoa(a.Port), setId)
	redisConn.Send("SADD", db.Addressable+":publisher:"+a.Publisher, setId)
	redisConn.Send("SADD", db.Addressable+":address:"+a.Address, setId)
	redisConn.Send("HSET", db.Addressable+":name", a.Name, setId)
	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func readOptionalString(i interface{}) string {
	if i == nil {
		return ""
	}

	return i.(string)
}

// XXX Assumes single entry that's in test data
func readOptionalResponses(i interface{}) []models.Response {
	if i == nil {
		return []models.Response{}
	}

	m := i.(map[string]interface{})
	return []models.Response{
		models.Response{
			Code:           readOptionalString(m["code"]),
			Description:    m["errorDescription"].(string),         // XXX sample data is not aligned
			ExpectedValues: []string{m["expectedValues"].(string)}, // XXX sample data is not aligned
		},
	}
}

func readAction(m map[string]interface{}) models.Action {
	return models.Action{
		Path:      readOptionalString(m["path"]),
		URL:       readOptionalString(m["url"]),
		Responses: readOptionalResponses(m["response"]),
	}
}

func readOptionalGet(m map[string]interface{}) *models.Get {
	if m == nil {
		return nil
	}

	return &models.Get{
		Action: readAction(m),
	}
}

func readOptionalParameterNames(i interface{}) []string {
	if i == nil {
		return nil
	}

	a := i.([]interface{})
	names := make([]string, len(a))
	for i, v := range a {
		names[i] = v.(map[string]interface{})["name"].(string)
	}

	return names
}

func readOptionalPut(m map[string]interface{}) *models.Put {
	if m == nil {
		return nil
	}

	return &models.Put{
		Action:         readAction(m),
		ParameterNames: readOptionalParameterNames(m["parameters"]), // XXX sample data is not aligned
	}
}

func processCommand(jsonMap map[string]interface{}) {
	var err error

	setId := jsonMap["_id"].(map[string]interface{})["$oid"].(string)
	c := models.Command{
		BaseObject: readOptionalBaseObject(jsonMap),
		Id:         bson.ObjectIdHex(setId),
		Name:       readOptionalString(jsonMap["name"]),
		Get:        readOptionalGet(jsonMap["get"].(map[string]interface{})),
		Put:        readOptionalPut(jsonMap["put"].(map[string]interface{})),
	}

	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(c)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", db.Command, 0, setId)
	redisConn.Send("SADD", db.Command+":name:"+c.Name, setId)
	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func readOptionalDeviceService(i interface{}) models.DeviceService {
	if i == nil {
		return models.DeviceService{}
	}

	return models.DeviceService{
		Service: models.Service{
			Id: bson.ObjectIdHex(i.(map[string]interface{})["$id"].(map[string]interface{})["$oid"].(string)),
		},
	}
}

func readOptionalDeviceProfile(i interface{}) models.DeviceProfile {
	if i == nil {
		return models.DeviceProfile{}
	}

	return models.DeviceProfile{
		Id: bson.ObjectIdHex(i.(map[string]interface{})["$id"].(map[string]interface{})["$oid"].(string)),
	}
}

func readOptionalDescribedObject(i interface{}) models.DescribedObject {
	if i == nil {
		return models.DescribedObject{}
	}

	m := i.(map[string]interface{})
	return models.DescribedObject{
		BaseObject:  readOptionalBaseObject(m),
		Description: m["description"].(string),
	}
}

func processDevice(jsonMap map[string]interface{}) {
	var err error

	setId := jsonMap["_id"].(map[string]interface{})["$oid"].(string)
	d := models.Device{
		DescribedObject: readOptionalDescribedObject(jsonMap),
		Id:              bson.ObjectIdHex(setId),
		Name:            jsonMap["name"].(string),
		AdminState:      models.AdminState(jsonMap["adminState"].(string)),
		OperatingState:  models.OperatingState(jsonMap["operatingState"].(string)),
		LastConnected:   int64(jsonMap["lastConnected"].(float64)),
		LastReported:    int64(jsonMap["lastReported"].(float64)),
		Addressable:     readOptionalAddressable(jsonMap["addressable"]),
		Service:         readOptionalDeviceService(jsonMap["service"]),
		Profile:         readOptionalDeviceProfile(jsonMap["profile"]),
		Labels:          readOptionalStringArray(jsonMap["labels"]),
	}

	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(d)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", db.Device, 0, setId)
	redisConn.Send("HSET", db.Device+":name", d.Name, setId)
	redisConn.Send("SADD", db.Device+":addressable:"+d.Addressable.Id.Hex(), setId)
	redisConn.Send("SADD", db.Device+":service:"+d.Service.Id.Hex(), setId)
	redisConn.Send("SADD", db.Device+":profile:"+d.Profile.Id.Hex(), setId)
	for _, label := range d.Labels {
		redisConn.Send("SADD", db.Device+":label:"+label, setId)
	}
	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func readOptionalCommands(i interface{}) []models.Command {
	if i == nil {
		return []models.Command{}
	}

	a := make([]models.Command, len(i.([]interface{})))
	for i, v := range i.([]interface{}) {
		a[i] = models.Command{
			Id: bson.ObjectIdHex(v.(map[string]interface{})["$id"].(map[string]interface{})["$oid"].(string)),
		}
	}

	return a
}

func processDeviceProfile(jsonMap map[string]interface{}) {
	var err error

	setId := jsonMap["_id"].(map[string]interface{})["$oid"].(string)

	// See $GOPATH/src/github.com/edgexfoundry/edgex-go/internal/pkg/db/redis/metadata.go:addDeviceProfile
	d := struct {
		models.DescribedObject
		Id              string
		Name            string
		Manufacturer    string
		Model           string
		Labels          []string
		Objects         interface{}
		DeviceResources []models.DeviceObject
		Resources       []models.ProfileResource
	}{
		DescribedObject: readOptionalDescribedObject(jsonMap),
		Id:              setId,
		Name:            jsonMap["name"].(string),
		Manufacturer:    jsonMap["manufacturer"].(string),
		Model:           jsonMap["model"].(string),
		Objects:         jsonMap["objects"],
		Labels:          readOptionalStringArray(jsonMap["labels"]),
	}

	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(d)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", db.DeviceProfile, 0, setId)
	redisConn.Send("HSET", db.DeviceProfile+":name", d.Name, setId)
	redisConn.Send("SADD", db.DeviceProfile+":manufacturer:"+d.Manufacturer, setId)
	redisConn.Send("SADD", db.DeviceProfile+":model:"+d.Model, setId)
	for _, label := range d.Labels {
		redisConn.Send("SADD", db.DeviceProfile+":label:"+label, setId)
	}

	commands := readOptionalCommands(jsonMap["commands"])
	if len(commands) > 0 {
		cids := redigo.Args{}.Add(db.DeviceProfile + ":commands:" + setId)
		for _, c := range commands {
			cid := c.Id.Hex()
			redisConn.Send("SADD", db.DeviceProfile+":command:"+cid, setId)
			cids = cids.Add(cid)
		}
		redisConn.Send("SADD", cids...)
	}
	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func processDeviceReport(jsonMap map[string]interface{}) {
	var err error

	setId := jsonMap["_id"].(map[string]interface{})["$oid"].(string)
	d := models.DeviceReport{
		BaseObject: readOptionalBaseObject(jsonMap),
		Id:         bson.ObjectIdHex(setId),
		Name:       jsonMap["name"].(string),
		Device:     jsonMap["device"].(string),
		Event:      jsonMap["event"].(string),
	}

	d.Expected = make([]string, len(jsonMap["expected"].([]interface{})))
	for i, v := range jsonMap["expected"].([]interface{}) {
		d.Expected[i] = v.(string)
	}

	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(d)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", db.DeviceReport, 0, setId)
	redisConn.Send("SADD", db.DeviceReport+":device:"+d.Device, setId)
	redisConn.Send("SADD", db.DeviceReport+":scheduleevent:"+d.Event, setId)
	redisConn.Send("HSET", db.DeviceReport+":name", d.Name, setId)
	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func processDeviceService(jsonMap map[string]interface{}) {
	var err error

	setId := jsonMap["_id"].(map[string]interface{})["$oid"].(string)
	d := struct {
		models.DescribedObject
		Id             string
		Name           string
		LastConnected  int64
		LastReported   int64
		OperatingState models.OperatingState
		Addressable    string
		Labels         []string
		AdminState     models.AdminState
	}{
		DescribedObject: readOptionalDescribedObject(jsonMap),
		Id:              setId,
		Name:            jsonMap["name"].(string),
		LastConnected:   int64(jsonMap["lastConnected"].(float64)),
		LastReported:    int64(jsonMap["lastReported"].(float64)),
		OperatingState:  models.OperatingState(jsonMap["operatingState"].(string)),
		Labels:          readOptionalStringArray(jsonMap["labels"]),
		Addressable:     jsonMap["addressable"].(map[string]interface{})["$id"].(map[string]interface{})["$oid"].(string),
		AdminState:      models.AdminState(jsonMap["adminState"].(string)),
	}

	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(d)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", db.DeviceService, 0, setId)
	redisConn.Send("HSET", db.DeviceService+":name", d.Name, setId)
	redisConn.Send("SADD", db.DeviceService+":addressable:"+d.Addressable, setId)
	for _, label := range d.Labels {
		redisConn.Send("SADD", db.DeviceService+":label:"+label, setId)
	}
	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func processProvisionWatcher(jsonMap map[string]interface{}) {
	var err error

	setId := jsonMap["_id"].(map[string]interface{})["$oid"].(string)
	p := struct {
		models.BaseObject
		Id             string
		Name           string
		Identifiers    map[string]string
		Profile        string
		Service        string
		OperatingState models.OperatingState
	}{
		BaseObject:     readOptionalBaseObject(jsonMap),
		Id:             setId,
		Name:           jsonMap["name"].(string),
		Profile:        jsonMap["profile"].(map[string]interface{})["$id"].(map[string]interface{})["$oid"].(string),
		Service:        jsonMap["service"].(map[string]interface{})["$id"].(map[string]interface{})["$oid"].(string),
		OperatingState: models.OperatingState(readOptionalString(jsonMap["operatingState"])),
	}

	p.Identifiers = make(map[string]string)
	for k, v := range jsonMap["identifiers"].(map[string]interface{}) {
		p.Identifiers[k] = v.(string)
	}

	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(p)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", db.ProvisionWatcher, 0, setId)
	redisConn.Send("HSET", db.ProvisionWatcher+":name", p.Name, setId)
	redisConn.Send("SADD", db.ProvisionWatcher+":service:"+p.Service, setId)
	redisConn.Send("SADD", db.ProvisionWatcher+":profile:"+p.Profile, setId)
	for k, v := range p.Identifiers {
		redisConn.Send("SADD", db.ProvisionWatcher+":identifier:"+k+":"+v, setId)
	}

	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func processSchedule(jsonMap map[string]interface{}) {
	var err error

	setId := jsonMap["_id"].(map[string]interface{})["$oid"].(string)
	s := models.Schedule{
		BaseObject: readOptionalBaseObject(jsonMap),
		Id:         bson.ObjectIdHex(setId),
		Name:       jsonMap["name"].(string),
		Start:      readOptionalString(jsonMap["start"]),
		End:        readOptionalString(jsonMap["end"]),
		Frequency:  readOptionalString(jsonMap["frequency"]),
		Cron:       readOptionalString(jsonMap["cron"]),
		RunOnce:    jsonMap["runOnce"].(bool),
	}

	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(s)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", db.Schedule, 0, setId)
	redisConn.Send("HSET", db.Schedule+":name", s.Name, setId)
	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func processScheduleEvent(jsonMap map[string]interface{}) {
	var err error

	setId := jsonMap["_id"].(map[string]interface{})["$oid"].(string)
	s := struct {
		models.BaseObject
		Id          string
		Name        string
		Schedule    string
		Addressable string
		Parameters  string
		Service     string
	}{
		BaseObject:  readOptionalBaseObject(jsonMap),
		Id:          setId,
		Name:        jsonMap["name"].(string),
		Schedule:    jsonMap["schedule"].(string),
		Addressable: jsonMap["addressable"].(map[string]interface{})["$id"].(map[string]interface{})["$oid"].(string),
		Parameters:  readOptionalString(jsonMap["parameters"]),
		Service:     readOptionalString(jsonMap["service"]),
	}

	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(s)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", db.ScheduleEvent, 0, setId)
	redisConn.Send("HSET", db.ScheduleEvent+":name", s.Name, setId)
	redisConn.Send("SADD", db.ScheduleEvent+":addressable:"+s.Addressable, setId)
	redisConn.Send("SADD", db.ScheduleEvent+":schedule:"+s.Schedule, setId)
	redisConn.Send("SADD", db.ScheduleEvent+":service:"+s.Service, setId)

	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

func processRegistration(jsonMap map[string]interface{}) {
	var err error

	fmt.Println(jsonMap)

	setId := jsonMap["_id"].(map[string]interface{})["$oid"].(string)

	r := export.Registration{
		ID:          bson.ObjectIdHex(setId),
		Created:     int64(jsonMap["created"].(float64)),
		Modified:    int64(jsonMap["modified"].(float64)),
		Origin:      int64(jsonMap["origin"].(float64)),
		Name:        jsonMap["name"].(string),
		Addressable: readOptionalAddressable(jsonMap["addressable"]),
		Format:      jsonMap["format"].(string),
		Filter: export.Filter{
			DeviceIDs:          readOptionalStringArray(jsonMap["filter"].(map[string]interface{})["deviceIdentifiers"]),
			ValueDescriptorIDs: readOptionalStringArray(jsonMap["filter"].(map[string]interface{})["valueDescripttorIdentifiers"]),
		},
		Encryption: export.EncryptionDetails{
			Algo:       readOptionalString(jsonMap["encryption"].(map[string]interface{})["encryptionAlgorithm"]),
			Key:        readOptionalString(jsonMap["encryption"].(map[string]interface{})["encryptionKey"]),
			InitVector: readOptionalString(jsonMap["encryption"].(map[string]interface{})["initializingVector"]),
		},
		Compression: jsonMap["compression"].(string),
		Enable:      jsonMap["enable"].(bool),
		Destination: jsonMap["destination"].(string),
	}

	redisConn.Send("MULTI")
	marshalled, _ := bson.Marshal(r)
	redisConn.Send("SET", setId, marshalled)
	redisConn.Send("ZADD", redis.EXPORT_COLLECTION, 0, setId)
	redisConn.Send("HSET", redis.EXPORT_COLLECTION+":name", r.Name, setId)

	_, err = redisConn.Do("EXEC")
	if err != nil {
		log.Fatal(err)
	}
}

var redisConn redigo.Conn

func main() {
	var err error

	handlers := map[string]handler{
		"event":            processEvent,
		"reading":          processReading,
		"valueDescriptor":  processValueDescriptors,
		"addressable":      processAddressable,
		"command":          processCommand,
		"device":           processDevice,
		"deviceProfile":    processDeviceProfile,
		"deviceReport":     processDeviceReport,
		"deviceService":    processDeviceService,
		"provisionWatcher": processProvisionWatcher,
		"schedule":         processSchedule,
		"scheduleEvent":    processScheduleEvent,
		"registration":     processRegistration,
	}

	usage := "Type of input JSON; one of\n"
	for k := range handlers {
		usage += "\t" + k + "\n"
	}

	inputType := flag.String("t", "", usage+"Input file is read from STDIN")

	flag.Parse()
	if *inputType == "" {
		flag.Usage()
		os.Exit(1)
	}

	processor := handlers[*inputType]
	if processor == nil {
		flag.Usage()
		log.Fatal("Unknown input type: " + *inputType)
	}

	// XXX FIXME Use Configuration.
	redisConn, err = redigo.DialURL("redis://localhost:6379")
	if err != nil {
		log.Fatal(err)
	}
	defer redisConn.Close()

	d := json.NewDecoder(os.Stdin)
	for {
		var jsonRecord interface{}

		err = d.Decode(&jsonRecord)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal(err)
		}

		processor(jsonRecord.(map[string]interface{}))
	}
}
