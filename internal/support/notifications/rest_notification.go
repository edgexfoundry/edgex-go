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

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

func notificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	switch r.Method {
	case http.MethodPost:
		var n models.Notification
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&n)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			LoggingClient.Error("Error decoding notification: " + err.Error())
			return
		}

		LoggingClient.Info("Posting Notification: " + n.String())
		n.Status = models.NotificationsStatus(models.New)
		n.ID, err = dbClient.AddNotification(n)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			LoggingClient.Error(err.Error())
			return
		}

		if n.Severity == models.NotificationsSeverity(models.Critical) {
			LoggingClient.Info("Critical severity scheduler is triggered for: " + n.Slug)
			err := distributeAndMark(n)
			if err != nil {
				return
			}
			LoggingClient.Info("Critical severity scheduler has completed for: " + n.Slug)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(n.Slug))
	}
}

func notificationBySlugHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]
	switch r.Method {
	case http.MethodGet:

		n, err := dbClient.GetNotificationBySlug(slug)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			w.Header().Set("Content-Type", applicationJson)
			encode(n, w)
			return
		}

		encode(n, w)
	case http.MethodDelete:
		_, err := dbClient.GetNotificationBySlug(slug)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		LoggingClient.Info("Deleting notification (and associated transmissions) by slug: " + slug)

		if err = dbClient.DeleteNotificationBySlug(slug); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
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

		n, err := dbClient.GetNotificationById(id)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			w.Header().Set("Content-Type", applicationJson)
			encode(n, w)
			return
		}

		encode(n, w)
	case http.MethodDelete:
		_, err := dbClient.GetNotificationById(id)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		LoggingClient.Info("Deleting notification (and associated transmissions): " + id)

		if err = dbClient.DeleteNotificationById(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
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
	switch r.Method {
	case http.MethodDelete:
		LoggingClient.Info("Deleting old notifications (and associated transmissions): " + vars["age"])
		err := dbClient.DeleteNotificationsOld(age)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Notifications not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
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

	switch r.Method {
	case http.MethodGet:

		if limitNum > Configuration.Service.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			LoggingClient.Error("Exceeded max limit")
			return
		}

		n, err := dbClient.GetNotificationBySender(vars["sender"], limitNum)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			w.Header().Set("Content-Type", applicationJson)
			encode(n, w)
			return
		}

		encode(n, w)
	}
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

	switch r.Method {
	case http.MethodGet:

		if limitNum > Configuration.Service.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			LoggingClient.Error("Exceeded max limit")
			return
		}

		n, err := dbClient.GetNotificationsByStartEnd(start, end, limitNum)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}
		encodeWithUTF8(n, w)
	}
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
	switch r.Method {
	case http.MethodGet:

		if limitNum > Configuration.Service.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			LoggingClient.Error("Exceeded max limit")
			return
		}

		n, err := dbClient.GetNotificationsByStart(start, limitNum)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			w.Header().Set("Content-Type", applicationJson)
			encode(n, w)
			return
		}

		encode(n, w)
	}
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

	switch r.Method {
	case http.MethodGet:

		if limitNum > Configuration.Service.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			LoggingClient.Error("Exceeded max limit")
			return
		}

		n, err := dbClient.GetNotificationsByEnd(end, limitNum)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			w.Header().Set("Content-Type", applicationJson)
			encode(n, w)
			return
		}

		encode(n, w)
	}
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

	switch r.Method {
	case http.MethodGet:

		if limitNum > Configuration.Service.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			LoggingClient.Error("Exceeded max limit")
			return
		}

		labels := splitVars(vars["labels"])

		n, err := dbClient.GetNotificationsByLabels(labels, limitNum)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			w.Header().Set("Content-Type", applicationJson)
			encode(n, w)
			return
		}

		encode(n, w)
	}
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

	switch r.Method {
	case http.MethodGet:

		if limitNum > Configuration.Service.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			LoggingClient.Error("Exceeded max limit")
			return
		}

		n, err := dbClient.GetNewNotifications(limitNum)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			w.Header().Set("Content-Type", applicationJson)
			encode(n, w)
			return
		}

		encode(n, w)
	}
}
