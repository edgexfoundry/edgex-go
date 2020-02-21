/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *******************************************************************************/

package notifications

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	notificationsConfig "github.com/edgexfoundry/edgex-go/internal/support/notifications/config"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/operators/notification"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

func notificationHandler(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	var n models.Notification
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&n)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		lc.Error("Error decoding notification: " + err.Error())
		return
	}

	lc.Info("Posting Notification: " + n.String())
	n.Status = models.NotificationsStatus(models.New)
	n.ID, err = dbClient.AddNotification(n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		lc.Error(err.Error())
		return
	}

	lc.Debug("The scheduler is triggered for: " + n.Slug)
	n, err = dbClient.GetNotificationById(n.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error(err.Error())
		return
	}

	err = distributeAndMark(n, lc, dbClient, config)
	if err != nil {
		return
	}
	lc.Debug("The scheduler has completed for: " + n.Slug)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(n.ID))

}

func restGetNotificationBySlug(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]

	op := notification.NewSlugExecutor(dbClient, slug)
	result, err := op.Execute()
	if err != nil {
		lc.Error(err.Error())
		switch err.(type) {
		case errors.ErrNotificationNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:

			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	pkg.Encode(result, w, lc)
}

func restDeleteNotificationBySlug(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]

	lc.Info("Deleting notification (and associated transmissions) by slug: " + slug)

	op := notification.NewDeleteBySlugExecutor(dbClient, slug)
	err := op.Execute()
	if err != nil {
		lc.Error(err.Error())
		switch err.(type) {
		case errors.ErrNotificationNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:

			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restGetNotificationByID(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	var id string = vars["id"]
	op := notification.NewIdExecutor(dbClient, id)
	result, err := op.Execute()
	if err != nil {
		lc.Error(err.Error())
		switch err.(type) {
		case errors.ErrNotificationNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:

			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	pkg.Encode(result, w, lc)
}

func restDeleteNotificationByID(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	id := vars["id"]

	lc.Info("Deleting notification (and associated transmissions): " + id)

	op := notification.NewDeleteByIDExecutor(dbClient, id)
	err := op.Execute()

	if err != nil {
		lc.Error(err.Error())
		switch err.(type) {
		case errors.ErrNotificationNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restDeleteNotificationsByAge(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	age, err := strconv.Atoi(vars["age"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting the age to an integer")
		return
	}
	lc.Info("Deleting old notifications (and associated transmissions): " + vars["age"])
	op := notification.NewDeleteByAgeExecutor(dbClient, age)
	err = op.Execute()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restGetNotificationsBySender(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting limit to integer: " + err.Error())
		return
	}

	// Check the length
	if err = checkMaxLimit(limitNum, lc, config); err != nil {
		http.Error(w, ExceededMaxResultCount, http.StatusRequestEntityTooLarge)
		return
	}

	op := notification.NewSenderExecutor(dbClient, vars["sender"], limitNum)
	results, err := op.Execute()
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Notification not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		lc.Error(err.Error())
		return
	}

	pkg.Encode(results, w, lc)
}

func restNotificationByStartEnd(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	start, err := strconv.ParseInt(vars["start"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting the start to an integer")
		return
	}
	end, err := strconv.ParseInt(vars["end"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting the end to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting limit to integer: " + err.Error())
		return
	}

	// Check the length
	if err = checkMaxLimit(limitNum, lc, config); err != nil {
		http.Error(w, ExceededMaxResultCount, http.StatusRequestEntityTooLarge)
		return
	}

	op := notification.NewStartEndExecutor(dbClient, start, end, limitNum)
	results, err := op.Execute()
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Notification not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		lc.Error(err.Error())
		return
	}

	pkg.Encode(results, w, lc)
}

func restNotificationByStart(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) {

	if r.Body != nil {
		defer r.Body.Close()
	}
	vars := mux.Vars(r)
	start, err := strconv.ParseInt(vars["start"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting the start to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting limit to integer: " + err.Error())
		return
	}

	// Check the length
	if err = checkMaxLimit(limitNum, lc, config); err != nil {
		http.Error(w, ExceededMaxResultCount, http.StatusRequestEntityTooLarge)
		return
	}

	op := notification.NewStartExecutor(dbClient, start, limitNum)
	results, err := op.Execute()
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Notification not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		lc.Error(err.Error())
		return
	}

	pkg.Encode(results, w, lc)
}

func restNotificationByEnd(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	end, err := strconv.ParseInt(vars["end"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting the end to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting limit to integer: " + err.Error())
		return
	}

	// Check the length
	if err = checkMaxLimit(limitNum, lc, config); err != nil {
		http.Error(w, ExceededMaxResultCount, http.StatusRequestEntityTooLarge)
		return
	}

	op := notification.NewEndExecutor(dbClient, end, limitNum)
	results, err := op.Execute()
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Notification not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		lc.Error(err.Error())
		return
	}

	pkg.Encode(results, w, lc)
}

func restNotificationsByLabels(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting limit to integer: " + err.Error())
		return
	}

	// Check the length
	if err = checkMaxLimit(limitNum, lc, config); err != nil {
		http.Error(w, ExceededMaxResultCount, http.StatusRequestEntityTooLarge)
		return
	}

	labels := splitVars(vars["labels"])

	op := notification.NewLabelsExecutor(dbClient, labels, limitNum)
	results, err := op.Execute()
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Notification not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		lc.Error(err.Error())
		return
	}

	pkg.Encode(results, w, lc)
}

func restNotificationsNew(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient,
	config notificationsConfig.ConfigurationStruct) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error("Error converting limit to integer: " + err.Error())
		return
	}

	// Check the length
	if err = checkMaxLimit(limitNum, lc, config); err != nil {
		http.Error(w, ExceededMaxResultCount, http.StatusRequestEntityTooLarge)
		return
	}

	op := notification.NewGetNewestExecutor(dbClient, limitNum)
	n, err := op.Execute()
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Notification not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		lc.Error(err.Error())
		return
	}

	pkg.Encode(n, w, lc)
}
