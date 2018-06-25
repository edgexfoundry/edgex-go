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

	"github.com/gorilla/mux"
)

func LoadRestRoutes() *mux.Router {
	r := mux.NewRouter()
	b := r.PathPrefix("/api/v1").Subrouter()

	// Notifications
	b.HandleFunc("/notification", notificationHandler).Methods(http.MethodPost)
	b.HandleFunc("/notification/{id}", notificationByIDHandler).Methods(http.MethodGet)
	b.HandleFunc("/notification/id/{id}", notificationByIDHandler).Methods(http.MethodDelete)
	b.HandleFunc("/notification/slug/{slug}", notificationBySlugHandler).Methods(http.MethodGet, http.MethodDelete)
	b.HandleFunc("/notification/age/{age:[0-9]+}", notificationOldHandler).Methods(http.MethodDelete)
	b.HandleFunc("/notification/sender/{sender}/{limit:[0-9]+}", notificationBySenderHandler).Methods(http.MethodGet)
	b.HandleFunc("/notification/start/{start}/end/{end}/{limit:[0-9]+}", notificationByStartEndHandler).Methods(http.MethodGet)
	b.HandleFunc("/notification/start/{start}/{limit:[0-9]+}", notificationByStartHandler).Methods(http.MethodGet)
	b.HandleFunc("/notification/end/{end}/{limit:[0-9]+}", notificationByEndHandler).Methods(http.MethodGet)
	b.HandleFunc("/notification/labels/{labels}/{limit:[0-9]+}", notificationsByLabelsHandler).Methods(http.MethodGet)
	b.HandleFunc("/notification/new/{limit:[0-9]+}", notificationsNewHandler).Methods(http.MethodGet)

	// Subscriptions
	b.HandleFunc("/subscription", subscriptionHandler).Methods(http.MethodPost)
	b.HandleFunc("/subscription/{id}", subscriptionByIDHandler).Methods(http.MethodGet)
	b.HandleFunc("/subscription/slug/{slug:.+}", subscriptionsBySlugHandler).Methods(http.MethodGet, http.MethodDelete)
	b.HandleFunc("/subscription/categories/{categories}/labels/{labels}", subscriptionsByCategoriesLabelsHandler).Methods(http.MethodGet)
	b.HandleFunc("/subscription/categories/{categories}", subscriptionsByCategoriesHandler).Methods(http.MethodGet)
	b.HandleFunc("/subscription/labels/{labels}", subscriptionsByLabelsHandler).Methods(http.MethodGet)
	b.HandleFunc("/subscription/receiver/{receiver:.+}", subscriptionsByReceiverHandler).Methods(http.MethodGet)

	// Transmissions
	b.HandleFunc("/transmission", transmissionHandler).Methods(http.MethodPost)
	b.HandleFunc("/transmission/slug/{slug}/{limit:[0-9]+}", transmissionBySlugHandler).Methods(http.MethodGet)
	b.HandleFunc("/transmission/start/{start}/end/{end}/{limit:[0-9]+}", transmissionByStartEndHandler).Methods(http.MethodGet)
	b.HandleFunc("/transmission/start/{start}/{limit:[0-9]+}", transmissionByStartHandler).Methods(http.MethodGet)
	b.HandleFunc("/transmission/end/{end}/{limit:[0-9]+}", transmissionByEndHandler).Methods(http.MethodGet)
	b.HandleFunc("/transmission/escalated/{limit:[0-9]+}", transmissionByEscalatedHandler).Methods(http.MethodGet)
	b.HandleFunc("/transmission/failed/{limit:[0-9]+}", transmissionByFailedHandler).Methods(http.MethodGet)
	b.HandleFunc("/transmission/sent/age/{age:[0-9]+}", transmissionByAgeSentHandler).Methods(http.MethodDelete)
	b.HandleFunc("/transmission/escalated/age/{age:[0-9]+}", transmissionByAgeEscalatedHandler).Methods(http.MethodDelete)
	b.HandleFunc("/transmission/acknowledged/age/{age:[0-9]+}", transmissionByAgeAcknowledgedHandler).Methods(http.MethodDelete)
	b.HandleFunc("/transmission/failed/age/{age:[0-9]+}", transmissionByAgeFailedHandler).Methods(http.MethodDelete)

	// Cleanup
	b.HandleFunc("/cleanup", cleanupHandler).Methods(http.MethodDelete)
	b.HandleFunc("/cleanup/age/{age:[0-9]+}", cleanupAgeHandler).Methods(http.MethodDelete)

	// Ping Resource
	// /api/v1/ping
	b.HandleFunc("/ping", pingHandler).Methods(http.MethodGet)

	return r
}
