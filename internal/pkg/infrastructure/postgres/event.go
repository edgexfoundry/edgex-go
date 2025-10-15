//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	stdErrs "errors"
	"fmt"
	"time"

	pgClient "github.com/edgexfoundry/edgex-go/internal/pkg/db/postgres"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	model "github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AllEvents queries the events with the given range, offset, and limit
func (c *Client) AllEvents(offset, limit int) ([]model.Event, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	events, err := queryEvents(ctx, c.ConnPool, sqlQueryAllEventWithPaginationDescByCol(originCol), offset, validLimit)
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

	deviceInfoId, err := c.deviceInfoIdByEvent(e)
	if err != nil {
		return model.Event{}, errors.NewCommonEdgeXWrapper(err)
	}

	event := model.Event{
		Id:          e.Id,
		DeviceName:  e.DeviceName,
		ProfileName: e.ProfileName,
		SourceName:  e.SourceName,
		Origin:      e.Origin,
		Tags:        e.Tags,
	}

	pgxErr := pgx.BeginFunc(ctx, c.ConnPool, func(tx pgx.Tx) error {
		// insert event in a transaction
		_, pgxErr := tx.Exec(
			ctx,
			sqlInsert(eventTableName, idCol, deviceInfoIdFKCol, originCol),
			event.Id,
			deviceInfoId,
			event.Origin,
		)
		if pgxErr != nil {
			return pgClient.WrapDBError("failed to insert event", pgxErr)
		}

		// insert readings in a transaction
		pgxErr = c.addReadingsInTx(tx, e.Readings, e.Id)
		if pgxErr != nil {
			return errors.NewCommonEdgeXWrapper(pgxErr)
		}
		return nil
	})

	if pgxErr != nil {
		return model.Event{}, errors.NewCommonEdgeXWrapper(pgxErr)
	}

	// return the event with readings to ensure readingsPersistedCounter will be increased
	event.Readings = e.Readings
	return event, nil
}

// EventById gets an event by id
func (c *Client) EventById(id string) (model.Event, errors.EdgeX) {
	ctx := context.Background()
	var event model.Event

	rows, err := c.ConnPool.Query(ctx, sqlQueryAllEventById(), id)
	if err != nil {
		return event, pgClient.WrapDBError(fmt.Sprintf("failed to query event with id '%s'", id), err)
	}

	event, err = pgx.CollectExactlyOneRow(rows, func(row pgx.CollectableRow) (model.Event, error) {
		e, err := pgx.RowToStructByNameLax[model.Event](row)
		if err != nil {
			return model.Event{}, err
		}

		// query reading by the specific even_id and origin descending
		readings, err := queryReadings(ctx, c.ConnPool, sqlQueryAllReadingAndDescWithConds(originCol, eventIdFKCol), e.Id)
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
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountEvent())
}

// EventCountByDeviceName returns the count of Event associated a specific Device from db
func (c *Client) EventCountByDeviceName(deviceName string) (uint32, errors.EdgeX) {
	sqlStatement := sqlQueryCountEventByCol(deviceNameCol)

	return getTotalRowsCount(context.Background(), c.ConnPool, sqlStatement, deviceName)
}

// EventCountByTimeRange returns the count of Event by time range from db
func (c *Client) EventCountByTimeRange(start int64, end int64) (uint32, errors.EdgeX) {
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlQueryCountEventByTimeRangeCol(originCol, nil), start, end)
}

// EventCountByDeviceNameAndSourceNameAndLimit returns the count of Event by deviceName, resourceName, and limit from db
// this is used to check whether the event number is reach the event retention maxCap
func (c *Client) EventCountByDeviceNameAndSourceNameAndLimit(deviceName, sourceName string, limit int) (uint32, errors.EdgeX) {
	sqlStatement := sqlCountEventByDeviceNameAndSourceNameAndLimit()
	return getTotalRowsCount(context.Background(), c.ConnPool, sqlStatement, deviceName, sourceName, limit)
}

