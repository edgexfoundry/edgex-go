/*******************************************************************************
 * Copyright 2018 Dell Inc.
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
 *
 *******************************************************************************/

package mongo

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo/models"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
)

const (
	NOTIFICATION_COLLECTION = "notification"
	SUBSCRIPTION_COLLECTION = "subscription"
	TRANSMISSION_COLLECTION = "transmission"
)

var currentReadMaxLimit int // configuration read max limit
var currentResendLimit int  // configuration transmission resent count limit
var cleanupDefaultAge int

/* ----------------------- Internal Common Utility Functions ------------------------*/

func idToQueryParameters(id string) (name string, value interface{}, err error) {
	if !bson.IsObjectIdHex(id) {
		_, err := uuid.Parse(id)
		if err != nil { // It is some unsupported type of string
			return "", "", db.ErrInvalidObjectId
		}
		name = "uuid"
		value = id
	} else {
		name = "_id"
		value = bson.ObjectIdHex(id)
	}
	return
}

func idToBsonM(id string) (q bson.M, err error) {
	var name string
	var value interface{}
	name, value, err = idToQueryParameters(id)
	if err != nil {
		return
	}
	q = bson.M{name: value}
	return
}

/* ----------------------- Notifications ------------------------*/

func (mc MongoClient) GetNotifications() ([]contract.Notification, error) {
	return mc.getNotifications(bson.M{})
}

func (mc MongoClient) GetNotificationById(id string) (contract.Notification, error) {
	var query bson.M
	var err error
	if query, err = idToBsonM(id); err != nil {
		return contract.Notification{}, err
	}
	return mc.getNotification(query)
}

func (mc MongoClient) GetNotificationBySlug(slug string) (contract.Notification, error) {
	return mc.getNotification(bson.M{"slug": slug})
}

func (mc MongoClient) GetNotificationBySender(sender string, limit int) ([]contract.Notification, error) {
	return mc.getNotificationsLimit(bson.M{"sender": sender}, limit)
}

func (mc MongoClient) GetNotificationsByLabels(labels []string, limit int) ([]contract.Notification, error) {
	return mc.getNotificationsLimit(bson.M{"labels": bson.M{"$in": labels}}, limit)
}

func (mc MongoClient) GetNotificationsByStartEnd(start int64, end int64, limit int) ([]contract.Notification, error) {
	query := bson.M{"created": bson.M{"$gt": start, "$lt": end}}
	return mc.getNotificationsLimit(query, limit)
}

func (mc MongoClient) GetNotificationsByStart(start int64, limit int) ([]contract.Notification, error) {
	query := bson.M{"created": bson.M{"$gt": start}}
	return mc.getNotificationsLimit(query, limit)
}

func (mc MongoClient) GetNotificationsByEnd(end int64, limit int) ([]contract.Notification, error) {
	query := bson.M{"created": bson.M{"$lt": end}}
	return mc.getNotificationsLimit(query, limit)
}

func (mc MongoClient) GetNewNotifications(limit int) ([]contract.Notification, error) {
	return mc.getNotificationsLimit(bson.M{"status": "NEW"}, limit)
}

func (mc MongoClient) GetNewNormalNotifications(limit int) ([]contract.Notification, error) {
	return mc.getNotificationsLimit(bson.M{"status": "NEW", "severity": "NORMAL"}, limit)
}

func (mc MongoClient) AddNotification(n contract.Notification) (string, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var err error
	var model models.Notification
	if err = model.FromContract(n); err != nil {
		return "", err
	}

	err = mc.checkNotificationSlugIntegrity(model.Slug)
	if err != nil {
		return "", err
	}

	err = s.DB(mc.database.Name).C(NOTIFICATION_COLLECTION).Insert(model)
	if err != nil {
		return "", err
	}

	return model.Uuid, err
}

func (mc MongoClient) UpdateNotification(n contract.Notification) error {
	return mc.updateNotification(n)
}

