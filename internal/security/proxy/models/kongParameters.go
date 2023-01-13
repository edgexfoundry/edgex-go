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
package models

type KongService struct {
	Name     string `url:"name,omitempty"`
	Host     string `url:"host,omitempty"`
	Port     int    `url:"port,omitempty"`
	Protocol string `url:"protocol,omitempty"`
}

// KongServiceResponse is the response from Kong when creating a service
type KongServiceResponse struct {
	ID             string `json:"id,omitempty"`
	CreatedAt      uint64 `json:"created_at,omitempty"`
	UpdatedAt      uint64 `json:"updated_at,omitempty"`
	ConnectTimeout int64  `json:"connect_timeout,omitempty"`
	Protocol       string `json:"protocol,omitempty"`
	Host           string `json:"host,omitempty"`
	Port           uint64 `json:"port,omitempty"`
	Path           string `json:"path,omitempty"`
	Name           string `json:"name,omitempty"`
	Retries        int64  `json:"retries,omitempty"`
	ReadTimeout    int64  `json:"read_timeout,omitempty"`
	WriteTimeout   int64  `json:"write_timeout,omitempty"`
}

type KongRoute struct {
	Paths []string `json:"paths,omitempty"`
	Name  string   `json:"name,omitempty"`
}

type CertInfo struct {
	Cert string   `json:"cert,omitempty"`
	Key  string   `json:"key,omitempty"`
	Snis []string `json:"snis,omitempty"`
}

type Item struct {
	ID string `json:"id"`
}

type DataCollect struct {
	Section []Item `json:"data"`
}
