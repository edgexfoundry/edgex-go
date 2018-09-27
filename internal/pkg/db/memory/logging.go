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
package memory

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func (m *MemDB) AddLog(le models.LogEntry) error {
	m.logs = append(m.logs, le)

	return nil
}

func (m *MemDB) DeleteLog(criteria db.LogMatcher) (int, error) {
	return 0, nil
}

func (m *MemDB) FindLog(criteria db.LogMatcher, limit int) ([]models.LogEntry, error) {
	return m.logs, nil
}

func (m *MemDB) ResetLogs() {
	m.logs = []models.LogEntry{}
}