// EventsByDeviceName query events by offset, limit and device name
func (c *Client) EventsByDeviceName(offset int, limit int, name string) ([]model.Event, errors.EdgeX) {
	offset, validLimit := getValidOffsetAndLimit(offset, limit)

	sqlStatement := sqlQueryAllEventAndDescWithCondsAndPage(originCol, deviceNameCol)

	events, err := queryEvents(context.Background(), c.ConnPool, sqlStatement, name, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindDatabaseError, fmt.Sprintf("failed to query events by device '%s'", name), err)
	}
	return events, nil
}

// EventsByTimeRange query events by time range, offset, and limit
func (c *Client) EventsByTimeRange(start int64, end int64, offset int, limit int) ([]model.Event, errors.EdgeX) {
	ctx := context.Background()
	offset, validLimit := getValidOffsetAndLimit(offset, limit)
	sqlStatement := sqlQueryAllEventWithPaginationAndTimeRangeDescByCol(originCol, originCol, nil)

	events, err := queryEvents(ctx, c.ConnPool, sqlStatement, start, end, offset, validLimit)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return events, nil
}

// DeleteEventById removes an event by id
func (c *Client) DeleteEventById(id string) errors.EdgeX {
	ctx := context.Background()
	sqlStatement := sqlDeleteById(eventTableName)

	event, err := c.EventById(id)
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "failed delete event", err)
	}

	// delete event and readings in a transaction
	pgxErr := pgx.BeginFunc(ctx, c.ConnPool, func(tx pgx.Tx) error {
		// use event.origin to filter readings to improve delete speed
		if err = deleteReadingsByOriginAndEventId(ctx, tx, event.Origin, id); err != nil {
			return err
		}

		if err = deleteEvents(ctx, tx, sqlStatement, id); err != nil {
			return err
		}
		return nil
	})
	if pgxErr != nil {
		return errors.NewCommonEdgeX(errors.Kind(pgxErr), "failed delete event", pgxErr)
	}
	return nil
}

// DeleteEventsByDeviceName deletes specific device's events and corresponding readings
// This function is implemented to starts up two goroutines to delete readings and events in the background to achieve better performance
func (c *Client) DeleteEventsByDeviceName(deviceName string) errors.EdgeX {
	// update deviceInfo as deletable, then event and reading will not return when the user query event or reading by the specific device
	if err := c.updateDeviceInfosDeletableByDeviceName(deviceName); err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "delete events by deviceName", err)
	}
	// delete events, readings, deviceInfos
	if err := c.deleteEventsByConditions([]string{deviceNameCol}, []any{deviceName}); err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "delete events by deviceName", err)
	}
	return nil
}

// DeleteEventsByDeviceNameAndSourceName deletes specific device's events and corresponding readings
func (c *Client) DeleteEventsByDeviceNameAndSourceName(deviceName, sourceName string) errors.EdgeX {
	return c.deleteEventsByConditions([]string{deviceNameCol, sourceNameCol}, []any{deviceName, sourceName})
}

// deleteEventsByConditions deletes specific device's events and corresponding readings
func (c *Client) deleteEventsByConditions(cols []string, values []any) errors.EdgeX {
	ctx := context.Background()

	sqlStatement := sqlDeleteEventsByColumn(cols...)

	go func() {
		// deviceInfos are used to remove data by id and remove the id value from the cache after finishing the transaction
		deviceInfos, err := c.deviceInfosByConds(cols, values)
		if err != nil {
			return
		}
		// delete events and readings in a transaction
		pgxErr := pgx.BeginFunc(ctx, c.ConnPool, func(tx pgx.Tx) error {
			// select the event-ids of the specified device name from event table as the sub-query of deleting readings
			subSqlStatement := sqlQueryEventIdFieldsByCol(cols...)
			if err = deleteReadingsBySubQuery(ctx, tx, subSqlStatement, values...); err != nil {
				c.loggingClient.Errorf("failed delete readings with conditions '%v' '%v': %v", cols, values, err)
				return err
			}

			err = deleteEvents(ctx, tx, sqlStatement, values...)
			if err != nil {
				c.loggingClient.Errorf("failed delete event with conditions '%v' '%v': %v", cols, values, err)
				return err
			}

			for _, deviceInfo := range deviceInfos {
				err = deleteDeviceInfoById(ctx, tx, deviceInfo.Id)
				if err != nil {
					return errors.NewCommonEdgeXWrapper(err)
				}
			}
			return nil
		})
		if pgxErr != nil {
			c.loggingClient.Errorf("failed delete events with conditions '%v' '%v': %v", cols, values, pgxErr)
			return
		}

		for _, deviceInfo := range deviceInfos {
			c.deviceInfoIdCache.Remove(deviceInfo)
		}
	}()

	return nil
}

