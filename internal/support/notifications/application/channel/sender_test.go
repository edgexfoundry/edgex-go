//
// Copyright (C) 2026 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package channel

import (
	"strconv"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type publishRecorder struct {
	mqtt.Client
	published bool
	qos       byte
}

func (c *publishRecorder) Publish(_ string, qos byte, _ bool, _ interface{}) mqtt.Token {
	c.published = true
	c.qos = qos
	return &mqtt.DummyToken{}
}

func TestMQTTSenderSendQoS(t *testing.T) {
	tests := []struct {
		qos         int
		expectError bool
	}{
		{0, false},
		{2, false},
		{-1, true},
		{3, true},
		// would wrap to 0 if converted to byte unchecked
		{256, true},
	}
	for _, testCase := range tests {
		t.Run(strconv.Itoa(testCase.qos), func(t *testing.T) {
			client := &publishRecorder{}
			address := models.MQTTPubAddress{
				BaseAddress: models.BaseAddress{Host: "localhost", Port: 1883},
				Publisher:   "publisher",
				QoS:         testCase.qos,
			}
			sender := &MQTTSender{clientCache: map[string]mqtt.Client{"publisher:localhost:1883": client}}

			_, err := sender.Send(models.Notification{}, address)
			if testCase.expectError {
				require.Error(t, err)
				assert.False(t, client.published)
				return
			}
			require.NoError(t, err)
			assert.EqualValues(t, testCase.qos, client.qos)
		})
	}
}
