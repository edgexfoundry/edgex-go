/********************************************************************************
 *  Copyright 2019 Dell Inc.
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

package secret

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
)

func (s *SecretProvider) GetDatabaseCredentials(database config.Database) (config.Credentials, error) {
	// If security is disabled or the database is Redis then we are to use the credentials supplied by the
	// configuration. The reason we do this for Redis is because Redis does not have an authentication nor an
	// authorization mechanism.
	if !s.isSecurityEnabled() || database.Type == db.RedisDB {
		return config.Credentials{
			Username: database.Username,
			Password: database.Password,
		}, nil
	}

	secrets, err := s.secretClient.GetSecrets(database.Type, "username", "password")
	if err != nil {
		return config.Credentials{}, err
	}

	return config.Credentials{
		Username: secrets["username"],
		Password: secrets["password"],
	}, nil

}
