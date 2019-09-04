/*******************************************************************************
 * Copyright 2019 VMware Inc.
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

package subscription

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

type IdExecutor interface {
	Execute() (contract.Subscription, error)
}

type CollectionExecutor interface {
	Execute() ([]contract.Subscription, error)
}

type subscriptionLoadById struct {
	database SubscriptionLoader
	id       string
}

type subscriptionLoadBySlug struct {
	database SubscriptionLoader
	slug     string
}

type subscriptionsLoadByCategories struct {
	database   SubscriptionLoader
	categories []string
}

func (op subscriptionLoadById) Execute() (contract.Subscription, error) {
	res, err := op.database.GetSubscriptionById(op.id)
	if err != nil {
		if err == db.ErrNotFound {
			err = errors.NewErrSubscriptionNotFound(op.id)
		}
		return res, err
	}
	return res, nil
}

func (op subscriptionLoadBySlug) Execute() (contract.Subscription, error) {
	s, err := op.database.GetSubscriptionBySlug(op.slug)
	if err != nil {
		if err == db.ErrNotFound {
			newErr := errors.NewErrSubscriptionNotFound(op.slug)
			return s, newErr
		}
		return s, err
	}
	return s, nil
}

func (op subscriptionsLoadByCategories) Execute() ([]contract.Subscription, error) {
	s, err := op.database.GetSubscriptionByCategories(op.categories)
	if err != nil {
		return s, err
	}
	if len(s) == 0 {
		return s, db.ErrNotFound
	}
	return s, nil
}

func NewIdExecutor(db SubscriptionLoader, id string) IdExecutor {
	return subscriptionLoadById{
		database: db,
		id:       id,
	}
}

func NewSlugExecutor(db SubscriptionLoader, slug string) IdExecutor {
	return subscriptionLoadBySlug{
		database: db,
		slug:     slug,
	}
}

func NewCategoriesExecutor(db SubscriptionLoader, categories []string) CollectionExecutor {
	return subscriptionsLoadByCategories{
		database:   db,
		categories: categories,
	}
}
