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
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/mux"
)

func getSubscriptions() (subscriptions []models.Subscription, err error) {
	subscriptions, err = dbClient.GetSubscriptions()
	if err != nil {
		LoggingClient.Error(err.Error())
		return subscriptions, err
	}
	return subscriptions, nil
}

func updateSubscription(s models.Subscription) error {
	LoggingClient.Info("Updating subscription by slug: " + s.Slug)
	if err := dbClient.UpdateSubscription(s); err != nil {
		LoggingClient.Error(err.Error())
		return err
	}
	return nil
}

func addSubscription(s models.Subscription) error {
	LoggingClient.Info("Posting Subscription: " + s.Slug)
	_, err := dbClient.AddSubscription(s)
	if err != nil {
		switch err {
		case db.ErrNotUnique:
			newErr := errors.NewErrNotificationInUse(s.Slug)
			LoggingClient.Error(newErr.Error(), "error message", err.Error())
			return newErr
		default:
			LoggingClient.Error(err.Error())
			return err
		}
	}
	return nil
}

func getSubscriptionByReceiver(receiver string) ([]models.Subscription, error) {
	s, err := dbClient.GetSubscriptionByReceiver(receiver)
	if err != nil {
		LoggingClient.Error(err.Error())
		return s, err
	}
	return s, nil
}

func getSubscriptionByID(id string) (s models.Subscription, err error) {
	s, err = dbClient.GetSubscriptionById(id)
	if err != nil {
		LoggingClient.Error(err.Error())
		switch err {
		case db.ErrNotFound:
			newErr := errors.NewErrSubscriptionNotFound(id)
			return s, newErr
		default:
			return s, err
		}
	}
	return s, nil
}

func deleteSubscriptionByID(id string) error {
	LoggingClient.Info("Deleting subscription: " + id)
	err := dbClient.DeleteSubscriptionById(id)
	if err != nil {
		LoggingClient.Error(err.Error())
		switch err {
		case db.ErrNotFound:
			newErr := errors.NewErrSubscriptionNotFound(id)
			return newErr
		default:
			return err
		}
	}
	return nil
}

func getSubscriptionBySlug(slug string) (s models.Subscription, err error) {
	s, err = dbClient.GetSubscriptionBySlug(slug)
	if err != nil {
		LoggingClient.Error(err.Error())
		switch err {
		case db.ErrNotFound:
			newErr := errors.NewErrSubscriptionNotFound(slug)
			return s, newErr
		default:
			return s, err
		}
	}
	return s, nil

}

func deleteSubscriptionBySlug(slug string) error {
	LoggingClient.Info("Deleting subscription by slug: " + slug)

	if err := dbClient.DeleteSubscriptionBySlug(slug); err != nil {
		LoggingClient.Error(err.Error())
		switch err {
		case db.ErrNotFound:
			newErr := errors.NewErrSubscriptionNotFound(slug)
			return newErr
		default:
			return err
		}
	}
	return nil
}

func getSubscriptionByCategories(categories []string) (s []models.Subscription, err error) {
	s, err = dbClient.GetSubscriptionByCategories(categories)
	if err != nil {
		LoggingClient.Error(err.Error())
		return s, err
	}
	return s, nil
}

func getSubscriptionByLabels(labels []string) (s []models.Subscription, err error) {
	s, err = dbClient.GetSubscriptionByLabels(labels)
	if err != nil {
		LoggingClient.Error(err.Error())
		return s, err
	}
	return s, nil
}

func getSubscriptionByCategoriesLabels(categories []string, labels []string) (s []models.Subscription, err error) {
	s, err = dbClient.GetSubscriptionByCategoriesLabels(categories, labels)
	if err != nil {
		LoggingClient.Error(err.Error())
		return s, err
	}
	return s, nil
}

func subscriptionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	switch r.Method {

	// Get all subscriptions
	case http.MethodGet:
		subscriptions, err := getSubscriptions()
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		pkg.Encode(subscriptions, w, LoggingClient)

		// Modify (an existing) subscription
	case http.MethodPut:
		var s models.Subscription
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&s)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			LoggingClient.Error("Error decoding subscription: " + err.Error())
			return
		}

		// Check if the subscription exists
		s2, err := getSubscriptionBySlug(s.Slug)
		if err != nil {
			switch err.(type) {
			case errors.ErrSubscriptionNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		} else {
			s.ID = s2.ID
		}

		if err = updateSubscription(s); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

		err = addSubscription(s)
		if err != nil {
			switch err.(type) {
			case errors.ErrNotificationInUse:
				http.Error(w, err.Error(), http.StatusConflict)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(s.Slug))

	}
}

func subscriptionByIDHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	id := vars["id"]
	switch r.Method {
	case http.MethodGet:

		s, err := getSubscriptionByID(id)
		if err != nil {
			switch err.(type) {
			case errors.ErrSubscriptionNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		pkg.Encode(s, w, LoggingClient)
		break
	case http.MethodDelete:

		if err := deleteSubscriptionByID(id); err != nil {
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
		w.Write([]byte("true"))
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

		s, err := getSubscriptionBySlug(slug)
		if err != nil {
			switch err.(type) {
			case errors.ErrSubscriptionNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		pkg.Encode(s, w, LoggingClient)

	case http.MethodDelete:

		LoggingClient.Info("Deleting subscription by slug: " + slug)

		if err := deleteSubscriptionBySlug(slug); err != nil {
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
		w.Write([]byte("true"))
	}
}

func subscriptionsByCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)

	categories := splitVars(vars["categories"])

	s, err := getSubscriptionByCategories(categories)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	s, err := getSubscriptionByLabels(labels)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	s, err := getSubscriptionByCategoriesLabels(categories, labels)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		LoggingClient.Error(err.Error())
		return
	}

	pkg.Encode(s, w, LoggingClient)

}

func splitVars(vars string) []string {
	return strings.Split(vars, ",")
}

func subscriptionsByReceiverHandler(w http.ResponseWriter, r *http.Request) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)

	s, err := getSubscriptionByReceiver(vars["receiver"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pkg.Encode(s, w, LoggingClient)

}
