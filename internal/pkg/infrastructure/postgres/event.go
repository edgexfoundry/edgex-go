//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	stdErrs "errors"
	"fmt"
	"time"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AllEvents queries the events with the given range, offset, and limit
func (c *Client) AllEvents(offset, limit int) ([]model.Event, errors.EdgeX) {
	ctx := context.Background()

	events, err := queryEvents(ctx, c.ConnPool, sqlQueryAllWithPaginationDescByCol(eventTableName, originCol), offset, limit)
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
	tagsBytes, err := json.Marshal(event.Tags)
	if err != nil {
		return model.Event{}, errors.NewCommonEdgeX(errors.KindServerError, "unable to JSON marshal event tags", err)
	}

	err = pgx.BeginFunc(ctx, c.ConnPool, func(tx pgx.Tx) error {
		// insert event in a transaction
		_, err = tx.Exec(
			ctx,
			sqlInsert(eventTableName, idCol, deviceNameCol, profileNameCol, sourceNameCol, originCol, tagsCol),
			event.Id,
			event.DeviceName,
			event.ProfileName,
			event.SourceName,
			event.Origin,
			tagsBytes,
		)
		if err != nil {
			return pgClient.WrapDBError("failed to insert event", err)
		}

		// insert readings in a transaction
		err = addReadingsInTx(tx, e.Readings, e.Id)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
		return nil
	})

	if err != nil {
		return model.Event{}, errors.NewCommonEdgeXWrapper(err)
	}
	return event, nil
}

// EventById gets an event by id
func (c *Client) EventById(id string) (model.Event, errors.EdgeX) {
	ctx := context.Background()
	var event model.Event

	rows, err := c.ConnPool.Query(ctx, sqlQueryAllById(eventTableName), id)
	if err != nil {
		return event, pgClient.WrapDBError(fmt.Sprintf("failed to query event with id '%s'", id), err)
	}

	event, err = pgx.CollectExactlyOneRow(rows, func(row pgx.CollectableRow) (model.Event, error) {
		e, err := pgx.RowToStructByNameLax[model.Event](row)
		if err != nil {
			return model.Event{}, err
		}

		// query reading by the specific even_id and origin descending
		readings, err := queryReadings(ctx, c.ConnPool, sqlQueryAllAndDescWithConds(readingTableName, originCol, eventIdFKCol), e.Id)
		if err != nil {
			return model.Event{}, err
		}

		e.Readings = readings
		return e, err
	})
	if err != nil {
		if stdErrs.Is(err, pgx.ErrNoRows) {
			return event, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no event with id '%s' found", id), err)
		}
		return event, pgClient.WrapDBError("failed to scan row to event model", err)
	}

	return event, nil
}

// EventTotalCount returns the total count of Event from db
func (c *Client) EventTotalCount() (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCount(eventTableName))
}

// EventCountByDeviceName returns the count of Event associated a specific Device from db
func (c *Client) EventCountByDeviceName(deviceName string) (uint32, errors.EdgeX) {
	sqlStatement := sqlQueryCountByCol(eventTableName, deviceNameCol)

	return getTotalRowsCount(context.Background(), c.ConnPool, sqlStatement, deviceName)
}

// EventCountByTimeRange returns the count of Event by time range from db
func (c *Client) EventCountByTimeRange(start int64, end int64) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountByTimeRangeCol(eventTableName, originCol, nil), start, end)
}

// EventsByDeviceName query events by offset, limit and device name
func (c *Client) EventsByDeviceName(offset int, limit int, name string) ([]model.Event, errors.EdgeX) {
	sqlStatement := sqlQueryAllAndDescWithCondsAndPag(eventTableName, originCol, deviceNameCol)

	events, err := queryEvents(context.Background(), c.ConnPool, sqlStatement, name, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to query events by device '%s'", name), err)
	}
	return events, nil
}

