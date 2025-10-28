//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"
)

// AddSubscription adds a new subscription to the database
func (c *Client) AddSubscription(s models.Subscription) (models.Subscription, errors.EdgeX) {
	ctx := context.Background()
	if len(s.Id) == 0 {
		s.Id = uuid.New().String()
	}

	exists, edgexErr := checkSubscriptionExists(ctx, c.ConnPool, s.Name)
	if edgexErr != nil {
		return s, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	if exists {
		return s, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("subscription name %s already exists", s.Name), nil)
	}

	timestamp := time.Now().UTC().UnixMilli()
	s.Created = timestamp
	s.Modified = timestamp
	dataBytes, err := json.Marshal(s)
	if err != nil {
		return s, errors.NewCommonEdgeX(errors.KindServerError, "failed to marshal Subscription model", err)
	}

	_, err = c.ConnPool.Exec(ctx, sqlInsert(subscriptionTableName, idCol, contentCol), s.Id, dataBytes)
	if err != nil {
		return s, pgClient.WrapDBError("failed to insert row to subscription table", err)
	}
	return s, nil
}

// SubscriptionById queries the subscription by id
func (c *Client) SubscriptionById(id string) (models.Subscription, errors.EdgeX) {
	subscription, err := querySubscription(context.Background(), c.ConnPool, sqlQueryContentById(subscriptionTableName), id)
	if err != nil {
		return subscription, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query subscription by id %s", id), err)
	}

	return subscription, nil
}

// AllSubscriptions queries subscriptions with the given offset, and limit
func (c *Client) AllSubscriptions(offset, limit int) ([]models.Subscription, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	subscriptions, err := querySubscriptions(context.Background(), c.ConnPool, sqlQueryContentWithPaginationAsNamedArgs(subscriptionTableName), pgx.NamedArgs{offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all subscriptions", err)
	}

	return subscriptions, nil
}

// SubscriptionByName queries the subscription by name
func (c *Client) SubscriptionByName(name string) (models.Subscription, errors.EdgeX) {
	queryObj := map[string]any{nameField: name}
	subscription, err := querySubscription(context.Background(), c.ConnPool, sqlQueryContentByJSONField(subscriptionTableName), queryObj)
	if err != nil {
		return subscription, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query subscription by name %s", name), err)
	}

	return subscription, nil
}

// SubscriptionsByCategory queries the subscription by category
func (c *Client) SubscriptionsByCategory(offset, limit int, category string) ([]models.Subscription, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{categoriesField: []string{category}}

	subscriptions, err := querySubscriptions(context.Background(), c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(subscriptionTableName),
		pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return subscriptions, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query subscription by category %s", category), err)
	}

	return subscriptions, nil
}

// SubscriptionsByLabel queries the subscription by label
func (c *Client) SubscriptionsByLabel(offset, limit int, label string) ([]models.Subscription, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{labelsField: []string{label}}

	subscriptions, err := querySubscriptions(context.Background(), c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(subscriptionTableName),
		pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return subscriptions, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query subscription by label %s", label), err)
	}

	return subscriptions, nil
}

// SubscriptionsByReceiver queries the subscription by receiver
func (c *Client) SubscriptionsByReceiver(offset, limit int, receiver string) ([]models.Subscription, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	queryObj := map[string]any{receiverField: receiver}

	subscriptions, err := querySubscriptions(context.Background(), c.ConnPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(subscriptionTableName),
		pgx.NamedArgs{jsonContentCondition: queryObj, offsetCondition: offset, limitCondition: validLimit})
	if err != nil {
		return subscriptions, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed to query subscription by receiver %s", receiver), err)
	}

	return subscriptions, nil
}

// DeleteSubscriptionByName deletes the subscription by name
func (c *Client) DeleteSubscriptionByName(name string) errors.EdgeX {
	queryObj := map[string]any{nameField: name}
	_, err := c.ConnPool.Exec(context.Background(), sqlDeleteByJSONField(subscriptionTableName), queryObj)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to delete subscription by name %s", name), err)
	}
	return nil
}

// UpdateSubscription updates the subscription
func (c *Client) UpdateSubscription(s models.Subscription) errors.EdgeX {
	modified := time.Now().UTC().UnixMilli()
	s.Modified = modified

	dataBytes, err := json.Marshal(s)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to marshal Subscription model", err)
	}

	_, err = c.ConnPool.Exec(context.Background(), sqlUpdateContentById(subscriptionTableName), dataBytes, s.Id)
	if err != nil {
		return pgClient.WrapDBError(fmt.Sprintf("failed to update row by subscription id '%s' from subscription table", s.Id), err)
	}

	return nil
}

