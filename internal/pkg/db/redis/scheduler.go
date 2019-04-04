/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package redis

import (
	"encoding/json"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/redis/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/imdario/mergo"
)

// Return all the schedule interval(s)
func (c *Client) Intervals() (intervals []contract.Interval, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, models.IntervalKey, 0, -1)
	if err != nil {
		return []contract.Interval{}, err
	}

	intervals = make([]contract.Interval, len(objects))
	for i, object := range objects {
		err = json.Unmarshal(object, &intervals[i])
		if err != nil {
			return []contract.Interval{}, err
		}
	}

	return intervals, nil
}

// Return schedule interval(s) up to the number specified
func (c *Client) IntervalsWithLimit(limit int) (intervals []contract.Interval, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, models.IntervalKey, 0, limit-1)
	if err != nil {
		if err != redis.ErrNil {
			return intervals, err
		}
	}

	intervals = make([]contract.Interval, len(objects))
	for i, object := range objects {
		err = json.Unmarshal(object, &intervals[i])
		if err != nil {
			return []contract.Interval{}, err
		}
	}

	return intervals, nil
}

// Return schedule interval by name
func (c *Client) IntervalByName(name string) (interval contract.Interval, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err = getObjectByKey(conn, models.IntervalNameKey, name, &interval)
	return interval, err
}

// Return schedule interval by ID
func (c *Client) IntervalById(id string) (interval contract.Interval, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return contract.Interval{}, db.ErrNotFound
	} else if err != nil {
		return contract.Interval{}, err
	}

	err = json.Unmarshal(object, &interval)
	if err != nil {
		return contract.Interval{}, err
	}

	return interval, err
}

// Add a new schedule interval
func (c *Client) AddInterval(from contract.Interval) (id string, err error) {
	interval := models.NewInterval(from)
	if interval.ID != "" {
		_, err = uuid.Parse(interval.ID)
		if err != nil {
			return "", db.ErrInvalidObjectId
		}
	} else {
		interval.ID = uuid.New().String()
	}

	if interval.Created == 0 {
		ts := db.MakeTimestamp()
		interval.Created = ts
		interval.Modified = ts
	}

	data, err := json.Marshal(interval)
	if err != nil {
		return "", err
	}

	conn := c.Pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	addObject(data, interval, interval.ID, conn)
	_, err = conn.Do("EXEC")

	return interval.ID, err
}

// Update a schedule interval
func (c *Client) UpdateInterval(from contract.Interval) (err error) {
	check, err := c.IntervalByName(from.Name)
	if err != nil && err != redis.ErrNil {
		return err
	}
	if err == nil && from.ID != check.ID {
		// IDs are different -> name not unique
		return db.ErrNotUnique
	}

	from.Modified = db.MakeTimestamp()
	err = mergo.Merge(&from, check)
	if err != nil {
		return err
	}

	interval := models.NewInterval(from)
	data, err := json.Marshal(interval)
	if err != nil {
		return err
	}

	conn := c.Pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	deleteObject(interval, interval.ID, conn)
	addObject(data, interval, interval.ID, conn)
	_, err = conn.Do("EXEC")

	return err
}

// Remove schedule interval by ID
func (c *Client) DeleteIntervalById(id string) (err error) {
	check, err := c.IntervalById(id)
	if err != nil {
		if err == db.ErrNotFound {
			return nil
		}
		return
	}

	interval := models.NewInterval(check)
	conn := c.Pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	deleteObject(interval, id, conn)

	_, err = conn.Do("EXEC")

	return err
}

// Scrub all scheduler intervals from the database (only used in test)
func (c *Client) ScrubAllIntervals() (count int, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	cols := []string{models.IntervalKey}

	for _, col := range cols {
		err = unlinkCollection(conn, col)
		if err != nil {
			return -1, err
		}
	}

	return 0, nil
}

// Get all schedule interval action(s)
func (c *Client) IntervalActions() (actions []contract.IntervalAction, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, models.IntervalActionKey, 0, -1)
	if err != nil {
		return []contract.IntervalAction{}, err
	}

	actions = make([]contract.IntervalAction, len(objects))
	for i, object := range objects {
		err = json.Unmarshal(object, &actions[i])
		if err != nil {
			return []contract.IntervalAction{}, err
		}
	}

	return actions, nil
}

