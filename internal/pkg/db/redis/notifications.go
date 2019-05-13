/*******************************************************************************
 * Copyright (C) 2018 IOTech Ltd
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
	"fmt"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// ******************************* NOTIFICATIONS **********************************
func (c Client) AddNotification(n contract.Notification) (string, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err := addNotification(conn, &n)
	if err != nil {
		return "", err
	}
	return n.ID, nil
}

func (c Client) UpdateNotification(n contract.Notification) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteNotification(conn, n.ID)
	if err != nil {
		return err
	}
	n.Modified = db.MakeTimestamp()
	return addNotification(conn, &n)
}

// Get all notifications
func (c Client) GetNotifications() (n []contract.Notification, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Notification, 0, -1)
	if err != nil {
		return nil, err
	}

	n, err = unmarshalNotifications(objects)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (c Client) GetNotificationById(id string) (notification contract.Notification, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err = getObjectById(conn, id, unmarshalObject, &notification)
	if err != nil {
		return notification, err
	}
	return notification, nil
}

func (c Client) GetNotificationBySlug(slug string) (notification contract.Notification, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	notification, err = notificationBySlug(conn, slug)
	if err != nil {
		if err == redis.ErrNil {
			return notification, db.ErrNotFound
		}
		return notification, err
	}

	return notification, nil
}

func (c Client) GetNotificationBySender(sender string, limit int) ([]contract.Notification, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Notification+":sender:"+sender, 0, limit-1)
	if err != nil {
		return nil, err
	}

	notifications, err := unmarshalNotifications(objects)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (c Client) GetNotificationsByLabels(labels []string, limit int) (notifications []contract.Notification, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	for _, label := range labels {
		var objects, err = getObjectsByRange(conn, db.Notification+":label:"+label, 0, limit-1)
		if err != nil {
			if err != redis.ErrNil {
				return notifications, err
			}
		}

		t, err := unmarshalNotifications(objects)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, t...)

		limit -= len(objects)
		if limit < 0 {
			break
		}
	}
	return notifications, nil
}

func (c Client) GetNotificationsByStartEnd(start int64, end int64, limit int) ([]contract.Notification, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByScore(conn, db.Notification+":created", start, end, limit)
	if err != nil {
		return nil, err
	}
	notifications, err := unmarshalNotifications(objects)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (c Client) GetNotificationsByStart(start int64, limit int) ([]contract.Notification, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByScore(conn, db.Notification+":created", start, -1, limit)
	if err != nil {
		return nil, err
	}
	notifications, err := unmarshalNotifications(objects)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (c Client) GetNotificationsByEnd(end int64, limit int) ([]contract.Notification, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByScore(conn, db.Notification+":created", 0, end, limit)
	if err != nil {
		return nil, err
	}
	notifications, err := unmarshalNotifications(objects)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (c Client) GetNewNotifications(limit int) ([]contract.Notification, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	//REF: using NEW, based on Mongo impln
	objects, err := getObjectsByScore(conn, db.Notification+":status:"+contract.New, 0, -1, limit)
	if err != nil {
		return nil, err
	}

	notifications, err := unmarshalNotifications(objects)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (c Client) GetNewNormalNotifications(limit int) ([]contract.Notification, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	//option 2: sorted set of status+severity should be created and get objects based on
	objects, err := getObjectsByValuesSorted(conn, limit, db.Notification+":status:"+contract.New, db.Notification+":severity:"+contract.Normal)
	if err != nil {
		return nil, err
	}
	notifications, err := unmarshalNotifications(objects)
	if err != nil {
		return nil, err
	}

	return notifications, nil

}

func (c Client) MarkNotificationProcessed(n contract.Notification) error {
	conn := c.Pool.Get()
	defer conn.Close()

	n.Status = contract.NotificationsStatus(contract.Processed)
	return c.UpdateNotification(n)
}

func (c Client) DeleteNotificationById(id string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	n, err := c.GetNotificationById(id)
	if err != nil {
		return nil
	}

	err = c.deleteTransmissionBySlug(conn, n.Slug)
	if err != nil {
		return err
	}

	return deleteNotification(conn, id)
}

func (c Client) DeleteNotificationBySlug(slug string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	n, err := notificationBySlug(conn, slug)
	if err != nil {
		return err
	}

	err = c.deleteTransmissionBySlug(conn, n.Slug)
	if err != nil {
		return err
	}

	return deleteNotification(conn, n.ID)
}

// DeleteNotificationsOld remove all the notifications that are older than the given age
func (c Client) DeleteNotificationsOld(age int) error {
	conn := c.Pool.Get()
	defer conn.Close()

	currentTime := db.MakeTimestamp()
	end := int64(int(currentTime) - age)

	objects, err := getObjectsByScore(conn, db.Notification+":modified", 0, end, 0)

	for _, object := range objects {
		if len(object) > 0 {
			var n contract.Notification
			err := unmarshalObject(object, &n)
			if err != nil {
				return err
			}
			// Delete processed notification
			if n.Status != contract.Processed {
				continue
			}
			err = deleteNotification(conn, n.ID)
			if err != nil {
				return err
			}
		}
	}
	return err
}

// ******************************* SUBSCRIPTIONS **********************************
func (c Client) AddSubscription(s contract.Subscription) (string, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err := addSubscription(conn, &s)
	if err != nil {
		return "", err
	}

	return s.ID, nil
}

func (c Client) UpdateSubscription(s contract.Subscription) error {
	conn := c.Pool.Get()
	defer conn.Close()

	// update modified, delete and add subscription
	err := deleteSubscription(conn, s.ID)
	if err != nil {
		return nil
	}

	s.Modified = db.MakeTimestamp()
	return addSubscription(conn, &s)
}

func (c Client) GetSubscriptions() ([]contract.Subscription, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Subscription, 0, -1)
	if err != nil {
		return nil, err
	}

	s, err := unmarshalSubscriptions(objects)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (c Client) GetSubscriptionById(id string) (s contract.Subscription, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err = getObjectById(conn, id, unmarshalObject, &s)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (c Client) GetSubscriptionBySlug(slug string) (s contract.Subscription, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	s, err = subscriptionBySlug(conn, slug)
	if err != nil {
		return s, err
	}
	return s, nil
}

func (c Client) GetSubscriptionByReceiver(receiver string) ([]contract.Subscription, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Subscription+":receiver:"+receiver, 0, -1)
	if err != nil {
		return nil, err
	}

	s, err := unmarshalSubscriptions(objects)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (c Client) GetSubscriptionByCategories(categories []string) (s []contract.Subscription, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	args := make([]string, len(categories))
	for i, c := range categories {
		args[i] = db.Subscription + ":category:" + c
	}

	objects, err := getUnionObjectsByValues(conn, args...)
	if err != nil {
		return s, err
	}

	s, err = unmarshalSubscriptions(objects)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (c Client) GetSubscriptionByLabels(labels []string) (s []contract.Subscription, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	args := make([]string, len(labels))
	for i, label := range labels {
		args[i] = db.Subscription + ":label:" + label
	}

	objects, err := getUnionObjectsByValues(conn, args...)
	if err != nil {
		return s, err
	}

	s, err = unmarshalSubscriptions(objects)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (c Client) GetSubscriptionByCategoriesLabels(categories []string, labels []string) (s []contract.Subscription, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	var args []string
	for _, c := range categories {
		args = append(args, db.Subscription+":category:"+c)
	}
	for _, label := range labels {
		args = append(args, db.Subscription+":label:"+label)
	}

	objects, err := getUnionObjectsByValues(conn, args...)
	if err != nil {
		return s, err
	}

	s, err = unmarshalSubscriptions(objects)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (c Client) DeleteSubscriptionBySlug(slug string) error {
	conn := c.Pool.Get()
	defer conn.Close()

	s, err := subscriptionBySlug(conn, slug)
	if err != nil {
		return err
	}

	return deleteSubscription(conn, s.ID)

}

// ******************************* TRANSMISSIONS **********************************
func (c Client) AddTransmission(t contract.Transmission) (string, error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err := addTransmission(conn, &t)
	if err != nil {
		return "", err
	}

	return t.ID, nil
}

func (c Client) UpdateTransmission(t contract.Transmission) error {
	conn := c.Pool.Get()
	defer conn.Close()

	err := deleteTransmission(conn, t.ID)
	if err != nil {
		return nil
	}

	t.Modified = db.MakeTimestamp()
	return addTransmission(conn, &t)
}

func (c Client) GetTransmissionsByNotificationSlug(slug string, resendLimit int) (transmissions []contract.Transmission, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByScore(conn, db.Transmission+":slug:"+slug, 0, int64(resendLimit), -1)
	if err != nil {
		return transmissions, err
	}

	transmissions, err = unmarshalTransmissions(objects)
	if err != nil {
		return transmissions, err
	}

	return transmissions, nil
}

func (c Client) GetTransmissionsByStartEnd(start int64, end int64, resendLimit int) (transmissions []contract.Transmission, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getTransmissionsByStartEndWithResendLimit(conn, start, end, resendLimit)
	if err != nil {
		return transmissions, err
	}

	transmissions, err = unmarshalTransmissions(objects)
	if err != nil {
		return transmissions, err
	}

	return transmissions, nil
}

func (c Client) GetTransmissionsByStart(start int64, resendLimit int) (transmissions []contract.Transmission, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getTransmissionsByStartEndWithResendLimit(conn, start, -1, resendLimit)
	if err != nil {
		return transmissions, err
	}

	transmissions, err = unmarshalTransmissions(objects)
	if err != nil {
		return transmissions, err
	}

	return transmissions, nil
}

func (c Client) GetTransmissionById(id string) (transmission contract.Transmission, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	err = getObjectById(conn, id, unmarshalObject, &transmission)
	if err != nil {
		return transmission, err
	}
	return transmission, nil
}

func (c Client) GetTransmissionsByEnd(end int64, resendLimit int) (transmissions []contract.Transmission, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getTransmissionsByStartEndWithResendLimit(conn, -1, end, resendLimit)
	if err != nil {
		return transmissions, err
	}

	transmissions, err = unmarshalTransmissions(objects)
	if err != nil {
		return transmissions, err
	}

	return transmissions, nil
}

func (c Client) GetTransmissionsByStatus(resendLimit int, status contract.TransmissionStatus) (transmissions []contract.Transmission, err error) {
	conn := c.Pool.Get()
	defer conn.Close()

	objects, err := getObjectsByRange(conn, db.Transmission+":status:"+string(status), 0, resendLimit)
	if err != nil {
		return transmissions, err
	}

	transmissions, err = unmarshalTransmissions(objects)
	if err != nil {
		return transmissions, err
	}

	return transmissions, nil
}

// DeleteTransmission delete old transmission with specified status
func (c Client) DeleteTransmission(age int64, status contract.TransmissionStatus) error {
	conn := c.Pool.Get()
	defer conn.Close()

	currentTime := db.MakeTimestamp()
	end := currentTime - age

	objects, err := getObjectsByRangeFilter(conn, db.Transmission+":modified", db.Transmission+":status:"+fmt.Sprintf("%s", status), 0, int(end))
	if err != nil {
		return err
	}
	transmissions, err := unmarshalTransmissions(objects)
	if err != nil {
		return err
	}

	for _, transmission := range transmissions {
		err = deleteTransmission(conn, transmission.ID)
		if err != nil {
			return err
		}
	}
	return err
}

// Cleanup delete all notifications and associated transmissions
func (c Client) Cleanup() error {
	//conn := c.Pool.Get()
	//defer conn.Close()
	//
	//cols := []string{
	//	db.Transmission, db.Notification,
	//}
	//
	//for _, col := range cols {
	//	err := unlinkCollection(conn, col)
	//	if err != nil {
	//		return err
	//	}
	//}
	err := c.CleanupOld(0)
	if err != nil {
		return err
	}
	return nil
}

// Cleanup delete old notifications and associated transmissions
func (c Client) CleanupOld(age int) error {
	conn := c.Pool.Get()
	defer conn.Close()

	currentTime := db.MakeTimestamp()
	end := currentTime - int64(age)
	objects, err := getObjectsByScore(conn, db.Notification+":created", 0, end, -1)
	if err != nil {
		return err
	}

	notifications, err := unmarshalNotifications(objects)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		err = c.DeleteNotificationById(notification.ID)
		if err != nil {
			return err
		}
		transmissions, err := c.GetTransmissionsByNotificationSlug(notification.Slug, -1)
		if err != nil {
			return err
		}
		for _, transmission := range transmissions {
			err = deleteTransmission(conn, transmission.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ************************** HELPER FUNCTIONS ***************************
func addNotification(conn redis.Conn, n *contract.Notification) error {
	exist, err := redis.Bool(conn.Do("HEXISTS", db.Notification+":slug", n.Slug))
	if err != nil {
		return err
	} else if exist {
		return errors.Errorf("%v, slug=%v", db.ErrNotUnique, n.Slug)
	}

	if n.Created == 0 {
		n.Created = db.MakeTimestamp()
		n.Modified = n.Created
	}

	if n.ID == "" {
		n.ID = uuid.New().String()
	}

	m, err := marshalObject(n)
	if err != nil {
		return err
	}
	id := n.ID

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", id, m)
	_ = conn.Send("ZADD", db.Notification, 0, id)
	_ = conn.Send("HSET", db.Notification+":slug", n.Slug, id)
	_ = conn.Send("ZADD", db.Notification+":sender:"+n.Sender, 0, id)
	_ = conn.Send("ZADD", db.Notification+":status:"+n.Status, 0, id)
	_ = conn.Send("ZADD", db.Notification+":severity:"+n.Severity, 0, id)
	_ = conn.Send("ZADD", db.Notification+":created", n.Created, id)
	_ = conn.Send("ZADD", db.Notification+":modified", n.Modified, id) //sorted set based on age
	for _, label := range n.Labels {
		_ = conn.Send("ZADD", db.Notification+":label:"+label, 0, id)
	}
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

func deleteNotification(conn redis.Conn, id string) error {
	var n contract.Notification
	err := getObjectById(conn, id, unmarshalObject, &n)
	if err != nil {
		return err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("DEL", id)
	_ = conn.Send("ZREM", db.Notification, id)
	_ = conn.Send("HDEL", db.Notification+":slug", n.Slug)
	_ = conn.Send("ZREM", db.Notification+":sender:"+n.Sender, id)
	_ = conn.Send("ZREM", db.Notification+":status:"+n.Status, id)
	_ = conn.Send("ZREM", db.Notification+":severity:"+n.Severity, id)
	_ = conn.Send("ZREM", db.Notification+":created", id)
	_ = conn.Send("ZREM", db.Notification+":modified", id)
	for _, label := range n.Labels {
		_ = conn.Send("ZREM", db.Notification+":label:"+label, id)
	}
	_, err = conn.Do("EXEC")

	return err

}

func addSubscription(conn redis.Conn, s *contract.Subscription) error {
	exist, err := redis.Bool(conn.Do("HEXISTS", db.Subscription+":slug", s.Slug))
	if err != nil {
		return err
	} else if exist {
		return errors.Errorf("%v, slug=%v", db.ErrNotUnique, s.Slug)
	}

	if s.Created == 0 {
		s.Created = db.MakeTimestamp()
		s.Modified = s.Created
	}

	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	m, err := marshalObject(s)
	if err != nil {
		return err
	}

	id := s.ID

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", id, m)
	_ = conn.Send("ZADD", db.Subscription, 0, id)
	_ = conn.Send("HSET", db.Subscription+":slug", s.Slug, id)
	_ = conn.Send("ZADD", db.Subscription+":receiver:"+s.Receiver, 0, id)
	for _, label := range s.SubscribedLabels {
		_ = conn.Send("SADD", db.Subscription+":label:"+label, id)
	}
	for _, category := range s.SubscribedCategories {
		_ = conn.Send("SADD", db.Subscription+":category:"+category, id)
	}
	_, err = conn.Do("EXEC")

	return err
}

func deleteSubscription(conn redis.Conn, id string) error {
	var s contract.Subscription
	err := getObjectById(conn, id, unmarshalObject, &s)
	if err != nil {
		return err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("DEL", id)
	_ = conn.Send("ZREM", db.Subscription, id)
	_ = conn.Send("HDEL", db.Subscription+":slug", s.Slug)
	_ = conn.Send("ZREM", db.Subscription+":receiver:"+s.Receiver, id)
	for _, label := range s.SubscribedLabels {
		_ = conn.Send("SREM", db.Subscription+":label:"+label, id)
	}
	for _, category := range s.SubscribedCategories {
		_ = conn.Send("SREM", db.Subscription+":category:"+category, id)
	}
	_, err = conn.Do("EXEC")

	return err
}

func addTransmission(conn redis.Conn, t *contract.Transmission) error {
	if t.Created == 0 {
		t.Created = db.MakeTimestamp()
		t.Modified = t.Created
	}

	if t.ID == "" {
		t.ID = uuid.New().String()
	}

	m, err := marshalObject(t)
	if err != nil {
		return err
	}
	id := t.ID

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", id, m)
	_ = conn.Send("ZADD", db.Transmission, 0, id)
	_ = conn.Send("ZADD", db.Transmission+":slug:"+t.Notification.Slug, t.ResendCount, id)
	_ = conn.Send("ZADD", db.Transmission+":status:"+t.Status, t.ResendCount, id)
	_ = conn.Send("ZADD", db.Transmission+":resendcount", t.ResendCount, id)
	_ = conn.Send("ZADD", db.Transmission+":created", t.Created, id)
	_ = conn.Send("ZADD", db.Transmission+":modified", t.Modified, id)
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

func deleteTransmission(conn redis.Conn, id string) error {
	var t contract.Transmission
	err := getObjectById(conn, id, unmarshalObject, &t)
	if err != nil {
		return err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("DEL", id)
	_ = conn.Send("ZREM", db.Transmission, id)
	_ = conn.Send("ZREM", db.Transmission+":slug"+t.Notification.Slug, t.Notification.Slug)
	_ = conn.Send("ZREM", db.Transmission+":status:"+t.Status, id)
	_ = conn.Send("ZREM", db.Transmission+":resendcount", id)
	_ = conn.Send("ZREM", db.Transmission+":created", id)
	_ = conn.Send("ZREM", db.Transmission+":modified", id)
	_, err = conn.Do("EXEC")

	return err
}

func (c Client) deleteTransmissionBySlug(conn redis.Conn, slug string) error {
	transmissions, err := c.GetTransmissionsByNotificationSlug(slug, -1)
	if err != nil {
		return err
	}
	for _, transmission := range transmissions {
		err = deleteTransmission(conn, transmission.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func unmarshalNotifications(objects [][]byte) ([]contract.Notification, error) {
	var unmarshalObjects []contract.Notification
	for _, o := range objects {
		if len(o) > 0 {
			var m contract.Notification
			err := unmarshalObject(o, &m)
			if err != nil {
				return unmarshalObjects, err
			}
			unmarshalObjects = append(unmarshalObjects, m)
		}
	}
	return unmarshalObjects, nil
}

func unmarshalSubscriptions(objects [][]byte) ([]contract.Subscription, error) {
	var unmarshalObjects []contract.Subscription
	for _, o := range objects {
		if len(o) > 0 {
			var m contract.Subscription
			err := unmarshalObject(o, &m)
			if err != nil {
				return unmarshalObjects, err
			}
			unmarshalObjects = append(unmarshalObjects, m)
		}
	}
	return unmarshalObjects, nil
}

func unmarshalTransmissions(objects [][]byte) ([]contract.Transmission, error) {
	var unmarshalObjects []contract.Transmission
	for _, o := range objects {
		if len(o) > 0 {
			var m contract.Transmission
			err := unmarshalObject(o, &m)
			if err != nil {
				return unmarshalObjects, err
			}
			unmarshalObjects = append(unmarshalObjects, m)
		}
	}
	return unmarshalObjects, nil
}

func notificationBySlug(conn redis.Conn, slug string) (notification contract.Notification, err error) {
	id, err := redis.String(conn.Do("HGET", db.Notification+":slug", slug))
	if err != nil {
		if err == redis.ErrNil {
			return notification, db.ErrNotFound
		}
		return notification, err
	}

	err = getObjectById(conn, id, unmarshalObject, &notification)
	if err != nil {
		return notification, err
	}

	return notification, nil
}

func subscriptionBySlug(conn redis.Conn, slug string) (subscription contract.Subscription, err error) {
	id, err := redis.String(conn.Do("HGET", db.Subscription+":slug", slug))
	if err != nil {
		if err == redis.ErrNil {
			return subscription, db.ErrNotFound
		}
		return subscription, err
	}

	err = getObjectById(conn, id, unmarshalObject, &subscription)
	if err != nil {
		return subscription, err
	}

	return subscription, nil
}

func getTransmissionsByStartEndWithResendLimit(conn redis.Conn, start, end int64, resendLimit int) (objects [][]byte, err error) {
	args := []interface{}{db.Transmission + ":created"}
	if start < 0 {
		args = append(args, "-inf")
	} else {
		args = append(args, start)
	}
	if end < 0 {
		args = append(args, "+inf")
	} else {
		args = append(args, end)
	}

	ids, err := redis.Values(conn.Do("ZRANGEBYSCORE", args...))
	if err != nil && err != redis.ErrNil {
		return nil, err
	}

	var filteredIDs []interface{}
	for _, id := range ids {
		score, err := redis.Int(conn.Do("ZSCORE", db.Transmission+":resendcount", id))
		if err != nil {
			return nil, err
		}
		if score <= resendLimit {
			filteredIDs = append(filteredIDs, id)
		}
	}

	if len(filteredIDs) > 0 {
		objects, err = redis.ByteSlices(conn.Do("MGET", filteredIDs...))
		if err != nil {
			return nil, err
		}
	}
	return objects, nil
}
