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
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
	"github.com/robfig/cron"
)

const (
	frequencyPattern = `^P(\d+Y)?(\d+M)?(\d+D)?(T(\d+H)?(\d+M)?(\d+S)?)?$`
)

func isIntervalValid(frequency string) bool {
	matched, _ := regexp.MatchString(frequencyPattern, frequency)
	if matched {
		if frequency == "P" || frequency == "PT" {
			matched = false
		}
	}
	return matched
}

// convert millisecond string to Time
func msToTime(ms string) (time.Time, error) {
	msInt, err := strconv.ParseInt(ms, 10, 64)
	if err != nil {
		// todo: support-scheduler will be removed later issue_650a
		t, err := time.Parse(SCHEDULER_TIMELAYOUT, ms)
		if err == nil {
			return t, nil
		}
		return time.Time{}, err
	}

	return time.Unix(0, msInt*int64(time.Millisecond)), nil
}

func restGetAllScheduleEvents(w http.ResponseWriter, r *http.Request) {
	res := make([]models.ScheduleEvent, 0)
	err := dbClient.GetAllScheduleEvents(&res)
	if err != nil {
		LoggingClient.Error("Problem getting schedule events: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(res) > Configuration.Service.ReadMaxLimit {
		err = errors.New("Max limit exceeded")
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		LoggingClient.Error(err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&res)
}

func restAddScheduleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var se models.ScheduleEvent
	if err := json.NewDecoder(r.Body).Decode(&se); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Check the Schedule name
	if se.Schedule == "" {
		// Schedule wasn't passed
		http.Error(w, "Schedule not passed", http.StatusConflict)
		LoggingClient.Error("Schedule not passed")
		return
	}
	var s models.Schedule
	if err := dbClient.GetScheduleByName(&s, se.Schedule); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Schedule not found for schedule event", http.StatusNotFound)
			LoggingClient.Error("Schedule not found for schedule event: " + err.Error())
		} else {
			LoggingClient.Error("Problem getting schedule for schedule event: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Check for the addressable
	// Try by ID
	a, err := dbClient.GetAddressableById(se.Addressable.Id)
	if err != nil {
		// Try by Name
		if a, err = dbClient.GetAddressableByName(se.Addressable.Name); err != nil {
			http.Error(w, "Address not found for schedule event", http.StatusNotFound)
			LoggingClient.Error("Addressable for schedule event not found: " + err.Error())
			return
		}
	}
	se.Addressable = a

	if err := dbClient.AddScheduleEvent(&se); err != nil {
		if err == db.ErrNotUnique {
			http.Error(w, "Duplicate name for schedule event", http.StatusConflict)
			LoggingClient.Error("Duplicate name for schedule event: " + err.Error())
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error("Problem adding schedule event: " + err.Error())
		}
		return
	}

	// Notify Associates
	if err := notifyScheduleEventAssociates(se, http.MethodPost); err != nil {
		LoggingClient.Warn("Problem notifying associated device services for schedule event: " + err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(se.Id.Hex()))
}

func restUpdateScheduleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var from models.ScheduleEvent
	if err := json.NewDecoder(r.Body).Decode(&from); err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the schedule event exists
	to, err := getScheduleEventByIdOrName(from, w)
	if err != nil {
		LoggingClient.Error("Problem getting schedule event: " + err.Error())
		return
	}

	if err := updateScheduleEventFields(from, &to, w); err != nil {
		LoggingClient.Error("Problem updating schedule event: " + err.Error())
		return
	}

	if err := dbClient.UpdateScheduleEvent(to); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Problem updating schedule event: " + err.Error())
		return
	}

	// Notify Associates
	if err := notifyScheduleEventAssociates(to, http.MethodPut); err != nil {
		LoggingClient.Error("Problem notifying associated device services with the schedule event: " + err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

// Return the schedule event for the ID or Name of the passed schedule event
func getScheduleEventByIdOrName(from models.ScheduleEvent, w http.ResponseWriter) (models.ScheduleEvent, error) {
	var se models.ScheduleEvent
	// Try by ID
	if err := dbClient.GetScheduleEventById(&se, from.Id.Hex()); err != nil {
		// Try by Name
		if err = dbClient.GetScheduleEventByName(&se, from.Name); err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Schedule Event not found", http.StatusNotFound)
				LoggingClient.Error(err.Error())
			} else {
				LoggingClient.Error("Problem getting schedule event: " + err.Error())
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
		addr, err := dbClient.GetAddressableById(from.Addressable.Id)
		if err != nil {
			// Try by name
			if addr, err = dbClient.GetAddressableByName(from.Addressable.Name); err != nil {
				if err == db.ErrNotFound {
					http.Error(w, "Addressable not found for schedule event", http.StatusNotFound)
				} else {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
				}
				return err
			}
		}

		to.Addressable = addr
	}
	if from.Service != "" {
		if from.Service != to.Service {
			serviceChanged = true
			// Verify that the new service exists
			if _, err := dbClient.GetDeviceServiceByName(from.Service); err != nil {
				if err == db.ErrNotFound {
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
			if err := dbClient.GetScheduleByName(&checkS, from.Schedule); err != nil {
				if err == db.ErrNotFound {
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
			reports, err := dbClient.GetDeviceReportsByScheduleEventName(to.Name)
			if err != nil {
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
			LoggingClient.Error("Problem notifying associated device services for the schedule event: " + err.Error())
		}
		// Add to new
		if err := notifyScheduleEventAssociates(*to, http.MethodPost); err != nil {
			LoggingClient.Error("Problem notifying associated device services for the schedule event: " + err.Error())
		}
	} else {
		// Changed schedule event
		if err := notifyScheduleEventAssociates(*to, http.MethodPut); err != nil {
			LoggingClient.Error("Problem notifying associated device services for the schedule event: " + err.Error())
		}
	}

	return nil
}

func restGetScheduleEventByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, err := url.QueryUnescape(vars[NAME])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	var res models.ScheduleEvent
	err = dbClient.GetScheduleEventByName(&res, n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Schedule event not found", http.StatusNotFound)
			LoggingClient.Error("Schedule event not found: " + err.Error())
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error("Problem getting schedule event: " + err.Error())
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
	err := dbClient.GetScheduleEventById(&se, id)
	if err != nil {
		http.Error(w, "Schedule event not found", http.StatusNotFound)
		LoggingClient.Error("Schedule event not found: " + err.Error())
		return
	}

	// Delete the schedule event
	if err := deleteScheduleEvent(se, w); err != nil {
		LoggingClient.Error("Problem deleting schedule event: " + err.Error())
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}

	// Check if the schedule event exists
	var se models.ScheduleEvent
	if err := dbClient.GetScheduleEventByName(&se, n); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Schedule event not found", http.StatusNotFound)
			LoggingClient.Error("Schedule event not found: " + err.Error())
		} else {
			LoggingClient.Error("Problem getting schedule event: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Delete the schedule event
	if err := deleteScheduleEvent(se, w); err != nil {
		LoggingClient.Error("Problem deleting schedule event: " + err.Error())
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
	dr, err := dbClient.GetDeviceReportsByScheduleEventName(se.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}
	if len(dr) > 0 {
		err := errors.New("Data integrity issue.  Schedule event is still referenced by device reports, can't delete")
		http.Error(w, err.Error(), http.StatusConflict)
		return err
	}

	if err := dbClient.DeleteScheduleEventById(se.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	// Notify Associates
	if err := notifyScheduleEventAssociates(se, http.MethodDelete); err != nil {
		LoggingClient.Error("Problem notifying associated device services for the schedule event: " + err.Error())
	}

	return nil
}

func restGetScheduleEventById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var did string = vars[ID]
	var res models.ScheduleEvent
	err := dbClient.GetScheduleEventById(&res, did)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Schedule event not found", http.StatusNotFound)
			LoggingClient.Error("Schedule event not found: " + err.Error())
		} else {
			LoggingClient.Error("Problem getting schedule event: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
	_, err := dbClient.GetAddressableById(aid)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Addressable not found for schedule event", http.StatusNotFound)
			LoggingClient.Error("Addressable not found for schedule event: " + err.Error())
		} else {
			LoggingClient.Error(err.Error())
			http.Error(w, "Problem getting addressable for schedule event: "+err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	// Get the schedule events
	if err := dbClient.GetScheduleEventsByAddressableId(&res, aid); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error("Problem getting schedule events: " + err.Error())
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error(err.Error())
		return
	}
	var res []models.ScheduleEvent = make([]models.ScheduleEvent, 0)

	// Check if the addressable exists
	a, err := dbClient.GetAddressableByName(an)
	if err != nil {
		if err == db.ErrNotFound {
			LoggingClient.Error("Addressable not found for schedule event: " + err.Error())
			http.Error(w, "Addressable not found for schedule event", http.StatusNotFound)
		} else {
			LoggingClient.Error("Problem getting addressable for schedule event: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Get the schedule events
	if err = dbClient.GetScheduleEventsByAddressableId(&res, a.Id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Problem getting schedule events: " + err.Error())
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
		LoggingClient.Error(err.Error())
		return
	}
	var res []models.ScheduleEvent = make([]models.ScheduleEvent, 0)

	// Check if the service exists
	if _, err = dbClient.GetDeviceServiceByName(sn); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Service not found for schedule event", http.StatusNotFound)
			LoggingClient.Error("Device service not found for schedule event: " + err.Error())
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			LoggingClient.Error("Problem getting device service for schedule event: " + err.Error())
		}
		return
	}

	// Get the schedule events
	if err = dbClient.GetScheduleEventsByServiceName(&res, sn); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		LoggingClient.Error("Problem getting schedule events: " + err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func restGetAllSchedules(w http.ResponseWriter, _ *http.Request) {
	res := make([]models.Schedule, 0)
	err := dbClient.GetAllSchedules(&res)
	if err != nil {
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check max length
	if len(res) > Configuration.Service.ReadMaxLimit {
		err = errors.New("Max limit exceeded")
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		LoggingClient.Error(err.Error())
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the name is unique
	var checkS models.Schedule
	if err := dbClient.GetScheduleByName(&checkS, s.Name); err != nil {
		if err != db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error("Schedule not found: " + err.Error())
			return
		}
	} else {
		err := errors.New("Schedule already exists with name: " + s.Name)
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Validate the time format
	if s.Start != "" {
		if _, err := msToTime(s.Start); err != nil {
			LoggingClient.Error("Incorrect start time format: " + err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if s.End != "" {
		if _, err := msToTime(s.End); err != nil {
			LoggingClient.Error("Incorrect end time format: " + err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if s.Frequency != "" {
		if !isIntervalValid(s.Frequency) {
			err := errors.New("Frequency format incorrect: " + s.Frequency)
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if err := dbClient.AddSchedule(&s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Problem adding schedule: " + err.Error())
		return
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
		LoggingClient.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the schedule exists
	var to models.Schedule
	// Try by ID
	if err := dbClient.GetScheduleById(&to, from.Id.Hex()); err != nil {
		// Try by name
		if err = dbClient.GetScheduleByName(&to, from.Name); err != nil {
			LoggingClient.Error("Schedule not found: " + err.Error())
			http.Error(w, "Schedule not found", http.StatusNotFound)
			return
		}
	}

	if err := updateScheduleFields(from, &to, w); err != nil {
		LoggingClient.Error("Problem updating schedule: " + err.Error())
		return
	}

	if err := dbClient.UpdateSchedule(to); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Problem updating schedule: " + err.Error())
		return
	}

	// Notify Associates
	if err := notifyScheduleAssociates(to, http.MethodPut); err != nil {
		LoggingClient.Error("Problem notifying associated device services for schedule: " + err.Error())
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
		if err := dbClient.GetScheduleByName(&checkS, from.Name); err != nil {
			if err != db.ErrNotFound {
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
	err := dbClient.GetScheduleById(&res, sid)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Schedule not found", http.StatusNotFound)
			LoggingClient.Error("Schedule not found: " + err.Error())
		} else {
			LoggingClient.Error("Problem getting schedule: " + err.Error())
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
		LoggingClient.Error(err.Error())
		return
	}

	var res models.Schedule
	err = dbClient.GetScheduleByName(&res, n)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Schedule not found", http.StatusNotFound)
			LoggingClient.Error("Schedule not found: " + err.Error())
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			LoggingClient.Error("Problem getting schedule: " + err.Error())
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
	if err := dbClient.GetScheduleById(&s, id); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Schedule not found", http.StatusNotFound)
			LoggingClient.Error("Schedule not found: " + err.Error())
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			LoggingClient.Error("Problem getting schedule: " + err.Error())
		}
		return
	}

	if err := deleteSchedule(s, w); err != nil {
		LoggingClient.Error("Problem deleting schedule: " + err.Error())
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
		LoggingClient.Error(err.Error())
		return
	}

	// Check if the schedule exists
	var s models.Schedule
	if err = dbClient.GetScheduleByName(&s, n); err != nil {
		if err == db.ErrNotFound {
			LoggingClient.Error("Schedule not found: " + err.Error())
			http.Error(w, "Schedule not found", http.StatusNotFound)
		} else {
			LoggingClient.Error("Problem getting schedule: " + err.Error())
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	// Delete the schedule
	if err = deleteSchedule(s, w); err != nil {
		LoggingClient.Error("Problem deleting schedule: " + err.Error())
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

	if err := dbClient.DeleteScheduleById(s.Id.Hex()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return err
	}

	return nil
}

// Determine if the scheule is still in use by schedule events
func isScheduleStillInUse(s models.Schedule) (bool, error) {
	var scheduleEvents []models.ScheduleEvent
	if err := dbClient.GetScheduleEventsByScheduleName(&scheduleEvents, s.Name); err != nil {
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
	if err := dbClient.GetScheduleEventsByScheduleName(&events, s.Name); err != nil {
		return err
	}

	// Get the device services for the schedule events
	services := []models.DeviceService{}
	for _, se := range events {
		ds, err := dbClient.GetDeviceServiceByName(se.Service)
		if err != nil {
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
	ds, err := dbClient.GetDeviceServiceByName(se.Service)
	if err != nil {
		return err
	}

	// Notify the associated device service
	if err := notifyAssociates([]models.DeviceService{ds}, se.Id.Hex(), action, models.SCHEDULEEVENT); err != nil {
		return err
	}

	return nil
}
