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
	bootstrapContainer "github.com/edgexfoundry/edgex-go/internal/pkg/bootstrap/container"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	notificationsConfig "github.com/edgexfoundry/edgex-go/internal/support/notifications/config"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/container"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	"github.com/gorilla/mux"
)

func LoadRestRoutes(dic *di.Container) *mux.Router {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute,
		func(writer http.ResponseWriter, request *http.Request) {
			configHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				*container.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute,
		func(writer http.ResponseWriter, request *http.Request) {
			metricsHandler(writer, request, bootstrapContainer.LoggingClientFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Version
	r.HandleFunc(clients.ApiVersionRoute, pkg.VersionHandler).Methods(http.MethodGet)

	b := r.PathPrefix(clients.ApiBase).Subrouter()

	// Notifications
	b.HandleFunc("/"+NOTIFICATION,
		func(writer http.ResponseWriter, request *http.Request) {
			notificationHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get),
				*container.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodPost)
	b.HandleFunc("/"+NOTIFICATION+"/{"+ID+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			restGetNotificationByID(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/{"+ID+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			restDeleteNotificationByID(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc("/"+NOTIFICATION+"/"+SLUG+"/{"+SLUG+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			restGetNotificationBySlug(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+SLUG+"/{"+SLUG+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			restDeleteNotificationBySlug(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc("/"+NOTIFICATION+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			restDeleteNotificationsByAge(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc("/"+NOTIFICATION+"/"+SENDER+"/{"+SENDER+"}/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			restGetNotificationsBySender(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get),
				*container.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+START+"/{"+START+"}/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			restNotificationByStartEnd(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get),
				*container.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+START+"/{"+START+"}/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			restNotificationByStart(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get),
				*container.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			restNotificationByEnd(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get),
				*container.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+LABELS+"/{"+LABELS+"}/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			restNotificationsByLabels(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get),
				*container.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+NOTIFICATION+"/"+NEW+"/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			restNotificationsNew(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get),
				*container.ConfigurationFrom(dic.Get))
		}).Methods(http.MethodGet)

	// GetSubscriptions
	b.HandleFunc("/"+SUBSCRIPTION,
		func(writer http.ResponseWriter, request *http.Request) {
			restGetSubscriptions(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION,
		func(writer http.ResponseWriter, request *http.Request) {
			restAddSubscription(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodPost)
	b.HandleFunc("/"+SUBSCRIPTION,
		func(writer http.ResponseWriter, request *http.Request) {
			restUpdateSubscription(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodPut)
	b.HandleFunc("/"+SUBSCRIPTION+"/{"+ID+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			restGetSubscriptionByID(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION+"/{"+ID+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			restDeleteSubscriptionByID(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+SLUG+"/{"+SLUG+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			restGetSubscriptionBySlug(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+SLUG+"/{"+SLUG+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			restDeleteSubscriptionBySlug(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+CATEGORIES+"/{"+CATEGORIES+"}/"+LABELS+"/{"+LABELS+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			subscriptionsByCategoriesLabelsHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+CATEGORIES+"/{"+CATEGORIES+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			restGetSubscriptionsByCategories(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+LABELS+"/{"+LABELS+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			subscriptionsByLabelsHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+SUBSCRIPTION+"/"+RECEIVER+"/{"+RECEIVER+"}",
		func(writer http.ResponseWriter, request *http.Request) {
			subscriptionsByReceiverHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)

	// Transmissions
	b.HandleFunc("/"+TRANSMISSION,
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodPost)
	b.HandleFunc("/"+TRANSMISSION+"/"+SLUG+"/{"+SLUG+"}/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionBySlugHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+SLUG+"/{"+SLUG+"}/"+START+"/{"+START+"}/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionBySlugAndStartEndHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+START+"/{"+START+"}/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionByStartEndHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+START+"/{"+START+"}/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionByStartHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+END+"/{"+END+"}/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionByEndHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+ESCALATED+"/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionByEscalatedHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+FAILED+"/{"+LIMIT+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionByFailedHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodGet)
	b.HandleFunc("/"+TRANSMISSION+"/"+SENT+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionByAgeSentHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc("/"+TRANSMISSION+"/"+ESCALATED+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionByAgeEscalatedHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc("/"+TRANSMISSION+"/"+ACKNOWLEDGED+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionByAgeAcknowledgedHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc("/"+TRANSMISSION+"/"+FAILED+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			transmissionByAgeFailedHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)

	// Cleanup
	b.HandleFunc("/"+CLEANUP,
		func(writer http.ResponseWriter, request *http.Request) {
			cleanupHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)
	b.HandleFunc("/"+CLEANUP+"/"+AGE+"/{"+AGE+":[0-9]+}",
		func(writer http.ResponseWriter, request *http.Request) {
			cleanupAgeHandler(
				writer,
				request,
				bootstrapContainer.LoggingClientFrom(dic.Get),
				bootstrapContainer.DBClientFrom(dic.Get))
		}).Methods(http.MethodDelete)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}

func configHandler(
	w http.ResponseWriter,
	_ *http.Request,
	loggingClient logger.LoggingClient,
	config notificationsConfig.ConfigurationStruct) {

	pkg.Encode(config, w, loggingClient)
}

func metricsHandler(w http.ResponseWriter, _ *http.Request, loggingClient logger.LoggingClient) {
	s := telemetry.NewSystemUsage()

	pkg.Encode(s, w, loggingClient)

	return
}