func (mc MongoClient) MarkNotificationProcessed(n contract.Notification) error {
	n.Status = contract.NotificationsStatus(contract.Processed)
	return mc.updateNotification(n)
}

func (mc MongoClient) DeleteNotificationById(id string) error {
	mn, err := mc.GetNotificationById(id)
	if err == db.ErrNotFound {
		return err
	}
	return mc.deleteNotificationAndAssociatedTransmissions(mn)
}

func (mc MongoClient) DeleteNotificationBySlug(slug string) error {
	mn, err := mc.GetNotificationBySlug(slug)
	if err == db.ErrNotFound {
		return err
	}
	return mc.deleteNotificationAndAssociatedTransmissions(mn)
}

func (mc MongoClient) DeleteNotificationsOld(age int) error {
	currentTime := db.MakeTimestamp()
	end := int(currentTime) - age
	query := bson.M{"modified": bson.M{
		"$lt": end}, "status": "PROCESSED"}
	mns, err := mc.getNotifications(query)
	if err == db.ErrNotFound {
		return err
	}
	for _, mn := range mns {
		err = mc.deleteNotificationAndAssociatedTransmissions(mn)
		if err != nil {
			return err
		}
	}
	return err
}

func (mc MongoClient) deleteNotificationAndAssociatedTransmissions(n contract.Notification) error {
	if err := mc.deleteAll(bson.M{"notification.slug": n.Slug}, TRANSMISSION_COLLECTION); err != nil {
		return err
	}

	return mc.deleteByObjectID(n.ID, NOTIFICATION_COLLECTION)
}

func (mc MongoClient) checkNotificationSlugIntegrity(slug string) error {
	if slug == "" {
		return db.ErrSlugEmpty
	}
	_, err := mc.getNotification(bson.M{"slug": slug})
	if err == nil {
		return db.ErrNotUnique
	}
	return nil
}

func (mc MongoClient) updateNotification(n contract.Notification) error {
	s := mc.getSessionCopy()
	defer s.Close()

	var model models.Notification
	if err := model.FromContract(n); err != nil {
		return err
	}

	model.Modified = db.MakeTimestamp()

	col := s.DB(mc.database.Name).C(NOTIFICATION_COLLECTION)
	if model.Id.Valid() {
		return col.UpdateId(model.Id, model)
	}
	return col.Update(bson.M{"uuid": model.Uuid}, model)
}

func (mc MongoClient) getNotification(q bson.M) (contract.Notification, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var model models.Notification
	err := errorMap(s.DB(mc.database.Name).C(NOTIFICATION_COLLECTION).Find(q).One(&model))
	if err == db.ErrNotFound {
		return contract.Notification{}, err
	}

	return model.ToContract(), nil
}

func (mc MongoClient) getNotifications(q bson.M) ([]contract.Notification, error) {
	return mc.getNotificationsLimit(q, currentReadMaxLimit)
}

func (mc MongoClient) getNotificationsLimit(q bson.M, limit int) ([]contract.Notification, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Check if limit is 0
	if limit == 0 {
		return []contract.Notification{}, nil
	}

	var notifications []models.Notification
	if err := s.DB(mc.database.Name).C(NOTIFICATION_COLLECTION).Find(q).Limit(limit).All(&notifications); err != nil {
		return []contract.Notification{}, err
	}

	var contracts []contract.Notification
	for _, model := range notifications {
		contracts = append(contracts, model.ToContract())
	}
	return contracts, nil
}

/* ----------------------- Subscriptions ------------------------*/

func (mc MongoClient) GetSubscriptionBySlug(slug string) (contract.Subscription, error) {
	return mc.getSubscription(bson.M{"slug": slug})
}

func (mc MongoClient) GetSubscriptionByCategories(categories []string) ([]contract.Subscription, error) {
	return mc.getSubscriptions(bson.M{"subscribedCategories": bson.M{"$in": categories}})
}

func (mc MongoClient) GetSubscriptionByLabels(labels []string) ([]contract.Subscription, error) {
	return mc.getSubscriptions(bson.M{"subscribedLabels": bson.M{"$in": labels}})
}

