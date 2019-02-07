//
// Copyright (c) 2018
// IOTech
//
// SPDX-License-Identifier: Apache-2.0

package distro

import (
	"fmt"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"

	"github.com/influxdata/influxdb/client/v2"
)

const (
	influxDBTimeout = 60000
)

type influxdbSender struct {
	client   client.Client
	httpInfo client.HTTPConfig
	database string
}

func newInfluxDBSender(addr contract.Addressable) sender {
	connStr := "http://" + addr.Address + ":" + strconv.Itoa(addr.Port)

	influxdbHTTPInfo := client.HTTPConfig{
		Addr:     connStr,
		Timeout:  time.Duration(influxDBTimeout) * time.Millisecond,
		Username: addr.User,
		Password: addr.Password,
	}

	db := addr.Topic

	sender := &influxdbSender{
		client:   nil,
		httpInfo: influxdbHTTPInfo,
		database: db,
	}

	return sender
}

func (sender *influxdbSender) Send(data []byte, event *models.Event) bool {
	if sender.client == nil {
		LoggingClient.Info("Connecting to InfluxDB server")
		c, err := client.NewHTTPClient(sender.httpInfo)

		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Failed to connect to InfluxDB server: %s", err))
			return false
		}

		sender.client = c
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  sender.database,
		Precision: "us",
	})

	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to craete batch points: %s", err))
		return false
	}

	for _, reading := range event.Readings {
		value, err := strconv.ParseFloat(reading.Value, 64)

		if err != nil {
			// not a valid numerical reading value, just ignore it
			continue
		}

		fields := map[string]interface{}{
			"created": reading.Created,
			"origin":  reading.Origin,
			"value":   value,
		}

		tags := map[string]string{
			"device":        reading.Device,
			"resource_name": reading.Name,
			"event_id":      event.ID,
		}

		pt, err := client.NewPoint(
			"readings",
			tags,
			fields,
			time.Now(),
		)

		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Failed to add data point: %s", err))
			return false
		}

		bp.AddPoint(pt)
	}

	err = sender.client.Write(bp)

	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to write data points to InfluxDB server: %s", err))
		sender.client = nil // Reset the client
		return false
	}

	return true
}
