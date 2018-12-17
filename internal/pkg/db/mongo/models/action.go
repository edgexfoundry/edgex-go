/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

package models

import (
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/pkg/errors"
)

type Action struct {
	Path      string     `bson:"path"`      // path used by service for action on a device or sensor
	Responses []Response `bson:"responses"` // responses from get or put requests to service
	URL       string     // url for requests from command service
}

func (a Action) ToContract() contract.Action {
	to := contract.Action{
		Path:      a.Path,
		Responses: []contract.Response{},
		URL:       a.URL,
	}
	for _, r := range a.Responses {
		to.Responses = append(to.Responses, r.ToContract())
	}
	return to
}

func (a *Action) FromContract(from contract.Action) error {
	a.Path = from.Path
	a.Responses = []Response{}
	for _, val := range from.Responses {
		r := &Response{}
		err := r.FromContract(val)
		if err != nil {
			return errors.New(err.Error() + " code: " + val.Code)
		}
		a.Responses = append(a.Responses, *r)

	}
	a.URL = from.URL

	return nil
}