// EventsByTimeRange query events by time range, offset, and limit
func (c *Client) EventsByTimeRange(start int64, end int64, offset int, limit int) ([]model.Event, errors.EdgeX) {
	ctx := context.Background()
	sqlStatement := sqlQueryAllWithPaginationAndTimeRangeDescByCol(eventTableName, originCol, originCol, nil)

	events, err := queryEvents(ctx, c.ConnPool, sqlStatement, start, end, offset, limit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return events, nil
}

// DeleteEventById removes an event by id
func (c *Client) DeleteEventById(id string) errors.EdgeX {
	ctx := context.Background()
	sqlStatement := sqlDeleteById(eventTableName)

	// delete event and readings in a transaction
	err := pgx.BeginFunc(ctx, c.ConnPool, func(tx pgx.Tx) error {
		if err := deleteReadings(ctx, tx, id); err != nil {
			return err
		}

		if err := deleteEvents(ctx, tx, sqlStatement, id); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "failed delete event", err)
	}
	return nil
}

// DeleteEventsByDeviceName deletes specific device's events and corresponding readings
// This function is implemented to starts up two goroutines to delete readings and events in the background to achieve better performance
func (c *Client) DeleteEventsByDeviceName(deviceName string) errors.EdgeX {
	ctx := context.Background()

	sqlStatement := sqlDeleteByColumn(eventTableName, deviceNameCol)

	go func() {
		// delete events and readings in a transaction
		_ = pgx.BeginFunc(ctx, c.ConnPool, func(tx pgx.Tx) error {
			// select the event-ids of the specified device name from event table as the sub-query of deleting readings
			subSqlStatement := sqlQueryFieldsByCol(eventTableName, []string{idCol}, deviceNameCol)
			if err := deleteReadingsBySubQuery(ctx, tx, subSqlStatement, deviceName); err != nil {
				c.loggingClient.Errorf("failed delete readings with device '%s': %v", deviceName, err)
				return err
			}

			err := deleteEvents(ctx, tx, sqlStatement, deviceName)
			if err != nil {
				c.loggingClient.Errorf("failed delete event with device '%s': %v", deviceName, err)
				return err
			}
			return nil
		})
	}()

	return nil
}

// DeleteEventsByAge deletes events and their corresponding readings that are older than age
// This function is implemented to starts up two goroutines to delete readings and events in the background to achieve better performance
func (c *Client) DeleteEventsByAge(age int64) errors.EdgeX {
	ctx := context.Background()
	expireTimestamp := time.Now().UnixNano() - age
	sqlStatement := sqlDeleteTimeRangeByColumn(eventTableName, originCol)

	go func() {
		// delete events and readings in a transaction
		_ = pgx.BeginFunc(ctx, c.ConnPool, func(tx pgx.Tx) error {
			// select the event ids within the origin time range from event table as the sub-query of deleting readings
			subSqlStatement := sqlQueryFieldsByTimeRange(eventTableName, []string{idCol}, originCol)
			if err := deleteReadingsBySubQuery(ctx, tx, subSqlStatement, expireTimestamp); err != nil {
				c.loggingClient.Errorf("failed delete readings by age '%d' nanoseconds: %v", age, err)
				return err
			}

			err := deleteEvents(ctx, tx, sqlStatement, expireTimestamp)
			if err != nil {
				c.loggingClient.Errorf("failed delete event by age '%d' nanoseconds: %v", age, err)
				return err
			}
			return nil
		})
	}()

	return nil
}

// queryEvents queries the data rows with given sql statement and passed args, and unmarshal the data rows to the Event model slice
func queryEvents(ctx context.Context, connPool *pgxpool.Pool, sql string, args ...any) ([]model.Event, errors.EdgeX) {
	rows, err := connPool.Query(ctx, sql, args...)
	if err != nil {
		return nil, pgClient.WrapDBError("query failed", err)
	}

	var events []model.Event
	events, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Event, error) {
		event, err := pgx.RowToStructByNameLax[model.Event](row)
		if err != nil {
			return model.Event{}, err
		}

		// query reading by the specific even_id and origin descending
		readings, err := queryReadings(ctx, connPool, sqlQueryAllAndDescWithConds(readingTableName, originCol, eventIdFKCol), event.Id)

		if err != nil {
			return model.Event{}, err
		}

		event.Readings = readings
		return event, nil
	})

	if err != nil {
		return nil, pgClient.WrapDBError("failed to scan events", err)
	}

	return events, nil
}

// deleteEvents delete the data rows with given sql statement and passed args in a db transaction
func deleteEvents(ctx context.Context, tx pgx.Tx, sqlStatement string, args ...any) errors.EdgeX {
	commandTag, err := tx.Exec(
		ctx,
		sqlStatement,
		args...,
	)
	if err != nil {
		return pgClient.WrapDBError("event(s) delete failed", err)
	}
	if commandTag.RowsAffected() == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "no event found", nil)
	}
	return nil
}
