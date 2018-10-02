//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

type matchCriteria struct {
	OriginServices []string
	LogLevels      []string
	Labels         []string
	Keywords       []string
	Start          int64
	End            int64
	Limit          int
}

func matchStringInSlice(s string, l []string) bool {
	if len(l) > 0 {
		for _, i := range l {
			if i == s {
				return true
			}
		}
		return false
	}
	return true
}

func (criteria matchCriteria) match(le models.LogEntry) bool {
	if !matchStringInSlice(le.OriginService, criteria.OriginServices) {
		return false
	}
	if !matchStringInSlice(le.Level, criteria.LogLevels) {
		return false
	}

	if criteria.Start > 0 {
		if criteria.Start > le.Created {
			return false
		}
	}
	if criteria.End > 0 {
		if criteria.End < le.Created {
			return false
		}
	}
	if len(criteria.Keywords) > 0 {
		found := false
		for _, keyword := range criteria.Keywords {
			if strings.Contains(le.Message, keyword) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(criteria.Labels) > 0 {
		found := false
		for _, label := range criteria.Labels {
			if matchStringInSlice(label, le.Labels) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func createConditions(conditions []bson.M, field string, elements []string) []bson.M {
	keyCond := []bson.M{}
	for _, value := range elements {
		keyCond = append(keyCond, bson.M{field: value})
	}

	return append(conditions, bson.M{"$or": keyCond})
}

func (criteria matchCriteria) CreateQuery() map[string]interface{} {
	conditions := []bson.M{{}}

	if len(criteria.Labels) > 0 {
		conditions = createConditions(conditions, "labels", criteria.Labels)
	}

	if len(criteria.Keywords) > 0 {
		keyCond := []bson.M{}
		for _, key := range criteria.Keywords {
			regex := fmt.Sprintf(".*%s.*", key)
			keyCond = append(keyCond, bson.M{"message": bson.M{"$regex": regex}})
		}
		conditions = append(conditions, bson.M{"$or": keyCond})
	}

	if len(criteria.OriginServices) > 0 {
		conditions = createConditions(conditions, "originService", criteria.OriginServices)
	}

	if len(criteria.LogLevels) > 0 {
		conditions = createConditions(conditions, "logLevel", criteria.LogLevels)
	}

	if criteria.Start != 0 {
		conditions = append(conditions, bson.M{"created": bson.M{"$gt": criteria.Start}})
	}

	if criteria.End != 0 {
		conditions = append(conditions, bson.M{"created": bson.M{"$lt": criteria.End}})
	}

	return bson.M{"$and": conditions}
}
