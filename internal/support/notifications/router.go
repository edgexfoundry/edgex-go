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
	"github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	notificationsContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"

	"github.com/gorilla/mux"
)

func loadV1Routes(r *mux.Router, dic *di.Container) {
	// Ping Resource
	r.HandleFunc(
		clients.ApiPingRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(clients.ContentType, clients.ContentTypeText)
			w.Write([]byte("pong"))
		}).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(
		clients.ApiConfigRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			pkg.Encode(*notificationsContainer.ConfigurationFrom(dic.Get), w, bootstrapContainer.LoggingClientFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(
		clients.ApiMetricsRoute,
		func(w http.ResponseWriter, _ *http.Request) {
			pkg.Encode(telemetry.NewSystemUsage(), w, bootstrapContainer.LoggingClientFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	b := r.PathPrefix(clients.ApiBase).Subrouter()

	// Notifications
	b.HandleFunc(
		"/"+NOTIFICATION,
		func(w http.ResponseWriter, r *http.Request) {
			notificationHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				*notificationsContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPost)
	b.HandleFunc(
		"/"+NOTIFICATION+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetNotificationByID(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+NOTIFICATION+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteNotificationByID(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc(
		"/"+NOTIFICATION+"/"+SLUG+"/{"+SLUG+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetNotificationBySlug(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+NOTIFICATION+"/"+SLUG+"/{"+SLUG+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteNotificationBySlug(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc(
		"/"+NOTIFICATION+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteNotificationsByAge(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc(
		"/"+NOTIFICATION+"/"+SENDER+"/{"+SENDER+"}/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetNotificationsBySender(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				*notificationsContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+NOTIFICATION+"/"+START+"/{"+START+"}/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			restNotificationByStartEnd(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				*notificationsContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+NOTIFICATION+"/"+START+"/{"+START+"}/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			restNotificationByStart(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				*notificationsContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+NOTIFICATION+"/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			restNotificationByEnd(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				*notificationsContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+NOTIFICATION+"/"+LABELS+"/{"+LABELS+"}/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			restNotificationsByLabels(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				*notificationsContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+NOTIFICATION+"/"+NEW+"/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			restNotificationsNew(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get),
				*notificationsContainer.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)

	// GetSubscriptions
	b.HandleFunc(
		"/"+SUBSCRIPTION,
		func(w http.ResponseWriter, r *http.Request) {
			restGetSubscriptions(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+SUBSCRIPTION,
		func(w http.ResponseWriter, r *http.Request) {
			restAddSubscription(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodPost)
	b.HandleFunc(
		"/"+SUBSCRIPTION,
		func(w http.ResponseWriter, r *http.Request) {
			restUpdateSubscription(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodPut)
	b.HandleFunc(
		"/"+SUBSCRIPTION+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetSubscriptionByID(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+SUBSCRIPTION+"/{"+ID+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteSubscriptionByID(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc(
		"/"+SUBSCRIPTION+"/"+SLUG+"/{"+SLUG+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetSubscriptionBySlug(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+SUBSCRIPTION+"/"+SLUG+"/{"+SLUG+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restDeleteSubscriptionBySlug(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc(
		"/"+SUBSCRIPTION+"/"+CATEGORIES+"/{"+CATEGORIES+"}/"+LABELS+"/{"+LABELS+"}",
		func(w http.ResponseWriter, r *http.Request) {
			subscriptionsByCategoriesLabelsHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+SUBSCRIPTION+"/"+CATEGORIES+"/{"+CATEGORIES+"}",
		func(w http.ResponseWriter, r *http.Request) {
			restGetSubscriptionsByCategories(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+SUBSCRIPTION+"/"+LABELS+"/{"+LABELS+"}",
		func(w http.ResponseWriter, r *http.Request) {
			subscriptionsByLabelsHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+SUBSCRIPTION+"/"+RECEIVER+"/{"+RECEIVER+"}",
		func(w http.ResponseWriter, r *http.Request) {
			subscriptionsByReceiverHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Transmissions
	b.HandleFunc(
		"/"+TRANSMISSION,
		func(w http.ResponseWriter, r *http.Request) {
			transmissionHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodPost)
	b.HandleFunc(
		"/"+TRANSMISSION+"/"+SLUG+"/{"+SLUG+"}/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			transmissionBySlugHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+TRANSMISSION+"/"+SLUG+"/{"+SLUG+"}/"+START+"/{"+START+"}/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			transmissionBySlugAndStartEndHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+TRANSMISSION+"/"+START+"/{"+START+"}/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			transmissionByStartEndHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+TRANSMISSION+"/"+START+"/{"+START+"}/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			transmissionByStartHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+TRANSMISSION+"/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			transmissionByEndHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+TRANSMISSION+"/"+ESCALATED+"/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			transmissionByEscalatedHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+TRANSMISSION+"/"+FAILED+"/{"+LIMIT+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			transmissionByFailedHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc(
		"/"+TRANSMISSION+"/"+SENT+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			transmissionByAgeSentHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc(
		"/"+TRANSMISSION+"/"+ESCALATED+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			transmissionByAgeEscalatedHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc(
		"/"+TRANSMISSION+"/"+ACKNOWLEDGED+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			transmissionByAgeAcknowledgedHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc(
		"/"+TRANSMISSION+"/"+FAILED+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			transmissionByAgeFailedHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)

	// Cleanup
	b.HandleFunc(
		"/"+CLEANUP,
		func(w http.ResponseWriter, r *http.Request) {
			cleanupHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc(
		"/"+CLEANUP+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(w http.ResponseWriter, r *http.Request) {
			cleanupAgeHandler(
				w,
				r,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				container.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)
}
