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

	"github.com/edgexfoundry/edgex-go/support/notifications/clients"
	"github.com/edgexfoundry/edgex-go/support/notifications/models"
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
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Error decoding notification: " + err.Error())
			return
		}

		loggingClient.Info("Posting Notification: " + n.String())
		n.Status = models.NotificationsStatus(models.New)
		id, err := dbc.AddNotification(&n)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		if n.Severity == models.NotificationsSeverity(models.Critical) {
			loggingClient.Info("Critical severity scheduler is triggered for: " + n.Slug)
			err := distributeAndMark(n)
			if err != nil {
				return
			}
			loggingClient.Info("Critical severity scheduler has completed for: " + n.Slug)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id.Hex()))

		break
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

		n, err := dbc.NotificationBySlug(slug)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(n, w)
	case http.MethodDelete:
		_, err := dbc.NotificationBySlug(slug)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		loggingClient.Info("Deleting notification (and associated transmissions) by slug: " + slug)

		if err = dbc.DeleteNotificationBySlug(slug); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
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

		n, err := dbc.NotificationById(id)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(n, w)
	case http.MethodDelete:
		_, err := dbc.NotificationById(id)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		loggingClient.Info("Deleting notification (and associated transmissions): " + id)

		if err = dbc.DeleteNotificationById(id); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
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
	age, err := strconv.ParseInt(vars["age"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the age to an integer")
		return
	}
	switch r.Method {
	case http.MethodDelete:
		loggingClient.Info("Deleting old notifications (and associated transmissions): " + vars["age"])
		err := dbc.DeleteNotificationsOld(age)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Notifications not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		if limitNum > configuration.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			loggingClient.Error("Exceeded max limit")
			return
		}

		n, err := dbc.NotificationBySender(vars["sender"], limitNum)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the start to an integer")
		return
	}
	end, err := strconv.ParseInt(vars["end"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the end to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		if limitNum > configuration.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			loggingClient.Error("Exceeded max limit")
			return
		}

		n, err := dbc.NotificationsByStartEnd(start, end, limitNum)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(n, w)
	}
}

func notificationByStartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	vars := mux.Vars(r)
	start, err := strconv.ParseInt(vars["start"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the start to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}
	switch r.Method {
	case http.MethodGet:

		if limitNum > configuration.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			loggingClient.Error("Exceeded max limit")
			return
		}

		n, err := dbc.NotificationsByStart(start, limitNum)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting the end to an integer")
		return
	}
	limitNum, err := strconv.Atoi(vars["limit"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		if limitNum > configuration.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			loggingClient.Error("Exceeded max limit")
			return
		}

		n, err := dbc.NotificationsByEnd(end, limitNum)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		if limitNum > configuration.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			loggingClient.Error("Exceeded max limit")
			return
		}

		labels := splitVars(vars["labels"])

		n, err := dbc.NotificationsByLabels(labels, limitNum)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
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
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		loggingClient.Error("Error converting limit to integer: " + err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:

		if limitNum > configuration.ReadMaxLimit {
			http.Error(w, "Exceeded max limit", http.StatusRequestEntityTooLarge)
			loggingClient.Error("Exceeded max limit")
			return
		}

		n, err := dbc.NotificationsNew(limitNum)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Notification not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(n, w)
	}
}
