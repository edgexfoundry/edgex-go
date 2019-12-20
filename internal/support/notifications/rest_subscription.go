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
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/operators/subscription"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/gorilla/mux"
)

func restGetSubscriptions(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	subscriptions, err := dbClient.GetSubscriptions()
	if err != nil {
		lc.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	pkg.Encode(subscriptions, w, lc)
}

func restAddSubscription(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	var s models.Subscription
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&s)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		lc.Error("Error decoding subscription: " + err.Error())
		return
	}

	// validate email addresses
	err = validateEmailAddresses(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		lc.Error(err.Error())
		return
	}

	lc.Info("Posting Subscription: " + s.String())
	op := subscription.NewAddExecutor(dbClient, s)
	err = op.Execute()
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		lc.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(s.Slug))
}

func restUpdateSubscription(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]

	var s models.Subscription
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&s)

	// validate email addresses
	err = validateEmailAddresses(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		lc.Error(err.Error())
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
		lc.Error(err.Error())
		return
	} else {
		s.ID = s2.ID
	}

	lc.Info("Updating subscription by slug: " + slug)

	if err = dbClient.UpdateSubscription(s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		lc.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restGetSubscriptionByID(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	id := vars["id"]
	op := subscription.NewIdExecutor(dbClient, id)
	s, err := op.Execute()
	if err != nil {
		lc.Error(err.Error())
		switch err.(type) {
		case errors.ErrSubscriptionNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	pkg.Encode(s, w, lc)
}

func restDeleteSubscriptionByID(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	id := vars["id"]
	lc.Info("Deleting subscription: " + id)

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
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
}

func restGetSubscriptionBySlug(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]

	op := subscription.NewSlugExecutor(dbClient, slug)
	s, err := op.Execute()

	if err != nil {
		switch err.(type) {
		case errors.ErrSubscriptionNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		lc.Error(err.Error())
		return
	}

	pkg.Encode(s, w, lc)
}

func restDeleteSubscriptionBySlug(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	slug := vars["slug"]

	lc.Info("Deleting subscription by slug: " + slug)

	op := subscription.NewDeleteBySlugExecutor(dbClient, slug)
	err := op.Execute()
	if err != nil {
		switch err.(type) {
		case errors.ErrSubscriptionNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		lc.Error(err.Error())
		return
	}
	w.Header().Set(clients.ContentType, clients.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func restGetSubscriptionsByCategories(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

	if r.Body != nil {
		defer r.Body.Close()
	}

	vars := mux.Vars(r)
	categories := splitVars(vars["categories"])
	op := subscription.NewCategoriesExecutor(dbClient, categories)
	s, err := op.Execute()
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "Subscription not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		lc.Error(err.Error())
		return
	}

	pkg.Encode(s, w, lc)
}

func subscriptionsByLabelsHandler(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

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
		lc.Error(err.Error())
		return
	}

	pkg.Encode(s, w, lc)

}

func subscriptionsByCategoriesLabelsHandler(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

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
		lc.Error(err.Error())
		return
	}

	pkg.Encode(s, w, lc)

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
		resp := "Addresses contain invalid CRLF characters"
		return errors.NewErrInvalidEmailAddresses(invalidAddrs, resp)
	}
	return nil
}

func subscriptionsByReceiverHandler(
	w http.ResponseWriter,
	r *http.Request,
	lc logger.LoggingClient,
	dbClient interfaces.DBClient) {

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
		lc.Error(err.Error())
		return
	}
	pkg.Encode(s, w, lc)

}
