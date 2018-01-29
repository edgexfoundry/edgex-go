//
// Copyright (c) 2017 Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"io/ioutil"
	"net/http"
	"testing"
)

func TestGetStatus(t *testing.T) {
	cases := []struct {
		body string
		code int
	}{
		{`{"running": true}`, 200},
	}

	url := ts.URL + "/status"

	for i, c := range cases {
		res, err := http.Get(url)
		if err != nil {
			t.Errorf("case %d: %s", i+1, err.Error())
		}

		if res.StatusCode != c.code {
			t.Errorf("case %d: expected status %d got %d", i+1, c.code, res.StatusCode)
		}

		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatalf("case %d: %s", i+1, err.Error())
		}

		if c.body != string(body) {
			t.Errorf("case %d: expected response %s got %s", i+1, c.body, string(body))
		}
	}
}
