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
	"encoding/json"
	"io/ioutil"
)

type userTokenPair struct {
	User  string
	Token string
}

type Writer interface {
	Save(u string, t string) error
}

type TokenFileWriter struct {
	Filename string
}

func (tf *TokenFileWriter) Save(u string, t string) error {

	data := userTokenPair{
		User:  u,
		Token: t,
	}

	jdata, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(tf.Filename, jdata, 0600)
	return err
}
