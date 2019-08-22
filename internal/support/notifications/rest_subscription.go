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

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/errors"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/operators/subscription"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

const (
	applicationJson = "application/json; charset=utf-8"
)

func subscriptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]

	switch r.Method {

	// Get all subscriptions
	case http.MethodGet:
		subscriptions, err := dbClient.GetSubscriptions()
		if err != nil {
			LoggingClient.Error(err.Error())
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		pkg.Encode(subscriptions, w, LoggingClient)

		// Modify (an existing) subscription
	case http.MethodPut:
		var s models.Subscription
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&s)

		// validate email addresses
		err = validateEmailAddresses(s)
		if err != nil {
			switch err.(type) {
			case errors.ErrInvalidEmailAddresses:
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		// Check if the subscription exists
		s2, err := dbClient.GetSubscriptionBySlug(s.Slug)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		} else {
			s.ID = s2.ID
		}

		LoggingClient.Info("Updating subscription by slug: " + slug)

		if err = dbClient.UpdateSubscription(s); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("true"))

	case http.MethodPost:
		var s models.Subscription
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&s)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			LoggingClient.Error("Error decoding subscription: " + err.Error())
			return
		}

		// validate email addresses
		err = validateEmailAddresses(s)
		if err != nil {
			switch err.(type) {
			case errors.ErrInvalidEmailAddresses:
				http.Error(w, err.Error(), http.StatusBadRequest)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		LoggingClient.Info("Posting Subscription: " + s.String())
		_, err = dbClient.AddSubscription(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			LoggingClient.Error(err.Error())
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(s.Slug))

	}
}

func restGetSubscriptionByID(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	id := vars["id"]
	op := subscription.NewIdExecutor(dbClient, id)
	s, err := op.Execute()
	if err != nil {
		LoggingClient.Error(err.Error())
		switch err.(type) {
		case errors.ErrSubscriptionNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	pkg.Encode(s, w, LoggingClient)
}

func restDeleteSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	id := vars["id"]
	LoggingClient.Info("Deleting subscription: " + id)

	op := subscription.NewDeleteByIDExecutor(dbClient, id)
	err := op.Execute()
	if err != nil {
		switch err.(type) {
		case errors.ErrSubscriptionNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func subscriptionsBySlugHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]
	switch r.Method {
	case http.MethodGet:

		s, err := dbClient.GetSubscriptionBySlug(slug)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			pkg.Encode(s, w, LoggingClient)
			return
		}

		pkg.Encode(s, w, LoggingClient)
	case http.MethodDelete:
		_, err := dbClient.GetSubscriptionBySlug(slug)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "Subscription not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			LoggingClient.Error(err.Error())
			return
		}

		LoggingClient.Info("Deleting subscription by slug: " + slug)

		if err = dbClient.DeleteSubscriptionBySlug(slug); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			LoggingClient.Error(err.Error())
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

	categories := splitVars(vars["categories"])

	s, err := dbClient.GetSubscriptionByCategories(categories)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Subscription not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	pkg.Encode(s, w, LoggingClient)

}

func subscriptionsByLabelsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)

	labels := splitVars(vars["labels"])

	s, err := dbClient.GetSubscriptionByLabels(labels)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Subscription not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	pkg.Encode(s, w, LoggingClient)

}

func subscriptionsByCategoriesLabelsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)

	labels := splitVars(vars["labels"])
	categories := splitVars(vars["categories"])

	s, err := dbClient.GetSubscriptionByCategoriesLabels(categories, labels)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Subscription not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}

	pkg.Encode(s, w, LoggingClient)

}

func splitVars(vars string) []string {
	return strings.Split(vars, ",")
}

func validateEmailAddresses(s models.Subscription) error {
	var invalidAddrs []string
	for _, c := range s.Channels {
		if c.Type == models.ChannelType(models.Email) {
			for _, m := range c.MailAddresses {
				if strings.ContainsAny(m, "\n\r") {
					invalidAddrs = append(invalidAddrs, m)
				}
			}
		}
	}
	if len(invalidAddrs) > 0 {
		resp := "Subscription " + s.Slug + " mail addresses contain CRLF: ["
		for _, m := range invalidAddrs {
			resp += m + ", "
		}
		resp = strings.TrimSuffix(resp, ", ")
		resp += "]"
		return errors.NewErrInvalidEmailAddresses(resp)
	}
	return nil
}

func subscriptionsByReceiverHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)

	s, err := dbClient.GetSubscriptionByReceiver(vars["receiver"])
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Subscription not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		LoggingClient.Error(err.Error())
		return
	}
	pkg.Encode(s, w, LoggingClient)

}
