//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/export"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/go-zoo/bone"
)

func replyPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	str := `pong`
	io.WriteString(w, str)
}

func replyConfig(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err := enc.Encode(Configuration)
	// Problems encoding
	if err != nil {
		LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func replyNotifyRegistrations(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed read body. Error: %s", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	update := models.NotifyUpdate{}
	if err := json.Unmarshal(data, &update); err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to parse %X", data))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}
	if update.Name == "" || update.Operation == "" {
		LoggingClient.Error(fmt.Sprintf("Missing json field: %s", update.Name))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if update.Operation != export.NotifyUpdateAdd &&
		update.Operation != export.NotifyUpdateUpdate &&
		update.Operation != export.NotifyUpdateDelete {
		LoggingClient.Error(fmt.Sprintf("Invalid value for operation %s", update.Operation))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	RefreshRegistrations(update)
}

// HTTPServer function
func httpServer() http.Handler {
	mux := bone.New()
	mux.Get(internal.ApiPingRoute, http.HandlerFunc(replyPing))
	mux.Get(internal.ApiConfigRoute, http.HandlerFunc(replyConfig))
	mux.Put(clients.ApiNotifyRegistrationRoute, http.HandlerFunc(replyNotifyRegistrations))

	return mux
}
