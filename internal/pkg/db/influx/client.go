/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
 *******************************************************************************/

package influx

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/influxdata/influxdb/client/v2"
)

var currentInfluxClient *InfluxClient // Singleton used so that InfluxEvent can use it to de-reference readings

/*
Core data client
Has functions for interacting with the core data influxdb
*/

type InfluxClient struct {
	Client   client.Client // Influxdb client
	Database string        // Influxdb database name
}

// Return a pointer to the InfluxClient
func NewClient(config db.Configuration) (*InfluxClient, error) {
	// Create the dial info for the Influx session
	connectionString := "http://" + config.Host + ":" + strconv.Itoa(config.Port)
	influxdbHTTPInfo := client.HTTPConfig{
		Addr:     connectionString,
		Timeout:  time.Duration(config.Timeout) * time.Millisecond,
		Username: config.Username,
		Password: config.Password,
	}
	c, err := client.NewHTTPClient(influxdbHTTPInfo)
	if err != nil {
		return nil, err
	}

	influxClient := &InfluxClient{Client: c, Database: config.DatabaseName}
	currentInfluxClient = influxClient // Set the singleton
	return influxClient, nil
}

func (ic *InfluxClient) Connect() error {
	return nil
}

// Perform an Influxdb query
func (ic *InfluxClient) queryDB(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: ic.Database,
	}
	response, err := ic.Client.Query(q)
	if err != nil {
		return res, err
	}
	if response.Error() != nil {
		return res, response.Error()
	}
	res = response.Results
	return response.Results, nil
}

// Get count
func (ic *InfluxClient) getCount(query string) (int, error) {
	res, err := ic.queryDB(query)
	if err != nil {
		return 0, err
	}
	var n int64
	n = 0
	if len(res) == 1 {
		if len(res[0].Series) == 1 {
			n, err = res[0].Series[0].Values[0][1].(json.Number).Int64()
			if err != nil {
				return 0, err
			}
		}
	}
	return int(n), nil
}

func (ic *InfluxClient) CloseSession() {
	ic.Client.Close()
}

func (ic *InfluxClient) deleteById(collection string, id string) error {
	q := fmt.Sprintf("DROP SERIES FROM %s WHERE id = '%s'", collection, id)
	_, err := ic.queryDB(q)
	if err != nil {
		return db.ErrNotFound
	}
	return nil
}

func (ic *InfluxClient) deleteAll(collection string) error {
	q := fmt.Sprintf("DELETE FROM %s", collection)
	_, err := ic.queryDB(q)
	if err != nil {
		return err
	}
	return nil
}
