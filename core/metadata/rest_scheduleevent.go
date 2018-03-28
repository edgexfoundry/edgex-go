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
 * @author: Spencer Bull & Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package metadata

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/gorilla/mux"
	"github.com/robfig/cron"
	"gopkg.in/mgo.v2"
)

func isIntervalValid(frequency string) bool {
	_, err := strconv.Atoi("-42")
	if err != nil {
		return false
	}
	return true
}

// convert millisecond string to Time
func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

func restGetAllScheduleEvents(w http.ResponseWriter, r *http.Request) {
	res := make([]models.ScheduleEvent, 0)
	err := getAllScheduleEvents(&res)
	if err != nil {
		loggingClient.Error("Problem getting schedule events: "+err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	if len(res) > configuration.ReadMaxLimit {
		err = errors.New("Max limit exceeded")
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		loggingClient.Error(err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&res)
}

func restAddScheduleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var se models.ScheduleEvent
	if err := json.NewDecoder(r.Body).Decode(&se); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	// Check the Schedule name
	if se.Schedule == "" {
		// Schedule wasn't passed
		http.Error(w, "Schedule not passed", http.StatusConflict)
		loggingClient.Error("Schedule not passed", "")
		return
	}
	var s models.Schedule
	if err := getScheduleByName(&s, se.Schedule); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, "Schedule not found for schedule event", http.StatusNotFound)
			loggingClient.Error("Schedule not found for schedule event: "+err.Error(), "")
		} else {
			loggingClient.Error("Problem getting schedule for schedule event: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	// Check for the addressable
	// Try by ID
	var a models.Addressable
	if err := getAddressableById(&a, se.Addressable.Id.Hex()); err != nil {
		// Try by Name
		if err = getAddressableByName(&a, se.Addressable.Name); err != nil {
			http.Error(w, "Address not found for schedule event", http.StatusNotFound)
			loggingClient.Error("Addressable for schedule event not found: "+err.Error(), "")
			return
		}
	}
	se.Addressable = a

	/*if len(se.Addressable.Name) != 0 {
		if err := getAddressableByName(&se.Addressable, se.Addressable.Name); err != nil {
			if err == mgo.ErrNotFound {
				http.Error(w, "Unknown Addressable", http.StatusNotFound)
				LOGGER.Println(err.Error())
				return
			}
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			LOGGER.Println(err.Error())
			return
		}
	} else if len(se.Addressable.Id.Hex()) != 0 {
		if err := getAddressableById(&se.Addressable, se.Addressable.Id.Hex()); err != nil {
			if err == mgo.ErrNotFound {
				http.Error(w, "Unknown Addressable", http.StatusNotFound)
				LOGGER.Println(err.Error())
				return
			}
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			LOGGER.Println(err.Error())
			return
		}
	} else {
		http.Error(w, "Unknown Addressable", http.StatusNotFound)
		LOGGER.Println("Unknown Addressable")
		return
	}*/

	if err := addScheduleEvent(&se); err != nil {
		if err == ErrDuplicateName {
			http.Error(w, "Duplicate name for schedule event", http.StatusConflict)
			loggingClient.Error("Duplicate name for schedule event: "+err.Error(), "")
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Problem adding schedule event: "+err.Error(), "")
		}
		return
	}

	// Notify Associates
	if err := notifyScheduleEventAssociates(se, http.MethodPost); err != nil {
		loggingClient.Error("Problem notifying associated device services for schedule event: "+err.Error(), "")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(se.Id.Hex()))
}

func restUpdateScheduleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var from models.ScheduleEvent
	if err := json.NewDecoder(r.Body).Decode(&from); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the schedule event exists
	var to models.ScheduleEvent
	to, err := getScheduleEventByIdOrName(from, w)
	if err != nil {
		loggingClient.Error("Problem getting schedule event: "+err.Error(), "")
		return
	}

	if err := updateScheduleEventFields(from, &to, w); err != nil {
		loggingClient.Error("Problem updating schedule event: "+err.Error(), "")
		return
	}

	if err := updateScheduleEvent(to); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem updating schedule event: "+err.Error(), "")
		return
	}

	// Notify Associates
	if err := notifyScheduleEventAssociates(to, http.MethodPut); err != nil {
		loggingClient.Error("Problem notifying associated device services with the schedule event: "+err.Error(), "")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Return the schedule event for the ID or Name of the passed schedule event
func getScheduleEventByIdOrName(from models.ScheduleEvent, w http.ResponseWriter) (models.ScheduleEvent, error) {
	var se models.ScheduleEvent
	// Try by ID
	if err := getScheduleEventById(&se, from.Id.Hex()); err != nil {
		// Try by Name
		if err = getScheduleEventByName(&se, from.Name); err != nil {
			if err == mgo.ErrNotFound {
				http.Error(w, "Schedule Event not found", http.StatusNotFound)
				loggingClient.Error(err.Error(), "")
			} else {
				loggingClient.Error("Problem getting schedule event: "+err.Error(), "")
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			return se, err
		}
	}

	return se, nil
}

// Update the relevant fields for the schedule event
func updateScheduleEventFields(from models.ScheduleEvent, to *models.ScheduleEvent, w http.ResponseWriter) error {
	// Boolean used to notify the proper associates based on if the service changed
	var serviceChanged bool
	oldService := "" // Hold the old service name in case it changes

	// Use .String() method to compare empty structs (not ideal, but there's no .equals() method)
	if (from.Addressable.String() != models.Addressable{}.String()) {
		// Check if the new addressable exists
		// Try by ID
		if err := getAddressableById(&from.Addressable, from.Addressable.Id.Hex()); err != nil {
			// Try by name
			if err = getAddressableByName(&from.Addressable, from.Addressable.Name); err != nil {
				if err == mgo.ErrNotFound {
					http.Error(w, "Addressable not found for schedule event", http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
				}
				return err
			}
		}

		to.Addressable = from.Addressable
	}
	if from.Service != "" {
		if from.Service != to.Service {
			serviceChanged = true
			// Verify that the new service exists
			var checkDS models.DeviceService
			if err := getDeviceServiceByName(&checkDS, from.Service); err != nil {
				if err == mgo.ErrNotFound {
					http.Error(w, "Device Service not found for schedule event", http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
				}
				return err
			}

			oldService = to.Service
			to.Service = from.Service
		}
	}
	if from.Schedule != "" {
		if from.Schedule != to.Schedule {
			// Verify that the new schedule exists
			var checkS models.Schedule
			if err := getScheduleByName(&checkS, from.Schedule); err != nil {
				if err == mgo.ErrNotFound {
					http.Error(w, "Schedule not found for schedule event", http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
				}
				return err
			}
		}

		to.Schedule = from.Schedule
	}
	if from.Name != "" {
		if from.Name != to.Name {
			// Verify data integrity
			var reports []models.DeviceReport
			if err := getDeviceReportsByScheduleEventName(&reports, to.Name); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return err
			}
			if len(reports) > 0 {
				err := errors.New("Data integrity issue.  Schedule is still referenced by device reports, can't change the name")
				http.Error(w, err.Error(), http.StatusConflict)
				return err
			}
		}

		to.Name = from.Name
	}
	if from.Origin != 0 {
		to.Origin = from.Origin
	}

	// Notify associates (based on if the service changed)
	if serviceChanged {
		// Delete from old
		if err := notifyScheduleEventAssociates(models.ScheduleEvent{Name: oldService}, http.MethodDelete); err != nil {
			loggingClient.Error("Problem notifying associated device services for the schedule event: "+err.Error(), "")
		}
		// Add to new
		if err := notifyScheduleEventAssociates(*to, http.MethodPost); err != nil {
			loggingClient.Error("Problem notifying associated device services for the schedule event: "+err.Error(), "")
		}
	} else {
		// Changed schedule event
		if err := notifyScheduleEventAssociates(*to, http.MethodPut); err != nil {
			loggingClient.Error("Problem notifying associated device services for the schedule event: "+err.Error(), "")
		}
	}

	return nil
}

func restGetScheduleEventByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	var res models.ScheduleEvent
	err = getScheduleEventByName(&res, n)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, "Schedule event not found", http.StatusNotFound)
			loggingClient.Error("Schedule event not found: "+err.Error(), "")
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Problem getting schedule event: "+err.Error(), "")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restDeleteScheduleEventById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]

	// Check if the schedule event exists
	var se models.ScheduleEvent
	err := getScheduleEventById(&se, id)
	if err != nil {
		http.Error(w, "Schedule event not found", http.StatusNotFound)
		loggingClient.Error("Schedule event not found: "+err.Error(), "")
		return
	}

	// Delete the schedule event
	if err := deleteScheduleEvent(se, w); err != nil {
		loggingClient.Error("Problem deleting schedule event: "+err.Error(), "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restDeleteScheduleEventByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	// Check if the schedule event exists
	var se models.ScheduleEvent
	if err := getScheduleEventByName(&se, n); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, "Schedule event not found", http.StatusNotFound)
			loggingClient.Error("Schedule event not found: "+err.Error(), "")
		} else {
			loggingClient.Error("Problem getting schedule event: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	// Delete the schedule event
	if err := deleteScheduleEvent(se, w); err != nil {
		loggingClient.Error("Problem deleting schedule event: "+err.Error(), "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the schedule event
// 409 error if the schedule event is still in use by device reports
func deleteScheduleEvent(se models.ScheduleEvent, w http.ResponseWriter) error {
	// Check if the schedule event is still in use by device reports
	var dr []models.DeviceReport
	if err := getDeviceReportsByScheduleEventName(&dr, se.Name); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}
	if len(dr) > 0 {
		err := errors.New("Data integrity issue.  Schedule event is still referenced by device reports, can't delete")
		http.Error(w, err.Error(), http.StatusConflict)
		return err
	}

	if err := deleteById(SECOL, se.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	// Notify Associates
	if err := notifyScheduleEventAssociates(se, http.MethodDelete); err != nil {
		loggingClient.Error("Problem notifying associated device services for the schedule event: "+err.Error(), "")
	}

	return nil
}

func restGetScheduleEventById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	var res models.ScheduleEvent
	err := getScheduleEventById(&res, did)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, "Schedule event not found", http.StatusNotFound)
			loggingClient.Error("Schedule event not found: "+err.Error(), "")
		} else {
			loggingClient.Error("Problem getting schedule event: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Get all the schedule events by their associated addressable
// 404 if addressable not found
func restGetScheduleEventByAddressableId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var aid string = vars[ADDRESSABLEID]
	var res []models.ScheduleEvent = make([]models.ScheduleEvent, 0)

	// Check if the addressable exists
	var a models.Addressable
	if err := getAddressableById(&a, aid); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, "Addressable not found for schedule event", http.StatusNotFound)
			loggingClient.Error("Addressable not found for schedule event: "+err.Error(), "")
		} else {
			loggingClient.Error(err.Error(), "")
			http.Error(w, "Problem getting addressable for schedule event: "+err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	// Get the schedule events
	if err := getScheduleEventsByAddressableId(&res, aid); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem getting schedule events: "+err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Get all of the schedule events by the associated addressable (by name)
// 404 if the addressable can't be found
func restGetScheduleEventByAddressableName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	an, err := url.QueryUnescape(vars[ADDRESSABLENAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}
	var res []models.ScheduleEvent = make([]models.ScheduleEvent, 0)

	// Check if the addressable exists
	var a models.Addressable
	if err = getAddressableByName(&a, an); err != nil {
		if err == mgo.ErrNotFound {
			loggingClient.Error("Addressable not found for schedule event: "+err.Error(), "")
			http.Error(w, "Addressable not found for schedule event", http.StatusNotFound)
		} else {
			loggingClient.Error("Problem getting addressable for schedule event: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	// Get the schedule events
	if err = getScheduleEventsByAddressableId(&res, a.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem getting schedule events: "+err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Get all of the schedule events by the service name
// 404 if no service matches on the name provided
func restGetScheduleEventsByServiceName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sn, err := url.QueryUnescape(vars[SERVICENAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}
	var res []models.ScheduleEvent = make([]models.ScheduleEvent, 0)

	// Check if the service exists
	var ds models.DeviceService
	if err = getDeviceServiceByName(&ds, sn); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, "Service not found for schedule event", http.StatusNotFound)
			loggingClient.Error("Device service not found for schedule event: "+err.Error(), "")
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Problem getting device service for schedule event: "+err.Error(), "")
		}
		return
	}

	// Get the schedule events
	if err = getScheduleEventsByServiceName(&res, sn); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem getting schedule events: "+err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetAllSchedules(w http.ResponseWriter, _ *http.Request) {
	res := make([]models.Schedule, 0)
	err := getAllSchedules(&res)
	if err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check max length
	if len(res) > MAX_LIMIT {
		err = errors.New("Max limit exceeded")
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		loggingClient.Error(err.Error(), "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&res)
}

// Add a new schedule - name must be unique
// Return  error 409 if the cron string is not properly formatted
func restAddSchedule(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var s models.Schedule
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the name is unique
	var checkS models.Schedule
	if err := getScheduleByName(&checkS, s.Name); err != nil {
		if err != mgo.ErrNotFound {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Schedule not found: "+err.Error(), "")
			return
		}
	} else {
		err := errors.New("Schedule already exists with name: " + s.Name)
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Validate the time format
	if s.Start != "" {
		if _, err := msToTime(s.Start); err != nil {
			loggingClient.Error("Incorrect start time format: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}
	if s.End != "" {
		if _, err := msToTime(s.End); err != nil {
			loggingClient.Error("Incorrect end time format: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}
	if s.Frequency != "" {
		if !isIntervalValid(s.Frequency) {
			err := errors.New("Frequency format incorrect: " + s.Frequency)
			loggingClient.Error("Frequency format is incorrect: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
	}

	if err := addSchedule(&s); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem adding schedule: "+err.Error(), "")
		return
	}

	// Notify Associates
	if err := notifyScheduleAssociates(s, http.MethodPost); err != nil {
		loggingClient.Error("Problem notifying associated device services for schedule: "+err.Error(), "")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(s.Id.Hex()))
}

// Update a schedule
// Use ID then name to find the schedule (404 if not found)
// 409 if the new cron string is not properly formatted
func restUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var from models.Schedule
	if err := json.NewDecoder(r.Body).Decode(&from); err != nil {
		loggingClient.Error(err.Error(), "")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check if the schedule exists
	var to models.Schedule
	// Try by ID
	if err := getScheduleById(&to, from.Id.Hex()); err != nil {
		// Try by name
		if err = getScheduleByName(&to, from.Name); err != nil {
			loggingClient.Error("Schedule not found: "+err.Error(), "")
			http.Error(w, "Schedule not found", http.StatusNotFound)
			return
		}
	}

	if err := updateScheduleFields(from, &to, w); err != nil {
		loggingClient.Error("Problem updating schedule: "+err.Error(), "")
		return
	}

	if err := updateSchedule(to); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Problem updating schedule: "+err.Error(), "")
		return
	}

	// Notify Associates
	if err := notifyScheduleAssociates(to, http.MethodPut); err != nil {
		loggingClient.Error("Problem notifying associated device services for schedule: "+err.Error(), "")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Update the relevant fields for the schedule
func updateScheduleFields(from models.Schedule, to *models.Schedule, w http.ResponseWriter) error {
	if from.Cron != "" {
		if _, err := cron.Parse(from.Cron); err != nil {
			err = errors.New("Invalid cron format")
			http.Error(w, err.Error(), http.StatusConflict)
			return err
		}
		to.Cron = from.Cron
	}
	if from.End != "" {
		if _, err := msToTime(from.End); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return err
		}
		to.End = from.End
	}
	if from.Frequency != "" {
		if !isIntervalValid(from.Frequency) {
			err := errors.New("Frequency format is incorrect: " + from.Frequency)
			http.Error(w, err.Error(), http.StatusConflict)
			return err
		}
		to.Frequency = from.Frequency
	}
	if from.Start != "" {
		if _, err := msToTime(from.Start); err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return err
		}
		to.Start = from.Start
	}
	if from.Origin != 0 {
		to.Origin = from.Origin
	}
	if from.Name != "" && from.Name != to.Name {
		// Check if new name is unique
		var checkS models.Schedule
		if err := getScheduleByName(&checkS, from.Name); err != nil {
			if err != mgo.ErrNotFound {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
		} else {
			if checkS.Id != to.Id {
				err := errors.New("Duplicate name for the schedule")
				http.Error(w, err.Error(), http.StatusConflict)
				return err
			}
		}

		// Check if the schedule still has attached schedule events
		stillInUse, err := isScheduleStillInUse(*to)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return err
		}
		if stillInUse {
			err = errors.New("Schedule is still in use, can't change the name")
			http.Error(w, err.Error(), http.StatusConflict)
			return err
		}

		to.Name = from.Name
	}

	return nil
}

// Get a schedule by its ID
func restGetScheduleById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var sid string = vars[ID]
	var res models.Schedule
	err := getScheduleById(&res, sid)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, "Schedule not found", http.StatusNotFound)
			loggingClient.Error("Schedule not found: "+err.Error(), "")
		} else {
			loggingClient.Error("Problem getting schedule: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// Get a schedule by its Name
func restGetScheduleByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	var res models.Schedule
	err = getScheduleByName(&res, n)
	if err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, "Schedule not found", http.StatusNotFound)
			loggingClient.Error("Schedule not found: "+err.Error(), "")
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Problem getting schedule: "+err.Error(), "")
		}

		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restDeleteScheduleById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var id string = vars[ID]

	// Check if the schedule exists
	var s models.Schedule
	if err := getScheduleById(&s, id); err != nil {
		if err == mgo.ErrNotFound {
			http.Error(w, "Schedule not found", http.StatusNotFound)
			loggingClient.Error("Schedule not found: "+err.Error(), "")
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Problem getting schedule: "+err.Error(), "")
		}
		return
	}

	if err := deleteSchedule(s, w); err != nil {
		loggingClient.Error("Problem deleting schedule: "+err.Error(), "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restDeleteScheduleByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error(err.Error(), "")
		return
	}

	// Check if the schedule exists
	var s models.Schedule
	if err = getScheduleByName(&s, n); err != nil {
		if err == mgo.ErrNotFound {
			loggingClient.Error("Schedule not found: "+err.Error(), "")
			http.Error(w, "Schedule not found", http.StatusNotFound)
		} else {
			loggingClient.Error("Problem getting schedule: "+err.Error(), "")
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	// Delete the schedule
	if err = deleteSchedule(s, w); err != nil {
		loggingClient.Error("Problem deleting schedule: "+err.Error(), "")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Delete the schedule
func deleteSchedule(s models.Schedule, w http.ResponseWriter) error {
	stillInUse, err := isScheduleStillInUse(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}
	if stillInUse {
		err = errors.New("Schedule is still in use by schedule events, can't delete")
		http.Error(w, err.Error(), http.StatusConflict)
		return err
	}

	if err := deleteById(SCOL, s.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	// Notify Associates
	if err := notifyScheduleAssociates(s, http.MethodDelete); err != nil {
		loggingClient.Error("Problem notifying associated device services for schedule: "+err.Error(), "")
	}

	return nil
}

// Determine if the scheule is still in use by schedule events
func isScheduleStillInUse(s models.Schedule) (bool, error) {
	var scheduleEvents []models.ScheduleEvent
	if err := getScheduleEventsByScheduleName(&scheduleEvents, s.Name); err != nil {
		return false, err
	}
	if len(scheduleEvents) > 0 {
		return true, nil
	}

	return false, nil
}

// Notify the associated device services for the schedule
func notifyScheduleAssociates(s models.Schedule, action string) error {
	// Get the associated schedule events
	var events []models.ScheduleEvent
	if err := getScheduleEventsByScheduleName(&events, s.Name); err != nil {
		return err
	}

	// Get the device services for the schedule events
	var services []models.DeviceService
	for _, se := range events {
		var ds models.DeviceService
		if err := getDeviceServiceByName(&ds, se.Service); err != nil {
			return err
		}
		services = append(services, ds)
	}

	// Notify the associated device services
	if err := notifyAssociates(services, s.Id.Hex(), action, models.SCHEDULE); err != nil {
		return err
	}

	return nil
}

// Notify the associated device service for the schedule event
func notifyScheduleEventAssociates(se models.ScheduleEvent, action string) error {
	// Get the associated device service
	var ds models.DeviceService
	if err := getDeviceServiceByName(&ds, se.Service); err != nil {
		return err
	}

	var services []models.DeviceService
	services = append(services, ds)

	// Notify the associated device service
	if err := notifyAssociates(services, se.Id.Hex(), action, models.SCHEDULEEVENT); err != nil {
		return err
	}

	return nil
}
