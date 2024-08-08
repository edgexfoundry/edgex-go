//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"
)

const (
	eventTableName = "\"core-data\".event"

	// constants relate to the event struct field names
	deviceNameField = "DeviceName"
	originField     = "Origin"
)

// AllEvents queries the events with the given range, offset, and limit
func (c *Client) AllEvents(offset, limit int) ([]model.Event, errors.EdgeX) {
	ctx := context.Background()

	events, err := queryEvents(ctx, c.ConnPool, sqlQueryContentWithPagination(eventTableName), offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query all events", err)
	}

	return events, nil
}

// AddEvent adds a new event model to DB
func (c *Client) AddEvent(e model.Event) (model.Event, errors.EdgeX) {
	ctx := context.Background()

	if e.Id == "" {
		e.Id = uuid.NewString()
	}
	event := model.Event{
		Id:          e.Id,
		DeviceName:  e.DeviceName,
		ProfileName: e.ProfileName,
		SourceName:  e.SourceName,
		Origin:      e.Origin,
		Tags:        e.Tags,
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return model.Event{}, errors.NewCommonEdgeX(errors.KindContractInvalid, "unable to JSON marshal event model", err)
	}

	_, err = c.ConnPool.Exec(
		ctx,
		sqlInsert(eventTableName, idCol, contentCol),
		e.Id,
		eventBytes,
	)
	if err != nil {
		//if pgClient.IsDupIdError(err) {
		//	return model.Event{}, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("event id '%s' already exists", e.Id), err)
		//}
		return model.Event{}, pgClient.WrapDBError("failed to insert event", err)
	}

	// TODO: readings included in this event will be added to database in the following PRs

	return event, nil
}

// EventById gets an event by id
func (c *Client) EventById(id string) (model.Event, errors.EdgeX) {
	ctx := context.Background()
	var event model.Event

	row := c.ConnPool.QueryRow(ctx, sqlQueryContentById(eventTableName), id)
	err := row.Scan(&event)
	if err != nil {
		return event, pgClient.WrapDBError(fmt.Sprintf("failed to query event with id '%s'", id), err)
	}

	return event, nil
}

// EventTotalCount returns the total count of Event from db
func (c *Client) EventTotalCount() (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCount(eventTableName))
}

// EventCountByDeviceName returns the count of Event associated a specific Device from db
func (c *Client) EventCountByDeviceName(deviceName string) (uint32, errors.EdgeX) {
	sqlStatement, err := sqlQueryCountByJSONField(eventTableName)
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}

	queryMap := map[string]any{deviceNameField: deviceName}
	queryJsonStr, err := pgClient.ConvertMapToJSONString(queryMap)
	if err != nil {
		return 0, errors.NewCommonEdgeXWrapper(err)
	}
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlStatement, queryJsonStr)
}

// EventCountByTimeRange returns the count of Event by time range from db
func (c *Client) EventCountByTimeRange(start int, end int) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByJSONFieldTimeRange(eventTableName, originField), start, end)
}

// EventsByDeviceName query events by offset, limit and device name
func (c *Client) EventsByDeviceName(offset int, limit int, name string) ([]model.Event, errors.EdgeX) {
	sqlStatement, err := sqlQueryContentByJSONField(eventTableName)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	queryMap := map[string]any{deviceNameField: name}
	queryJsonStr, err := pgClient.ConvertMapToJSONString(queryMap)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	events, err := queryEvents(context.Background(), c.ConnPool, sqlStatement, queryJsonStr, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to query events by device '%s'", name), err)
	}
	return events, nil
}

// EventsByTimeRange query events by time range, offset, and limit
func (c *Client) EventsByTimeRange(start int, end int, offset int, limit int) ([]model.Event, errors.EdgeX) {
	ctx := context.Background()
	sqlStatement := sqlQueryContentByJSONFieldTimeRange(eventTableName, originField)

	events, err := queryEvents(ctx, c.ConnPool, sqlStatement, start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return events, nil
}

// DeleteEventById removes an event by id
func (c *Client) DeleteEventById(id string) errors.EdgeX {
	sqlStatement := sqlDeleteById(eventTableName)

	err := deleteEvents(context.Background(), c.ConnPool, sqlStatement, id)
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("failed delete event with id '%s'", id), err)
	}

	// TODO: delete related readings associated to the deleted events

	return nil
}

// DeleteEventsByDeviceName deletes specific device's events and corresponding readings
// This function is implemented to starts up two goroutines to delete readings and events in the background to achieve better performance
func (c *Client) DeleteEventsByDeviceName(deviceName string) errors.EdgeX {
	ctx := context.Background()

	sqlStatement, edgexErr := sqlDeleteByJSONField(eventTableName)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	queryFieldMap := map[string]any{deviceNameField: deviceName}
	queryJsonStr, edgexErr := pgClient.ConvertMapToJSONString(queryFieldMap)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	go func() {
		err := deleteEvents(ctx, c.ConnPool, sqlStatement, queryJsonStr)
		if err != nil {
			c.loggingClient.Errorf("failed delete event with device '%s': %v", deviceName, err)
		}
	}()

	// TODO: delete related readings associated to the deleted events

	return nil
}

// DeleteEventsByAge deletes events and their corresponding readings that are older than age
// This function is implemented to starts up two goroutines to delete readings and events in the background to achieve better performance
func (c *Client) DeleteEventsByAge(age int64) errors.EdgeX {
	ctx := context.Background()
	expireTimestamp := time.Now().UnixNano() - age
	sqlStatement := sqlDeleteByAgeInJSONField(eventTableName, originField)

	go func() {
		err := deleteEvents(ctx, c.ConnPool, sqlStatement, expireTimestamp)
		if err != nil {
			c.loggingClient.Errorf("failed delete event by age '%d' nanoseconds: %v", age, err)
		}
	}()

	// TODO: delete related readings associated to the deleted events

	return nil
}

// queryEvents queries the data rows with given sql statement and passed args, and unmarshal the data rows to the Event model slice
func queryEvents(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]model.Event, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("query failed", err)
	}

	defer rows.Close()

	var events []model.Event
	events, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Event, error) {
		var event model.Event
		err := rows.Scan(&event)

		// TODO: readings data will be added to the event model in the following PRs

		return event, err
	})

	if err != nil {
		return nil, pgClient.WrapDBError("failed to scan events", err)
	}

	return events, nil
}

// deleteEvents delete the data rows with given sql statement and passed args
func deleteEvents(ctx context.Context, connPool *pgxpool.Pool, sqlStatement string, args ...any) errors.EdgeX {
	commandTag, err := connPool.Exec(
		ctx,
		sqlStatement,
		args...,
	)
	if commandTag.RowsAffected() == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "no event found", nil)
	}
	if err != nil {
		return pgClient.WrapDBError("event(s) delete failed", err)
	}
	return nil
}
