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
 *
 * @microservice: core-data-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package data

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)


// Helper function for encoding things for returning from REST calls
func encode(i interface{}, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err := enc.Encode(i)
	// Problems encoding
	if err != nil {
		loggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
}

// Printing function purely for debugging purposes
// Print the body of a request to the console
func printBody(r io.ReadCloser) {
	body, err := ioutil.ReadAll(r)
	bodyString := string(body)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(bodyString)
}

// Test if the service is working
func pingHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	_, err := w.Write([]byte("pong"))
	if err != nil {
		loggingClient.Error("Error writing pong: " + err.Error())
	}
}