// DeleteEventsByAge deletes events and their corresponding readings that are older than age
// This function is implemented to starts up two goroutines to delete readings and events in the background to achieve better performance
func (c *Client) DeleteEventsByAge(age int64) errors.EdgeX {
	return c.deleteEventsByAgeAndConditions(age, nil, nil)
}

// DeleteEventsByAgeAndDeviceNameAndSourceName deletes events and their corresponding readings that are older than age
// This function is implemented to starts up two goroutines to delete readings and events in the background to achieve better performance
func (c *Client) DeleteEventsByAgeAndDeviceNameAndSourceName(age int64, deviceName, sourceName string) errors.EdgeX {
	return c.deleteEventsByAgeAndConditions(age, []string{deviceNameCol, sourceNameCol}, []any{deviceName, sourceName})
}

// deleteEventsByAgeAndConditions deletes events and their corresponding readings that are older than age
// This function is implemented to starts up two goroutines to delete readings and events in the background to achieve better performance
func (c *Client) deleteEventsByAgeAndConditions(age int64, cols []string, values []any) errors.EdgeX {
	ctx := context.Background()
	expireTimestamp := time.Now().UnixNano() - age
	sqlStatement := sqlDeleteEventsByTimeRangeAndColumn(originCol, cols...)
	args := append([]any{expireTimestamp}, values...)

	go func() {
		// delete events and readings in a transaction
		_ = pgx.BeginFunc(ctx, c.ConnPool, func(tx pgx.Tx) error {
			// select the event ids within the origin time range from event table as the sub-query of deleting readings
			subSqlStatement := sqlQueryEventIdFieldByTimeRangeAndConditions(originCol, cols...)
			if err := deleteReadingsBySubQuery(ctx, tx, subSqlStatement, args...); err != nil {
				c.loggingClient.Errorf("failed delete readings by age '%d' nanoseconds: %v", age, err)
				return err
			}

			err := deleteEvents(ctx, tx, sqlStatement, args...)
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
		readings, err := queryReadings(ctx, connPool, sqlQueryAllReadingAndDescWithConds(originCol, eventIdFKCol), event.Id)

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

func (c *Client) LatestEventByDeviceNameAndSourceNameAndOffset(deviceName, sourceName string, offset uint32) (model.Event, errors.EdgeX) {
	ctx := context.Background()

	events, err := queryEvents(
		ctx, c.ConnPool,
		sqlQueryAllEventAndDescWithCondsAndPage(originCol, deviceNameCol, sourceNameCol),
		deviceName, sourceName, offset, 1,
	)
	if err != nil {
		return model.Event{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query all events", err)
	}

	if len(events) == 0 {
		return model.Event{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no event found with offset '%d'", offset), err)
	}
	return events[0], nil
}

// LatestEventByDeviceNameAndSourceNameAndAgeAndOffset query an event with specified conditions and age and offset
func (c *Client) LatestEventByDeviceNameAndSourceNameAndAgeAndOffset(deviceName, sourceName string, age int64, offset uint32) (model.Event, errors.EdgeX) {
	ctx := context.Background()
	expireTimestamp := time.Now().UnixNano() - age

	sqlStmt := sqlQueryAllEventAndDescWithCondsAndPagAndUpperLimitTime(originCol, originCol, deviceNameCol, sourceNameCol)
	events, err := queryEvents(
		ctx, c.ConnPool, sqlStmt,
		expireTimestamp, deviceName, sourceName, offset, 1,
	)
	if err != nil {
		return model.Event{}, errors.NewCommonEdgeX(errors.KindDatabaseError, "failed to query all events", err)
	}

	if len(events) == 0 {
		return model.Event{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("no event found with offset '%d'", offset), err)
	}
	return events[0], nil
}
