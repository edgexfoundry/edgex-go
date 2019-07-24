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
 * @version: 1.1.0
 *******************************************************************************/

package proxy

import (
	"errors"
	"fmt"
	"github.com/dghubble/sling"
	"net/http"
)

type Resource struct {
	name   string
	client Requestor
}

func NewResource(name string, r Requestor) Resource {
	return Resource{name: name, client: r}
}

func (r *Resource) Remove(path string) error {
	req, err := sling.New().Base(Configuration.KongURL.GetProxyBaseURL()).Path(path).Delete(r.name).Request()
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
		LoggingClient.Info(fmt.Sprintf("successful to delete %s at %s", r.name, path))
		break
	default:
		e := fmt.Sprintf("failed to delete %s at %s with errocode %d.", r.name, path, resp.StatusCode)
		LoggingClient.Error(e)
		return errors.New(e)
	}
	return nil
}
