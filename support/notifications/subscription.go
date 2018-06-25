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
	"strings"

	"github.com/edgexfoundry/edgex-go/support/notifications/clients"
	"github.com/edgexfoundry/edgex-go/support/notifications/models"
	"github.com/gorilla/mux"
)

func subscriptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	switch r.Method {
	case http.MethodPost:
		var s models.Subscription
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&s)

		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error("Error decoding subscription: " + err.Error())
			return
		}

		loggingClient.Info("Posting Subscription: " + s.String())
		id, err := dbc.AddSubscription(&s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(id.Hex()))

		break
	}
}

func subscriptionByIDHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:

		s, err := dbc.SubscriptionById(vars["id"])
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(s, w)
	}
}

func subscriptionsBySlugHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]
	switch r.Method {
	case http.MethodGet:

		s, err := dbc.SubscriptionBySlug(slug)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(s, w)
	case http.MethodDelete:
		_, err := dbc.SubscriptionBySlug(slug)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		loggingClient.Info("Deleting subscription by slug: " + slug)

		if err = dbc.DeleteSubscriptionBySlug(slug); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			loggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))
	}
}

func subscriptionsByCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:

		categories := splitVars(vars["categories"])

		s, err := dbc.SubscriptionByCategories(categories)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(s, w)
	}
}

func subscriptionsByLabelsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:

		labels := splitVars(vars["labels"])

		s, err := dbc.SubscriptionByLabels(labels)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(s, w)
	}
}

func subscriptionsByCategoriesLabelsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:

		labels := splitVars(vars["labels"])
		categories := splitVars(vars["categories"])

		s, err := dbc.SubscriptionByCategoriesLabels(categories, labels)
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(s, w)
	}
}

func splitVars(vars string) []string {
	return strings.Split(vars, ",")
}

func subscriptionsByReceiverHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:

		s, err := dbc.SubscriptionByReceiver(vars["receiver"])
		if err != nil {
			if err == clients.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			}
			loggingClient.Error(err.Error())
			return
		}

		encode(s, w)
	}
}
