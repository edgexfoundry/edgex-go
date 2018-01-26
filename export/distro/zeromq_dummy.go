//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

// +build !zeromq

package distro

import (
	"encoding/json"
	"time"

	"github.com/edgexfoundry/edgex-go/export"
	"go.uber.org/zap"
)

const sampleEvent string = `{"pushed":0,"device":"livingroomthermostat",
	"readings":[
	{"pushed":0,"name":"temperature","value":"72","id":"57ed24f0502fdf73bb637915","created":1475159280744,"modified":1475159280744,"origin":1471806386919},
	{"pushed":0,"name":"humidity","value":"172","id":"27ed24f0502fdf73bb637915","created":2475159280744,"modified":2475159280744,"origin":2471806386919},
	{"pushed":0,"name":"humidity","value":"58","id":"57ed24f0502fdf73bb637916","created":1475159280756,"modified":1475159280756,"origin":1471806386919},
	{"pushed":0,"name":"rpm","value":"58","id":"57ed24f0502fdf73bb637916","created":1475159280756,"modified":1475159280756,"origin":1471806386919}],
	"id":"57ed24f0502fdf73bb637917","created":1475159280762,"modified":1475159280762,"origin":1471806386919}`

func ZeroMQReceiver(eventCh chan *export.Event) {
	go func() {
		ev := parseEvent(sampleEvent)
		for {
			time.Sleep(time.Second)
			eventCh <- ev
			logger.Info("Event generated")
		}
	}()
}

func parseEvent(str string) *export.Event {
	event := export.Event{}
	if err := json.Unmarshal([]byte(str), &event); err != nil {
		logger.Error("Failed to query add registration", zap.Error(err))
		return nil
	}

	return &event
}
