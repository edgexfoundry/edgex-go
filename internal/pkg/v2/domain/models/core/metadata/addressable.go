/*******************************************************************************
 * Copyright 2020 Dell Inc.
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

package metadata

import "github.com/edgexfoundry/edgex-go/internal/pkg/v2/infrastructure"

// Addressable defines the persistence model.
type Addressable struct {
	ID        infrastructure.Identity
	Name      string
	Protocol  string
	Method    string
	Address   string
	Port      string
	Path      string
	Publisher string
	User      string
	Password  string
	Topic     string
}

// NewIdentity returns an initialized Addressable model.
func New(
	id infrastructure.Identity,
	name string,
	protocol string,
	method string,
	address string,
	port string,
	path string,
	publisher string,
	user string,
	password string,
	topic string) *Addressable {

	return &Addressable{
		ID:        id,
		Name:      name,
		Protocol:  protocol,
		Method:    method,
		Address:   address,
		Port:      port,
		Path:      path,
		Publisher: publisher,
		User:      user,
		Password:  password,
		Topic:     topic,
	}
}
