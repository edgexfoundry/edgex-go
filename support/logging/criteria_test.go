//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/support/domain"
)

const ()

func TestCriteriaMatch(t *testing.T) {
	var services = []string{"service1", "service2"}
	var levels = []string{support_domain.TRACE, support_domain.DEBUG}
	var keywords = []string{"2"}
	var keywordsEmptyString = []string{""}
	var labels1 = []string{"label1"}
	var labels2 = []string{"label2"}
	var labels12 = []string{"label2", "label1"}
	var labels3 = []string{"1", "2", "label2"}

	var tests = []struct {
		name     string
		log      support_domain.LogEntry
		criteria matchCriteria
		result   bool
	}{
		{"empty/empty", support_domain.LogEntry{}, matchCriteria{}, true},
		// services
		{"emptyService", support_domain.LogEntry{}, matchCriteria{OriginServices: services}, false},
		{"wrongService", support_domain.LogEntry{OriginService: "service11"}, matchCriteria{OriginServices: services}, false},
		{"matchService", support_domain.LogEntry{OriginService: "service1"}, matchCriteria{OriginServices: services}, true},
		// Levels
		{"wrongLevel", support_domain.LogEntry{Level: support_domain.WARN}, matchCriteria{LogLevels: levels}, false},
		{"matchLevel", support_domain.LogEntry{Level: support_domain.DEBUG}, matchCriteria{LogLevels: levels}, true},
		// Start
		{"0Start", support_domain.LogEntry{Created: 5}, matchCriteria{Start: 0}, true},
		{"wrongStart", support_domain.LogEntry{Created: 5}, matchCriteria{Start: 6}, false},
		{"matchStart1", support_domain.LogEntry{Created: 6}, matchCriteria{Start: 6}, true},
		{"matchStart2", support_domain.LogEntry{Created: 7}, matchCriteria{Start: 6}, true},
		// End
		{"0End", support_domain.LogEntry{Created: 5}, matchCriteria{End: 0}, true},
		{"matchEnd", support_domain.LogEntry{Created: 5}, matchCriteria{End: 6}, true},
		{"matchEnd1", support_domain.LogEntry{Created: 6}, matchCriteria{End: 6}, true},
		{"wrongEnd", support_domain.LogEntry{Created: 7}, matchCriteria{End: 6}, false},
		// keywords
		{"noKeywords", support_domain.LogEntry{Message: "111111"}, matchCriteria{}, true},
		{"wrongKeywords", support_domain.LogEntry{Message: "111111"}, matchCriteria{Keywords: keywords}, false},
		{"matchKeywords", support_domain.LogEntry{Message: "222222"}, matchCriteria{Keywords: keywords}, true},
		{"KeywordsEmptyString", support_domain.LogEntry{Message: "222222"}, matchCriteria{Keywords: keywordsEmptyString}, true},
		{"KeywordsEmptyString2", support_domain.LogEntry{Message: ""}, matchCriteria{Keywords: keywordsEmptyString}, true},
		// labels
		{"noLabels", support_domain.LogEntry{Labels: labels1}, matchCriteria{}, true},
		{"matchLabels", support_domain.LogEntry{Labels: labels1}, matchCriteria{Labels: labels1}, true},
		{"matchLabels2", support_domain.LogEntry{Labels: labels1}, matchCriteria{Labels: labels12}, true},
		{"wrongLabels", support_domain.LogEntry{Labels: labels1}, matchCriteria{Labels: labels2}, false},
		{"wrongLabels", support_domain.LogEntry{Labels: labels1}, matchCriteria{Labels: labels3}, false},
	}
	le := support_domain.LogEntry{}

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
