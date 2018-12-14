package scheduler

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
)

// Utility function for adding configured locally schedulers and scheduled events
func AddSchedulers() error {

	// ensure maps are clean
	clearMaps()

	// ensure queue is empty
	clearQueue()

	LoggingClient.Info("Loading schedules, schedule events, and addressables ...")

	// load data from core-metadata
	err := loadCoreMetadataInformation()
	if err != nil {
		LoggingClient.Error("failed to load information from core-metadata", "message", err.Error())
		return err
	}

	// load config schedules
	errCS := loadConfigSchedules()
	if errCS != nil {
		LoggingClient.Error("failed to load scheduler config data", "message", errCS.Error())
		return errCS
	}

	// load config schedule events
	errCSE := loadConfigScheduleEvents()
	if errCSE != nil {
		LoggingClient.Error("failed to load scheduler events config data", "message", errCSE.Error())
		return errCSE
	}

	LoggingClient.Info("Finished loading schedules, schedule events, and addressables")

	return nil
}

// check meta data service
func RetryService(timeout int, wait *sync.WaitGroup, ch chan string) {
	now := time.Now()
	until := now.Add(time.Millisecond * time.Duration(timeout))

	var status = 0
	for time.Now().Before(until) {
		var err error
		//When looping, only handle configuration if it hasn't already been set.
		if status != 200 {
			status, err = callMetaDataService()
			if err != nil {
				ch <- "Support Scheduler failed to connect to Core Metadata...retrying"
			} else {
				//Check against boot timeout default
				if Configuration.Service.BootTimeout != timeout {
					until = now.Add(time.Millisecond * time.Duration(Configuration.Service.BootTimeout))
				}
			}
		}

		if status == 200 {
			ch <- "Support Scheduler established connection to Core Metadata"
			break
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

// ensure we have core metadata available
func callMetaDataService() (int, error) {

	client := &http.Client{
		Timeout: time.Duration(Configuration.Service.Timeout) * time.Millisecond,
	}
	executingUrl := fmt.Sprintf("%s%s", Configuration.Clients["Metadata"].Url(), clients.ApiPingRoute)

	req, _ := http.NewRequest(http.MethodGet, executingUrl, nil)
	req.Header.Set(ContentTypeKey, ContentTypeJsonValue)

	_, statusCode, err := sendRequestAndGetResponse(client, req)
	if err != nil {
		return statusCode, err
	}
	LoggingClient.Debug("execution complete", "code", statusCode)
	return statusCode, nil
}

// Query core-metadata scheduler client get schedules
func getMetadataSchedules() ([]models.Schedule, error) {

	var receivedSchedules []models.Schedule
	receivedSchedules, errSchedule := msc.Schedules()
	if errSchedule != nil {
		LoggingClient.Error("error connecting to metadata and retrieving schedules", "message", errSchedule.Error())
		return receivedSchedules, errSchedule
	}

	if receivedSchedules != nil {
		LoggingClient.Debug("Successfully queried core-metadata schedules...")
		for _, v := range receivedSchedules {
			LoggingClient.Debug("found schedule", "id", v.Id.Hex(), "schedule", v.Name, "start", v.Start)
		}
	}
	return receivedSchedules, nil
}

// Query core-metadata schedulerEvent client get scheduledEvents
func getMetadataScheduleEvents() ([]models.ScheduleEvent, error) {

	var receivedScheduleEvents []models.ScheduleEvent
	receivedScheduleEvents, err := msec.ScheduleEvents()
	if err != nil {
		LoggingClient.Error("error connecting to metadata and retrieving schedule events", "message", err.Error())
		return receivedScheduleEvents, err
	}

	// debug information only
	if receivedScheduleEvents != nil {
		LoggingClient.Debug("Successfully queried core-metadata schedule events...")
		for _, v := range receivedScheduleEvents {
			LoggingClient.Debug("found schedule", "id", v.Id.Hex(), "event", v.Name, "schedule", v.Schedule, "service", v.Service)
		}
	}

	return receivedScheduleEvents, nil
}

// Iterate over the received schedules add them to scheduler
func addReceivedSchedules(schedules []models.Schedule) error {

	for _, schedule := range schedules {
		// todo: need to remove this naming convention based inference
		matched, err := regexp.MatchString("device.*", schedule.Name)
		if err != nil {
			LoggingClient.Error("error parsing received core-metadata schedules", "message", err.Error())
			return err
		}
		// we have a service related notification
		if !matched {
			err := addSchedule(schedule)
			if err != nil {
				LoggingClient.Error("error adding core-metadata schedule", "schedule", schedule.Name, "message", err.Error())
				return err
			}
			LoggingClient.Info("associated schedule name and id", "schedule", schedule.Name, "id", schedule.Id.Hex())
		}
	}
	return nil
}

// Iterate over the received schedule event(s)
func addReceivedScheduleEvents(scheduleEvents []models.ScheduleEvent) error {

	for _, scheduleEvent := range scheduleEvents {
		// todo: need to remove this naming convention based inference
		matched, err := regexp.MatchString("device.*", scheduleEvent.Service)
		if err != nil {
			LoggingClient.Error("error parsing received core-metadata schedules", "message", err.Error())
			return err
		}
		// schedule event service should not be device.*
		if !matched {
			err := addScheduleEvent(scheduleEvent)
			if err != nil {
				LoggingClient.Error("error adding core-metadata schedule event", "event", scheduleEvent.Name, "message", err.Error())
				return err
			}
			LoggingClient.Info("added event to schedule", "event", scheduleEvent.Name, "schedule", scheduleEvent.Schedule, "id", scheduleEvent.Id.Hex())
		}
	}

	return nil
}

// Add schedule to core-metadata
func addScheduleToCoreMetaData(schedule models.Schedule) (string, error) {

	addedScheduleId, err := msc.Add(&schedule)
	if err != nil {
		LoggingClient.Error("error trying to add schedule to core-metadata service", "message", err.Error())
		return "", err
	}
	LoggingClient.Info("Added schedule to the core-metadata", "schedule", schedule.Name, "id", addedScheduleId)
	return addedScheduleId, nil
}

// Add schedule event to core-metadata
func addScheduleEventToCoreMetadata(scheduleEvent models.ScheduleEvent) (string, error) {

	addedScheduleEventId, err := msec.Add(&scheduleEvent)
	if err != nil {
		LoggingClient.Error("error trying to add schedule event to core-metadata service", "message", err.Error())
		return "", err
	}
	LoggingClient.Info("Added schedule event to the core-metadata", "event", scheduleEvent.Name, "id", addedScheduleEventId)
	return addedScheduleEventId, nil
}

// Load schedules
func loadConfigSchedules() error {

	schedules := Configuration.Schedules
	for i := range schedules {
		schedule := models.Schedule{
			BaseObject: models.BaseObject{},
			Name:       schedules[i].Name,
			Start:      schedules[i].Start,
			End:        schedules[i].End,
			Frequency:  schedules[i].Frequency,
			Cron:       schedules[i].Cron,
			RunOnce:    schedules[i].RunOnce,
		}
		_, errExistingSchedule := queryScheduleByName(schedule.Name)

		if errExistingSchedule != nil {
			// add the schedule core-metadata
			newScheduleId, errAddedSchedule := addScheduleToCoreMetaData(schedule)
			if errAddedSchedule != nil {
				LoggingClient.Error("error adding to the scheduler", "message", errAddedSchedule.Error())
				return errAddedSchedule
			}

			// add the core-metadata scheduler.id
			schedule.Id = bson.ObjectId(newScheduleId)

			// add the schedule to the scheduler
			err := addSchedule(schedule)

			if err != nil {
				LoggingClient.Error("error loading schedule %s from the scheduler config", "message", err.Error())
				return err
			}
		} else {
			LoggingClient.Debug("did not add schedule %s as it already exists in the scheduler", "schedule", schedule.Name)
		}
	}

	return nil
}

// Load schedule events and associated addressable(s) if required
func loadConfigScheduleEvents() error {

	scheduleEvents := Configuration.ScheduleEvents

	for e := range scheduleEvents {

		addressable := models.Addressable{
			Name:       fmt.Sprintf("schedule-%s", scheduleEvents[e].Name),
			Path:       scheduleEvents[e].Path,
			Port:       scheduleEvents[e].Port,
			Protocol:   scheduleEvents[e].Protocol,
			HTTPMethod: scheduleEvents[e].Method,
			Address:    scheduleEvents[e].Host,
		}

		scheduleEvent := models.ScheduleEvent{
			Name:        scheduleEvents[e].Name,
			Schedule:    scheduleEvents[e].Schedule,
			Parameters:  scheduleEvents[e].Parameters,
			Service:     scheduleEvents[e].Service,
			Addressable: addressable,
		}

		// fetch existing queue and determine of scheduleEvent exists
		_, err := queryScheduleEventByName(scheduleEvent.Name)

		if err != nil {
			// query core-metadata for addressable
			_, err := mac.AddressableForName(addressable.Name)
			if err != nil {
				// we don't have that addressable yet now add it
				addressableId, err := mac.Add(&addressable)
				if err != nil {
					LoggingClient.Error("error adding new addressable into core-metadata", "message", err.Error())
					return err
				}
				LoggingClient.Info("Added addressable into core-metadata", "addressable", addressable.Name, "id", addressableId, "path", addressable.Path)

				// add the core-metadata id value
				addressable.Id = addressableId
			}

			// add the schedule event with addressable event to core-metadata
			newScheduleEventId, err := addScheduleEventToCoreMetadata(scheduleEvent)
			if err != nil {
				LoggingClient.Error("error adding schedule event %s into core-metadata", "message", err.Error())
				return err
			}

			// add the core-metadata version of the scheduleEvent.Id
			scheduleEvent.Id = bson.ObjectId(newScheduleEventId)

			errAddSE := addScheduleEvent(scheduleEvent)
			if errAddSE != nil {
				LoggingClient.Error("error loading schedule event %s into scheduler", "message", errAddSE.Error())
				return errAddSE
			}
		} else {
			LoggingClient.Debug("Did not load schedule event as it exists in the scheduler", "event", scheduleEvent.Name)
		}
	}

	return nil
}

// Query core-metadata information
func loadCoreMetadataInformation() error {

	receivedSchedules, err := getMetadataSchedules()
	if err != nil {
		LoggingClient.Error("Failed to receive schedules from core-metadata", "message", err.Error())
		return err
	}

	err = addReceivedSchedules(receivedSchedules)
	if err != nil {
		LoggingClient.Error("Failed to add received schedules from core-metadata", "message", err.Error())
		return err
	}

	receivedScheduleEvents, err := getMetadataScheduleEvents()
	if err != nil {
		LoggingClient.Error("Failed to receive schedule events from core-metadata", "message", err.Error())
		return err
	}

	err = addReceivedScheduleEvents(receivedScheduleEvents)
	if err != nil {
		LoggingClient.Error("Failed to add received schedule events from core-metadata", "message", err.Error())
		return err
	}

	return nil
}
