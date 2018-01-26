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
 * @microservice: core-domain-go library
 * @author: Jim White, Dell
 * @version: 0.5.0
 *******************************************************************************/

package models

import "testing"

func TestProtocol_UnmarshalJSON(t *testing.T) {
	var http = HTTP
	var tcp = TCP
	var mac = MAC
	var zmq = ZMQ
	var other = OTHER
	var ssl = SSL
	var foo Protocol
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		p       *Protocol
		args    args
		wantErr bool
	}{
		{"HTTP unmarshal", &http, args{[]byte("\"HTTP\"")}, false},
		{"TCP unmarshal", &tcp, args{[]byte("\"TCP\"")}, false},
		{"MAC unmarshal", &mac, args{[]byte("\"MAC\"")}, false},
		{"ZMQ unmarshal", &zmq, args{[]byte("\"ZMQ\"")}, false},
		{"OTHER unmarshal", &other, args{[]byte("\"OTHER\"")}, false},
		{"SSL unmarshal", &ssl, args{[]byte("\"SSL\"")}, false},
		{"bad unmarshal", &foo, args{[]byte("\"foo\"")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("Protocol.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
