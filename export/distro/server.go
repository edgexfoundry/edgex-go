//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/edgex-go/export"
	"github.com/go-zoo/bone"
	"go.uber.org/zap"
)

func replyPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/text; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	str := `pong`
	io.WriteString(w, str)
}

func replyNotifyRegistrations(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error("Failed read body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	update := export.NotifyUpdate{}
	if err := json.Unmarshal(data, &update); err != nil {
		logger.Error("Failed to parse", zap.ByteString("json", data))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}
	if update.Name == "" || update.Operation == "" {
		logger.Error("Missing json field", zap.Any("update", update))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if update.Operation != export.NotifyUpdateAdd &&
		update.Operation != export.NotifyUpdateUpdate &&
		update.Operation != export.NotifyUpdateDelete {
		logger.Error("Invalid value for operation",
			zap.String("operation", update.Operation))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	RefreshRegistrations(update)
}

// HTTPServer function
func httpServer() http.Handler {
	mux := bone.New()

	mux.Get("/api/v1/ping", http.HandlerFunc(replyPing))
	mux.Put("/api/v1/notify/registrations", http.HandlerFunc(replyNotifyRegistrations))

	return mux
}
