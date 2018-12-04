//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/internal/support/logging/models"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

func TestCriteriaMatch(t *testing.T) {
	var services = []string{"service1", "service2"}
	var levels = []string{logger.TraceLog, logger.DebugLog}
	var keywords = []string{"2"}
	var keywordsEmptyString = []string{""}
	var labels1 = []string{"label1"}
	var args1 = make([]interface{}, len(labels1))
	args1[0] = labels1[0]
	var labels2 = []string{"label2"}
	var labels12 = []string{"label2", "label1"}
	var labels3 = []string{"1", "2", "label2"}

	var tests = []struct {
		name     string
		log      models.LogEntry
		criteria matchCriteria
		result   bool
	}{
		{"empty/empty", models.LogEntry{}, matchCriteria{}, true},
		// services
		{"emptyService", models.LogEntry{}, matchCriteria{OriginServices: services}, false},
		{"wrongService", models.LogEntry{OriginService: "service11"}, matchCriteria{OriginServices: services}, false},
		{"matchService", models.LogEntry{OriginService: "service1"}, matchCriteria{OriginServices: services}, true},
		// Levels
		{"wrongLevel", models.LogEntry{Level: logger.WarnLog}, matchCriteria{LogLevels: levels}, false},
		{"matchLevel", models.LogEntry{Level: logger.DebugLog}, matchCriteria{LogLevels: levels}, true},
		// Start
		{"0Start", models.LogEntry{Created: 5}, matchCriteria{Start: 0}, true},
		{"wrongStart", models.LogEntry{Created: 5}, matchCriteria{Start: 6}, false},
		{"matchStart1", models.LogEntry{Created: 6}, matchCriteria{Start: 6}, true},
		{"matchStart2", models.LogEntry{Created: 7}, matchCriteria{Start: 6}, true},
		// End
		{"0End", models.LogEntry{Created: 5}, matchCriteria{End: 0}, true},
		{"matchEnd", models.LogEntry{Created: 5}, matchCriteria{End: 6}, true},
		{"matchEnd1", models.LogEntry{Created: 6}, matchCriteria{End: 6}, true},
		{"wrongEnd", models.LogEntry{Created: 7}, matchCriteria{End: 6}, false},
		// keywords
		{"noKeywords", models.LogEntry{Message: "111111"}, matchCriteria{}, true},
		{"wrongKeywords", models.LogEntry{Message: "111111"}, matchCriteria{Keywords: keywords}, false},
		{"matchKeywords", models.LogEntry{Message: "222222"}, matchCriteria{Keywords: keywords}, true},
		{"KeywordsEmptyString", models.LogEntry{Message: "222222"}, matchCriteria{Keywords: keywordsEmptyString}, true},
		{"KeywordsEmptyString2", models.LogEntry{Message: ""}, matchCriteria{Keywords: keywordsEmptyString}, true},
		// labels
		{"noLabels", models.LogEntry{Args: args1}, matchCriteria{}, true},
		{"matchLabels", models.LogEntry{Args: args1}, matchCriteria{Labels: labels1}, true},
		{"matchLabels2", models.LogEntry{Args: args1}, matchCriteria{Labels: labels12}, true},
		{"wrongLabels", models.LogEntry{Args: args1}, matchCriteria{Labels: labels2}, false},
		{"wrongLabels", models.LogEntry{Args: args1}, matchCriteria{Labels: labels3}, false},
	}
	le := models.LogEntry{}

	criteria := matchCriteria{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result != tt.criteria.match(tt.log) {
				t.Errorf("matching log %v criteria %v should be %v",
					tt.log, tt.criteria, tt.result)
			}
		})
	}

	if !criteria.match(le) {
		t.Errorf("log %v should match criteria %v", le, criteria)
	}
}
