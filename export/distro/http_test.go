//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/core/domain/models"
)

const ()

func TestHttpSender(t *testing.T) {
	const (
		msgStr = "test message"
		path   = "/somepath/foo"
	)

	var tests = []struct {
		name string
		addr models.Addressable
	}{
		{"noMethod", models.Addressable{
			Protocol: "http",
			Path:     path}},
		{"get", models.Addressable{
			Protocol:   "http",
			HTTPMethod: http.MethodGet,
			Path:       path}},
		{"post", models.Addressable{
			Protocol:   "http",
			HTTPMethod: http.MethodPost,
			Path:       path}},
		{"postInvalidPort", models.Addressable{
			Protocol:   "http",
			HTTPMethod: http.MethodPost,
			Path:       path,
			Port:       -1}},
	}

	var addressableTest models.Addressable
	var msg = []byte(msgStr)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)

				readMsg, _ := ioutil.ReadAll(r.Body)
				r.Body.Close()

				if bytes.Compare(readMsg, msg) != 0 {
					t.Errorf("Invalid msg received %v, expected %v", readMsg, msg)

				}

				if r.Method != addressableTest.HTTPMethod {
					t.Errorf("Invalid method received %s, expected %s",
						r.Method, addressableTest.HTTPMethod)
				}
				if r.URL.EscapedPath() != path {
					t.Errorf("Invalid path received %s, expected %s",
						r.URL.EscapedPath(), path)
				}

			}

			// create test server with handler
			ts := httptest.NewServer(http.HandlerFunc(handler))
			defer ts.Close()

			url, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal("Could not parse url")
			}

			h, p, err := net.SplitHostPort(url.Host)
			if err != nil {
				t.Fatal("Could get and port")
			}
			port, err := strconv.Atoi(p)
			if err != nil {
				t.Fatal("Could not parse port")
			}

			addressableTest = tt.addr
			addressableTest.Address = h
			// Only overwrite the port if it had the default value
			if addressableTest.Port == 0 {
				addressableTest.Port = port
			}
			sender := NewHTTPSender(addressableTest)
			sender.Send(msg)
		})
	}
}
