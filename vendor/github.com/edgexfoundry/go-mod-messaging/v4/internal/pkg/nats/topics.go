//
// Copyright (c) 2022 One Track Consulting
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

//go:build include_nats_messaging

package nats

import (
	"strings"
)

const (
	StandardSeparator           = "/"
	StandardSingleLevelWildcard = "+"
	StandardMultiLevelWildcard  = "#"
	Separator                   = "."
	SingleLevelWildcard         = "*"
	MultiLevelWildcard          = ">"
)

var subjectReplacer = strings.NewReplacer(StandardSeparator, Separator, StandardSingleLevelWildcard, SingleLevelWildcard, StandardMultiLevelWildcard, MultiLevelWildcard)

// TopicToSubject formats an EdgeX topic into a NATS subject
func TopicToSubject(topic string) string {
	return subjectReplacer.Replace(topic)
}

var topicReplacer = strings.NewReplacer(Separator, StandardSeparator, SingleLevelWildcard, StandardSingleLevelWildcard, MultiLevelWildcard, StandardMultiLevelWildcard)

// subjectToTopic formats a NATS subject into an EdgeX topic
func subjectToTopic(subject string) string {
	return topicReplacer.Replace(subject)
}