// Return schedule interval action(s) up to the number specified
func (c *Client) IntervalActionsWithLimit(limit int) (actions []contract.IntervalAction, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, models.IntervalActionKey, 0, limit-1)
	if err != nil {
		if err != redis.ErrNil {
			return actions, err
		}
	}

	actions = make([]contract.IntervalAction, len(objects))
	for i, object := range objects {
		err = json.Unmarshal(object, &actions[i])
		if err != nil {
			return []contract.IntervalAction{}, err
		}
	}

	return actions, nil
}

// Get all schedule interval action(s) by interval name
func (c *Client) IntervalActionsByIntervalName(name string) (actions []contract.IntervalAction, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, models.IntervalActionParentKey+":"+name)
	if err != nil {
		if err != redis.ErrNil {
			return actions, err
		}
	}

	actions = make([]contract.IntervalAction, len(objects))
	for i, action := range objects {
		err = unmarshalObject(action, &actions[i])
		if err != nil {
			return actions, err
		}
	}
	return actions, err
}

// Get all schedule interval action(s) by target name
func (c *Client) IntervalActionsByTarget(name string) (actions []contract.IntervalAction, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByValue(conn, models.IntervalActionTargetKey+":"+name)
	if err != nil {
		if err != redis.ErrNil {
			return actions, err
		}
	}

	actions = make([]contract.IntervalAction, len(objects))
	for i, action := range objects {
		err = unmarshalObject(action, &actions[i])
		if err != nil {
			return actions, err
		}
	}
	return actions, err
}

// Get schedule interval action by id
func (c *Client) IntervalActionById(id string) (action contract.IntervalAction, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	object, err := redis.Bytes(conn.Do("GET", id))
	if err == redis.ErrNil {
		return contract.IntervalAction{}, db.ErrNotFound
	} else if err != nil {
		return contract.IntervalAction{}, err
	}

	err = json.Unmarshal(object, &action)
	if err != nil {
		return contract.IntervalAction{}, err
	}

	return action, err
}

// Get schedule interval action by name
func (c *Client) IntervalActionByName(name string) (action contract.IntervalAction, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err = getObjectByKey(conn, models.IntervalActionNameKey, name, &action)
	return action, err
}

// Add schedule interval action
func (c *Client) AddIntervalAction(from contract.IntervalAction) (id string, err error) {
	action := models.NewIntervalAction(from)
	if action.ID != "" {
		_, err = uuid.Parse(action.ID)
		if err != nil {
			return "", db.ErrInvalidObjectId
		}
	} else {
		action.ID = uuid.New().String()
	}

	if action.Created == 0 {
		ts := db.MakeTimestamp()
		action.Created = ts
		action.Modified = ts
	}

	data, err := json.Marshal(action)
	if err != nil {
		return "", err
	}

	conn := c.Pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	addObject(data, action, action.ID, conn)
	_, err = conn.Do("EXEC")

	return action.ID, err
}

// Update schedule interval action
func (c *Client) UpdateIntervalAction(from contract.IntervalAction) (err error) {
	check, err := c.IntervalActionByName(from.Name)
	if err != nil && err != redis.ErrNil {
		return err
	}
	if err == nil && from.ID != check.ID {
		// IDs are different -> name not unique
		return db.ErrNotUnique
	}

	from.Modified = db.MakeTimestamp()
	err = mergo.Merge(&from, check)
	if err != nil {
		return err
	}

	action := models.NewIntervalAction(from)
	data, err := json.Marshal(action)
	if err != nil {
		return err
	}

	conn := c.Pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	deleteObject(action, action.ID, conn)
	addObject(data, action, action.ID, conn)
	_, err = conn.Do("EXEC")

	return err
}

// Remove schedule interval action by id
func (c *Client) DeleteIntervalActionById(id string) (err error) {
	check, err := c.IntervalActionById(id)
	if err != nil {
		if err == db.ErrNotFound {
			return nil
		}
		return
	}

	action := models.NewIntervalAction(check)
	conn := c.Pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	deleteObject(action, id, conn)

	_, err = conn.Do("EXEC")

	return err
}

// Scrub all scheduler interval actions from the database data (only used in test)
func (c *Client) ScrubAllIntervalActions() (count int, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	cols := []string{models.IntervalActionKey}

	for _, col := range cols {
		err = unlinkCollection(conn, col)
		if err != nil {
			return -1, err
		}
	}

	return 0, nil
}
