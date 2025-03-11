/*
	Copyright 2019 NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package ziti

import (
	"crypto/x509"
	"github.com/golang-jwt/jwt/v5"
	"github.com/michaelquigley/pfxlog"
	"net/url"
)

var EnrollUrl, _ = url.Parse("/edge/client/v1/enroll")

const EnrollmentMethodCa = "ca"

type Versions struct {
	Api           string `json:"api"`
	EnrollmentApi string `json:"enrollmentApi"`
}

type EnrollmentClaims struct {
	jwt.RegisteredClaims
	EnrollmentMethod string            `json:"em"`
	Controllers      []string          `json:"ctrls"`
	SignatureCert    *x509.Certificate `json:"-"`
}

func (t *EnrollmentClaims) EnrolmentUrl() string {
	enrollmentUrl, err := url.Parse(t.Issuer)

	if err != nil {
		pfxlog.Logger().WithError(err).WithField("url", t.Issuer).Panic("could not parse issuer as URL")
	}

	enrollmentUrl = enrollmentUrl.ResolveReference(EnrollUrl)

	query := enrollmentUrl.Query()
	query.Add("method", t.EnrollmentMethod)

	if t.EnrollmentMethod != EnrollmentMethodCa {
		query.Add("token", t.ID)
	}

	enrollmentUrl.RawQuery = query.Encode()

	return enrollmentUrl.String()
}
