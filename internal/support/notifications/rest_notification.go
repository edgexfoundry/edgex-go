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
	"fmt"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

func addNotification(n *models.Notification) (e error) {

	var err error
	n.Status = models.NotificationsStatus(models.New)
	LoggingClient.Info("Posting Notification: " + n.String())
	n.ID, err = dbClient.AddNotification(*n)
	if err != nil {
		switch err {
		case db.ErrNotUnique:
			newErr := errors.NewErrNotificationInUse(n.Slug)
			LoggingClient.Error(newErr.Error(), "error message", err.Error())
			return newErr
		default:
			LoggingClient.Error(err.Error())
			return err
		}
	}
	return
}

func checkSeverity(n *models.Notification) (err error) {

	if n.Severity == models.NotificationsSeverity(models.Critical) {

		LoggingClient.Info("Critical severity scheduler is triggered for: " + n.Slug)
		err := distributeAndMark(*n)
		if err != nil {
			LoggingClient.Error(err.Error())
			return err
		}
		LoggingClient.Info("Critical severity scheduler has completed for: " + n.Slug)

	}
	return
}

func getNotificationBySlug(slug string) (n models.Notification, err error) {
	n, err = dbClient.GetNotificationBySlug(slug)
	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			err = errors.NewErrNotificationNotFound(slug)
		}
		return n, err
	}
	return n, nil
}

func getNotificationByID(id string) (n models.Notification, err error) {
	n, err = dbClient.GetNotificationById(id)
	if err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			err = errors.NewErrNotificationNotFound(id)
		}
		return n, err
	}
	return n, nil
}

func deleteNotificationBySlug(slug string) (err error) {
	LoggingClient.Info("Deleting notification (and associated transmissions) by slug: " + slug)

	if err = dbClient.DeleteNotificationBySlug(slug); err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			err = errors.NewErrNotificationNotFound(slug)
		}
		return err
	}
	return
}

func deleteNotificationByID(slug string) (err error) {
	LoggingClient.Info("Deleting notification (and associated transmissions) by slug: " + slug)

	if err = dbClient.DeleteNotificationById(slug); err != nil {
		LoggingClient.Error(err.Error())
		if err == db.ErrNotFound {
			err = errors.NewErrNotificationNotFound(slug)
		}
		return err
	}
	return
}

func deleteNotificationsOld(age int) (err error) {
	LoggingClient.Info("Deleting old notifications (and associated transmissions): " + string(age))
	err = dbClient.DeleteNotificationsOld(age)
	if err != nil {
		LoggingClient.Error(err.Error())
		return err
	}
	return
}

func getNotificationsBySender(sender string, limit int) (n []models.Notification, err error) {
	n, err = dbClient.GetNotificationBySender(sender, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return n, err
	}
	return n, nil
}

func getNotificationsByStartEnd(start int64, end int64, limit int) (n []models.Notification, err error) {
	n, err = dbClient.GetNotificationsByStartEnd(start, end, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return n, err
	}
	return n, nil
}

func getNotificationsByStart(start int64, limit int) (n []models.Notification, err error) {
	n, err = dbClient.GetNotificationsByStart(start, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return n, err
	}
	return n, nil
}

func getNotificationsByEnd(end int64, limit int) (n []models.Notification, err error) {
	n, err = dbClient.GetNotificationsByEnd(end, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return n, err
	}
	return n, nil
}

func getNotificationsByLabels(labels []string, limit int) (n []models.Notification, err error) {
	n, err = dbClient.GetNotificationsByLabels(labels, limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return n, err
	}
	return n, nil
}

func getNewNotifications(limit int) (n []models.Notification, err error) {
	n, err = dbClient.GetNewNotifications(limit)
	if err != nil {
		LoggingClient.Error(err.Error())
		return n, err
	}
	return n, nil
}

func notificationHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	var n models.Notification
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&n)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		LoggingClient.Error("Error decoding notification: " + err.Error())
		return
	}

	err = addNotification(&n)
	if err != nil {
		switch err.(type) {
		case errors.ErrNotificationInUse:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	err = checkSeverity(&n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	fmt.Println(n.ID)
	w.Write([]byte(n.ID))

}

func notificationBySlugHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]
	switch r.Method {
	case http.MethodGet:
		n, err := getNotificationBySlug(slug)
		if err != nil {
			switch err.(type) {
			case errors.ErrNotificationNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		pkg.Encode(n, w, LoggingClient)
	case http.MethodDelete:

		if err := deleteNotificationBySlug(slug); err != nil {
			switch err.(type) {
			case errors.ErrNotificationNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}

func notificationByIDHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	id := vars["id"]
	switch r.Method {
	case http.MethodGet:

		n, err := getNotificationByID(id)
		if err != nil {
			switch err.(type) {
			case errors.ErrNotificationNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		pkg.Encode(n, w, LoggingClient)

	case http.MethodDelete:
		LoggingClient.Info("Deleting notification (and associated transmissions): " + id)

		if err := deleteNotificationByID(id); err != nil {
			switch err.(type) {
			case errors.ErrNotificationNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}

func notificationOldHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	age, err := strconv.Atoi(vars["age"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting the age to an integer")
		return
	}
	err = deleteNotificationsOld(age)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))

}

func notificationBySenderHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	if limitNum > Configuration.Service.MaxResultCount {
		http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
		LoggingClient.Error("Exceeded max limit")
		return
	}

	n, err := getNotificationsBySender(vars["sender"], limitNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg.Encode(n, w, LoggingClient)

}

func notificationByStartEndHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	start, err := strconv.ParseInt(vars["start"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting the start to an integer")
		return
	}
	end, err := strconv.ParseInt(vars["end"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting the end to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	if limitNum > Configuration.Service.MaxResultCount {
		http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
		LoggingClient.Error("Exceeded max limit")
		return
	}

	n, err := getNotificationsByStartEnd(start, end, limitNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkg.Encode(n, w, LoggingClient)

}

func notificationByStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	vars := mux.Vars(r)
	start, err := strconv.ParseInt(vars["start"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting the start to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	if limitNum > Configuration.Service.MaxResultCount {
		http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
		LoggingClient.Error("Exceeded max limit")
		return
	}

	n, err := getNotificationsByStart(start, limitNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkg.Encode(n, w, LoggingClient)

}

func notificationByEndHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	end, err := strconv.ParseInt(vars["end"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting the end to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	if limitNum > Configuration.Service.MaxResultCount {
		http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
		LoggingClient.Error("Exceeded max limit")
		return
	}

	n, err := getNotificationsByEnd(end, limitNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkg.Encode(n, w, LoggingClient)

}

func notificationsByLabelsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	if limitNum > Configuration.Service.MaxResultCount {
		http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
		LoggingClient.Error("Exceeded max limit")
		return
	}

	labels := splitVars(vars["labels"])

	n, err := getNotificationsByLabels(labels, limitNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkg.Encode(n, w, LoggingClient)

}

func notificationsNewHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	if limitNum > Configuration.Service.MaxResultCount {
		http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
		LoggingClient.Error("Exceeded max limit")
		return
	}

	n, err := getNewNotifications(limitNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pkg.Encode(n, w, LoggingClient)

}
