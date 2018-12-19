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

type Put struct {
	Path           string     `bson:"path"`      // path used by service for action on a device or sensor
	Responses      []Response `bson:"responses"` // responses from get or put requests to service
	URL            string     // url for requests from command service
	ParameterNames []string   `bson:"parameterNames"`
}

func (p *Put) ToContract() contract.Put {
	to := contract.Put{}
	to.Path = p.Path
	to.Responses = []contract.Response{}
	for _, r := range p.Responses {
		to.Responses = append(to.Responses, r.ToContract())
	}
	to.URL = p.URL
	to.ParameterNames = p.ParameterNames
	return to
}

func (p *Put) FromContract(from contract.Put) error {
	p.Path = from.Path
	p.Responses = []Response{}
	for _, val := range from.Responses {
		r := &Response{}
		err := r.FromContract(val)
		if err != nil {
			return errors.New(err.Error() + " code: " + val.Code)
		}
		p.Responses = append(p.Responses, *r)

	}
	p.URL = from.URL
	p.ParameterNames = from.ParameterNames
	return nil
}
