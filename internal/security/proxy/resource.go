/*******************************************************************************
 * Copyright 2019 Dell Inc.
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
 * @author: Tingyu Zeng, Dell
 *******************************************************************************/

package proxy

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/internal"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type Resource struct {
	name             string
	client           internal.HttpCaller
	kongProxyBaseUrl string
	loggingClient    logger.LoggingClient
}

func NewResource(
	name string,
	r internal.HttpCaller,
	kongProxyBaseUrl string,
	lc logger.LoggingClient) *Resource {

	return &Resource{
		name:             name,
		client:           r,
		kongProxyBaseUrl: kongProxyBaseUrl,
		loggingClient:    lc,
	}
}

func (r *Resource) Remove(path string) error {
	tokens := []string{r.kongProxyBaseUrl, path, r.name}
	req, err := http.NewRequest(http.MethodDelete, strings.Join(tokens, "/"), nil)
	if err != nil {
		e := fmt.Sprintf("failed to delete %s at %s with error %s", r.name, path, err.Error())
		return errors.New(e)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		e := fmt.Sprintf("failed to delete %s at %s with error %s", r.name, path, err.Error())
		return errors.New(e)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		r.loggingClient.Infof("successful to delete %s at %s", r.name, path)
	default:
		e := fmt.Sprintf("failed to delete %s at %s with errocode %d.", r.name, path, resp.StatusCode)
		r.loggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}