func (mc MongoClient) GetSubscriptionByCategoriesLabels(categories []string, labels []string) ([]contract.Subscription, error) {
	return mc.getSubscriptions(bson.M{"subscribedCategories": bson.M{"$in": categories}, "subscribedLabels": bson.M{"$in": labels}})
}

func (mc MongoClient) GetSubscriptionByReceiver(receiver string) ([]contract.Subscription, error) {
	return mc.getSubscriptions(bson.M{"receiver": receiver})
}

func (mc MongoClient) GetSubscriptionById(id string) (contract.Subscription, error) {
	var query bson.M
	var err error
	if query, err = idToBsonM(id); err != nil {
		return contract.Subscription{}, err
	}
	return mc.getSubscription(query)
}

func (mc MongoClient) AddSubscription(sub contract.Subscription) (string, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var err error
	var model models.Subscription
	if err = model.FromContract(sub); err != nil {
		return "", err
	}

	if err = mc.checkSubscriptionSlugIntegrity(model.Slug); err != nil {
		return "", err
	}

	if err = s.DB(mc.database.Name).C(SUBSCRIPTION_COLLECTION).Insert(model); err != nil {
		return "", err
	}

	return model.Uuid, err
}

func (mc MongoClient) UpdateSubscription(sub contract.Subscription) error {
	s := mc.getSessionCopy()
	defer s.Close()

	var model models.Subscription
	if err := model.FromContract(sub); err != nil {
		return err
	}

	model.Modified = db.MakeTimestamp()

	col := s.DB(mc.database.Name).C(SUBSCRIPTION_COLLECTION)
	if model.Id.Valid() {
		return col.UpdateId(model.Id, model)
	}
	return col.Update(bson.M{"uuid": model.Uuid}, model)
}

func (mc MongoClient) DeleteSubscriptionBySlug(slug string) error {
	ms, err := mc.GetSubscriptionBySlug(slug)
	if err == db.ErrNotFound {
		return err
	}
	return mc.deleteByObjectID(ms.ID, SUBSCRIPTION_COLLECTION)
}

// Return all the subscriptions
// UnexpectedError - failed to retrieve subscriptions from the database
func (mc MongoClient) GetSubscriptions() ([]contract.Subscription, error) {
	return mc.getSubscriptions(bson.M{})
}

func (mc MongoClient) checkSubscriptionSlugIntegrity(slug string) error {
	if slug == "" {
		return db.ErrSlugEmpty
	}
	_, err := mc.getSubscription(bson.M{"slug": slug})
	if err == nil {
		return db.ErrNotUnique
	}
	return nil
}

func (mc MongoClient) getSubscription(q bson.M) (c contract.Subscription, err error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var sub models.Subscription
	if err = s.DB(mc.database.Name).C(SUBSCRIPTION_COLLECTION).Find(q).One(&sub); err != nil {
		return c, errorMap(err)
	}
	c = sub.ToContract()
	return
}

func (mc MongoClient) getSubscriptions(q bson.M) (c []contract.Subscription, err error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var subs []models.Subscription
	if err = s.DB(mc.database.Name).C(SUBSCRIPTION_COLLECTION).Find(q).All(&subs); err != nil {
		return
	}

	for _, sub := range subs {
		c = append(c, sub.ToContract())
	}
	return
}

/* ----------------------- Transmissions ------------------------*/

// limits for transmissions here refer to resend counts

func (mc MongoClient) AddTransmission(t contract.Transmission) (string, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var model models.Transmission
	if err := model.FromContract(t); err != nil {
		return "", err
	}

	if err := s.DB(mc.database.Name).C(TRANSMISSION_COLLECTION).Insert(model); err != nil {
		return "", err
	}

	return model.Uuid, nil
}

