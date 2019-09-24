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
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
)

func LoadRestRoutes() *mux.Router {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, configHandler).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, metricsHandler).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	b := r.PathPrefix(clients.ApiBase).Subrouter()

	// Notifications
	b.HandleFunc("/"+NOTIFICATION, notificationHandler).Methods(http.MethodPost)
	b.HandleFunc("/"+NOTIFICATION+"/{"+ID+"}", restGetNotificationByID).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/{"+ID+"}", restDeleteNotificationByID).Methods(http.MethodDelete)
	b.HandleFunc("/"+NOTIFICATION+"/"+SLUG+"/{"+SLUG+"}", restGetNotificationBySlug).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+SLUG+"/{"+SLUG+"}", restDeleteNotificationBySlug).Methods(http.MethodDelete)
	b.HandleFunc("/"+NOTIFICATION+"/"+AGE+"/{"+AGE+":[0-9]+}", restDeleteNotificationsByAge).Methods(http.MethodDelete)
	b.HandleFunc("/"+NOTIFICATION+"/"+SENDER+"/{"+SENDER+"}/{"+LIMIT+":[0-9]+}", restGetNotificationsBySender).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+START+"/{"+START+"}/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}", restNotificationByStartEnd).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+START+"/{"+START+"}/{"+LIMIT+":[0-9]+}", restNotificationByStart).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}", restNotificationByEnd).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+LABELS+"/{"+LABELS+"}/{"+LIMIT+":[0-9]+}", restNotificationsByLabels).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+NEW+"/{"+LIMIT+":[0-9]+}", restNotificationsNew).Methods(http.MethodGet)

	// GetSubscriptions
	b.HandleFunc("/"+SUBSCRIPTION, restGetSubscriptions).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION, subscriptionHandler).Methods(http.MethodPut, http.MethodPost)
	b.HandleFunc("/"+SUBSCRIPTION+"/{"+ID+"}", restGetSubscriptionByID).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION+"/{"+ID+"}", restDeleteSubscriptionByID).Methods(http.MethodDelete)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+SLUG+"/{"+SLUG+"}", restGetSubscriptionBySlug).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+SLUG+"/{"+SLUG+"}", restDeleteSubscriptionBySlug).Methods(http.MethodDelete)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+CATEGORIES+"/{"+CATEGORIES+"}/"+LABELS+"/{"+LABELS+"}", subscriptionsByCategoriesLabelsHandler).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+CATEGORIES+"/{"+CATEGORIES+"}", restGetSubscriptionsByCategories).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+LABELS+"/{"+LABELS+"}", subscriptionsByLabelsHandler).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+RECEIVER+"/{"+RECEIVER+"}", subscriptionsByReceiverHandler).Methods(http.MethodGet)

	// Transmissions
	b.HandleFunc("/"+TRANSMISSION, transmissionHandler).Methods(http.MethodPost)
	b.HandleFunc("/"+TRANSMISSION+"/"+SLUG+"/{"+SLUG+"}/{"+LIMIT+":[0-9]+}", transmissionBySlugHandler).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+SLUG+"/{"+SLUG+"}/"+START+"/{"+START+"}/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}", transmissionBySlugAndStartEndHandler).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+START+"/{"+START+"}/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}", transmissionByStartEndHandler).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+START+"/{"+START+"}/{"+LIMIT+":[0-9]+}", transmissionByStartHandler).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}", transmissionByEndHandler).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+ESCALATED+"/{"+LIMIT+":[0-9]+}", transmissionByEscalatedHandler).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+FAILED+"/{"+LIMIT+":[0-9]+}", transmissionByFailedHandler).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+SENT+"/"+AGE+"/{"+AGE+":[0-9]+}", transmissionByAgeSentHandler).Methods(http.MethodDelete)
	b.HandleFunc("/"+TRANSMISSION+"/"+ESCALATED+"/"+AGE+"/{"+AGE+":[0-9]+}", transmissionByAgeEscalatedHandler).Methods(http.MethodDelete)
	b.HandleFunc("/"+TRANSMISSION+"/"+ACKNOWLEDGED+"/"+AGE+"/{"+AGE+":[0-9]+}", transmissionByAgeAcknowledgedHandler).Methods(http.MethodDelete)
	b.HandleFunc("/"+TRANSMISSION+"/"+FAILED+"/"+AGE+"/{"+AGE+":[0-9]+}", transmissionByAgeFailedHandler).Methods(http.MethodDelete)

	// Cleanup
	b.HandleFunc("/"+CLEANUP, cleanupHandler).Methods(http.MethodDelete)
	b.HandleFunc("/"+CLEANUP+"/"+AGE+"/{"+AGE+":[0-9]+}", cleanupAgeHandler).Methods(http.MethodDelete)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}

func configHandler(w http.ResponseWriter, _ *http.Request) {
	pkg.Encode(Configuration, w, LoggingClient)
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	s := telemetry.NewSystemUsage()

	pkg.Encode(s, w, LoggingClient)

	return
}
