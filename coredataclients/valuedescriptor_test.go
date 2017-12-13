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
 * @microservice: core-clients-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package coredataclients

import (
	"fmt"
	"testing"
)

func TestGetvaluedescriptors(t *testing.T) {
	v, err := vdc.ValueDescriptors()
	if err != nil {
		t.FailNow()
	}
	fmt.Println(v)
}

var vdc ValueDescriptorClient

func TestMain(m *testing.M) {
	vdc = NewValueDescriptorClient("http://localhost:48080/api/v1/valuedescriptor")

	m.Run()
}