func (mc MongoClient) UpdateTransmission(t contract.Transmission) error {
	s := mc.getSessionCopy()
	defer s.Close()

	var model models.Transmission
	if err := model.FromContract(t); err != nil {
		return err
	}

	model.Modified = db.MakeTimestamp()

	col := s.DB(mc.database.Name).C(TRANSMISSION_COLLECTION)
	if model.Id.Valid() {
		return col.UpdateId(model.Id, model)
	}
	return col.Update(bson.M{"uuid": model.Uuid}, model)
}

func (mc MongoClient) DeleteTransmission(age int64, status contract.TransmissionStatus) error {
	currentTime := db.MakeTimestamp()
	end := currentTime - age
	query := bson.M{"modified": bson.M{"$lt": end}, "status": status}
	return mc.deleteAll(query, TRANSMISSION_COLLECTION)
}

func (mc MongoClient) GetTransmissionsByNotificationSlug(slug string, resendLimit int) ([]contract.Transmission, error) {
	return mc.getTransmissionsLimit(bson.M{"resendcount": bson.M{"$lt": resendLimit}, "notification.slug": slug})
}

func (mc MongoClient) GetTransmissionsByStartEnd(start int64, end int64, resendLimit int) ([]contract.Transmission, error) {
	return mc.getTransmissionsLimit(bson.M{"resendcount": bson.M{"$lt": resendLimit}, "created": bson.M{"$gt": start, "$lt": end}})
}

func (mc MongoClient) GetTransmissionsByStart(start int64, resendLimit int) ([]contract.Transmission, error) {
	return mc.getTransmissionsLimit(bson.M{"resendcount": bson.M{"$lt": resendLimit}, "created": bson.M{"$gt": start}})
}

func (mc MongoClient) GetTransmissionsByEnd(end int64, resendLimit int) ([]contract.Transmission, error) {
	return mc.getTransmissionsLimit(bson.M{"resendcount": bson.M{"$lt": resendLimit}, "created": bson.M{"$lt": end}})
}

func (mc MongoClient) GetTransmissionsByStatus(resendLimit int, status contract.TransmissionStatus) ([]contract.Transmission, error) {
	return mc.getTransmissionsLimit(bson.M{"resendcount": bson.M{"$lt": resendLimit}, "status": status})
}

func (mc MongoClient) getTransmission(q bson.M) (c contract.Transmission, err error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var t models.Transmission
	if err = s.DB(mc.database.Name).C(TRANSMISSION_COLLECTION).Find(q).One(&t); err != nil {
		return
	}
	c = t.ToContract()
	return
}

func (mc MongoClient) getTransmissionsLimit(q bson.M) (c []contract.Transmission, err error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var transmissions []models.Transmission
	if err = s.DB(mc.database.Name).C(TRANSMISSION_COLLECTION).Find(q).All(&transmissions); err != nil {
		return
	}

	c = []contract.Transmission{}
	for _, transmission := range transmissions {
		c = append(c, transmission.ToContract())
	}
	return
}

/* ----------------------- General Deletion ------------------------*/

func (mc MongoClient) deleteByObjectID(id string, col string) (err error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var q bson.M
	if q, err = idToBsonM(id); err != nil {
		return
	}
	return s.DB(mc.database.Name).C(col).Remove(q)
}

func (mc MongoClient) deleteAll(q bson.M, col string) error {
	s := mc.getSessionCopy()
	defer s.Close()

	_, err := s.DB(mc.database.Name).C(col).RemoveAll(q)
	return err
}

/* ----------------------- General Cleanup ------------------------*/

func (mc MongoClient) Cleanup() error {
	return mc.CleanupOld(cleanupDefaultAge)
}

func (mc MongoClient) CleanupOld(age int) error {
	currentTime := db.MakeTimestamp()
	end := int(currentTime) - age
	query := bson.M{"modified": bson.M{"$lt": end}}
	mns, err := mc.getNotifications(query)
	if err == db.ErrNotFound {
		return err
	}
	for _, mn := range mns {
		err = mc.deleteNotificationAndAssociatedTransmissions(mn)
		if err != nil {
			return err
		}
	}
	return err
}