// SubscriptionsByCategoriesAndLabels queries the subscription by categories and labels
func (c *Client) SubscriptionsByCategoriesAndLabels(offset, limit int, categories []string, labels []string) ([]models.Subscription, errors.EdgeX) {
	subscriptions, err := subscriptionsByCategoriesAndLabels(c.ConnPool, offset, limit, categories, labels)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), "failed to query all subscriptions by categories and labels", err)
	}

	return subscriptions, nil
}

// SubscriptionTotalCount returns the total count of subscriptions
func (c *Client) SubscriptionTotalCount() (int64, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCount(subscriptionTableName))
}

// SubscriptionCountByCategory returns the count of subscriptions by category
func (c *Client) SubscriptionCountByCategory(category string) (int64, errors.EdgeX) {
	queryObj := map[string]any{categoriesField: []string{category}}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByJSONField(subscriptionTableName), queryObj)
}

// SubscriptionCountByLabel returns the count of subscriptions by label
func (c *Client) SubscriptionCountByLabel(label string) (int64, errors.EdgeX) {
	queryObj := map[string]any{labelsField: []string{label}}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByJSONField(subscriptionTableName), queryObj)
}

// SubscriptionCountByReceiver returns the count of subscriptions by receiver
func (c *Client) SubscriptionCountByReceiver(receiver string) (int64, errors.EdgeX) {
	queryObj := map[string]any{receiverField: receiver}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByJSONField(subscriptionTableName), queryObj)
}

func querySubscription(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) (models.Subscription, errors.EdgeX) {
	var subscription models.Subscription
	row := connPool.QueryRow(ctx, sql, args...)

	if err := row.Scan(&subscription); err != nil {
		return subscription, pgClient.WrapDBError("failed to query subscription", err)
	}
	return subscription, nil
}

func querySubscriptions(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]models.Subscription, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("failed to query rows from subscription table", err)
	}

	subscriptions, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Subscription, error) {
		var s models.Subscription
		scanErr := row.Scan(&s)
		return s, scanErr
	})
	if err != nil {
		return nil, pgClient.WrapDBError("failed to collect rows to Subscription model", err)
	}

	return subscriptions, nil
}

func checkSubscriptionExists(ctx context.Context, connPool *pgxpool.Pool, name string) (bool, errors.EdgeX) {
	var exists bool
	queryObj := map[string]any{nameField: name}
	err := connPool.QueryRow(ctx, sqlCheckExistsByJSONField(subscriptionTableName), queryObj).Scan(&exists)
	if err != nil {
		return false, pgClient.WrapDBError(fmt.Sprintf("failed to query row by name '%s' from subscription table", name), err)
	}
	return exists, nil
}

func subscriptionsByCategoriesAndLabels(connPool *pgxpool.Pool, offset, limit int, categories []string, labels []string) (subscriptions []models.Subscription, err errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	categoriesObj := map[string]any{categoriesField: categories}
	labelsObj := map[string]any{labelsField: labels}

	switch {
	case len(labels) == 0:
		subscriptions, err = querySubscriptions(context.Background(), connPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(subscriptionTableName),
			pgx.NamedArgs{jsonContentCondition: categoriesObj, offsetCondition: offset, limitCondition: validLimit})
	case len(categories) == 0:
		subscriptions, err = querySubscriptions(context.Background(), connPool, sqlQueryContentByJSONFieldWithPaginationAsNamedArgs(subscriptionTableName),
			pgx.NamedArgs{jsonContentCondition: labelsObj, offsetCondition: offset, limitCondition: validLimit})
	default:
		sql := fmt.Sprintf(`
			SELECT content
				FROM (
	    			SELECT content, COALESCE((content->>'%s')::bigint, 0) AS sort_key
						FROM %s 
						WHERE content @> @%s::jsonb
					INTERSECT
					SELECT content, COALESCE((content->>'%s')::bigint, 0) AS sort_key 
						FROM %s 
						WHERE content @> @%s::jsonb
				)		
			ORDER BY sort_key OFFSET @%s LIMIT @%s;
	`, createdField, subscriptionTableName, categoryCondition, createdField, subscriptionTableName, labelsCondition, offsetCondition, limitCondition)
		subscriptions, err = querySubscriptions(context.Background(), connPool, sql,
			pgx.NamedArgs{categoryCondition: categoriesObj, labelsCondition: labelsObj, offsetCondition: offset, limitCondition: validLimit})
	}
	if err != nil {
		return subscriptions, errors.NewCommonEdgeXWrapper(err)
	}
	return subscriptions, nil
}
