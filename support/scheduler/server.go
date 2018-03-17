//
// Copyright (c) 2018 Tencent
//
// SPDX-License-Identifier: Apache-2.0
//

package scheduler

import (
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/go-zoo/bone"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	ContentTypeKey = "Content-Type"
	ContentTypeTextValue = "application/text; charset=utf-8"
)

func replyPing(rw http.ResponseWriter, req *http.Request)  {
	rw.Header().Set(ContentTypeKey, ContentTypeTextValue)
	rw.WriteHeader(http.StatusOK)
	str := `pong`
	io.WriteString(rw, str)
}

func addSchedule(rw http.ResponseWriter, r *http.Request)  {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		io.WriteString(rw, err.Error())
		return
	}

	l := models.Schedule{}
	if err := json.Unmarshal(data, &l); err != nil {
		fmt.Println("Failed to parse Schedule : ", err)
		rw.WriteHeader(http.StatusBadRequest)
		io.WriteString(rw, err.Error())
		return
	}

	rw.WriteHeader(http.StatusAccepted)
}

func updateSchedule(rw http.ResponseWriter, r *http.Request) {

}

func removeSchedule(rw http.ResponseWriter, r *http.Request)  {

}

func addScheduleEvent(rw http.ResponseWriter, r *http.Request)  {

}

func updateScheduleEvent(rw http.ResponseWriter, r *http.Request)  {

}

func removeScheduleEvent(rw http.ResponseWriter, r *http.Request) {

}

func httpServer() http.Handler {
	mux := bone.New()
	mv1 := mux.Prefix("api/v1")

	mv1.Get("/ping", http.HandlerFunc(replyPing))

	// schedule
	mv1.Post("/schedule", http.HandlerFunc(addSchedule))
	mv1.Put("/schedule", http.HandlerFunc(updateSchedule))
	mv1.Delete("/schedule/:id", http.HandlerFunc(removeSchedule))

	// schedule event
	mv1.Post("/scheduleevent", http.HandlerFunc(addScheduleEvent))
	mv1.Put("/scheduleevent", http.HandlerFunc(updateScheduleEvent))
	mv1.Delete("/scheduleevent", http.HandlerFunc(removeScheduleEvent))

	return mux
}

func StartHttpServer(config Config, errChan chan error) {
	go func() {
		p := fmt.Sprintf(":%d", config.Port)
		errChan <- http.ListenAndServe(p, httpServer())
	}()
}
