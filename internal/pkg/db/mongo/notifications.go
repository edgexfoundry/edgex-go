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
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

const (
	NOTIFICATION_COLLECTION = "notification"
	SUBSCRIPTION_COLLECTION = "subscription"
	TRANSMISSION_COLLECTION = "transmission"
)

var currentReadMaxLimit int // configuration read max limit
var currentResendLimit int  // configuration transmission resent count limit
var cleanupDefaultAge int

// ******************************* NOTIFICATIONS **********************************

func (mc *MongoClient) Notifications() ([]models.Notification, error) {
	return mc.getNotifications(bson.M{})
}

func (mc *MongoClient) NotificationById(id string) (models.Notification, error) {
	if !bson.IsObjectIdHex(id) {
		return models.Notification{}, db.ErrInvalidObjectId
	}
	return mc.getNotificaiton(bson.M{"_id": bson.ObjectIdHex(id)})
}

func (mc *MongoClient) NotificationBySlug(slug string) (models.Notification, error) {
	return mc.getNotificaiton(bson.M{"slug": slug})
}

func (mc *MongoClient) NotificationBySender(sender string, limit int) ([]models.Notification, error) {
	return mc.getNotificationsLimit(bson.M{"sender": sender}, limit)
}

func (mc *MongoClient) NotificationsByLabels(labels []string, limit int) ([]models.Notification, error) {
	return mc.getNotificationsLimit(bson.M{"labels": bson.M{"$in": labels}}, limit)
}

func (mc *MongoClient) NotificationsByStartEnd(start int64, end int64, limit int) ([]models.Notification, error) {
	query := bson.M{"created": bson.M{"$gt": start, "$lt": end}}
	return mc.getNotificationsLimit(query, limit)
}

func (mc *MongoClient) NotificationsByStart(start int64, limit int) ([]models.Notification, error) {
	query := bson.M{"created": bson.M{"$gt": start}}
	return mc.getNotificationsLimit(query, limit)
}

func (mc *MongoClient) NotificationsByEnd(end int64, limit int) ([]models.Notification, error) {
	query := bson.M{"created": bson.M{"$lt": end}}
	return mc.getNotificationsLimit(query, limit)
}

func (mc *MongoClient) NotificationsNew(limit int) ([]models.Notification, error) {
	return mc.getNotificationsLimit(bson.M{"status": "NEW"}, limit)
}

func (mc *MongoClient) NotificationsNewNormal(limit int) ([]models.Notification, error) {
	return mc.getNotificationsLimit(bson.M{"status": "NEW", "severity": "NORMAL"}, limit)
}

func (mc *MongoClient) AddNotification(n *models.Notification) (bson.ObjectId, error) {
	return mc.addNotification(n)
}

func (mc *MongoClient) UpdateNotification(n models.Notification) error {
	return mc.updateNotification(n)
}

func (mc *MongoClient) MarkNotificationProcessed(n models.Notification) error {
	n.Status = models.NotificationsStatus(models.Processed)
	return mc.updateNotification(n)
}

func (mc *MongoClient) DeleteNotificationById(id string) error {
	mn, err := mc.NotificationById(id)
	if err == db.ErrNotFound {
		return db.ErrNotFound
	}
	return mc.deleteNotificationAndAssociatedTransmissions(mn)
}

func (mc *MongoClient) DeleteNotificationBySlug(slug string) error {
	mn, err := mc.NotificationBySlug(slug)
	if err == db.ErrNotFound {
		return db.ErrNotFound
	}
	return mc.deleteNotificationAndAssociatedTransmissions(mn)
}

func (mc *MongoClient) DeleteNotificationsOld(age int) error {
	currentTime := db.MakeTimestamp()
	end := int(currentTime) - age
	query := bson.M{"modified": bson.M{
		"$lt": end}, "status": "PROCESSED"}
	mns, err := mc.getNotifications(query)
	if err == db.ErrNotFound {
		return db.ErrNotFound
	}
	for _, mn := range mns {
		err = mc.deleteNotificationAndAssociatedTransmissions(mn)
		if err != nil {
			return err
		}
	}
	return err
}

// ******************************* SUBSCRIPTIONS **********************************

func (mc *MongoClient) SubscriptionBySlug(slug string) (models.Subscription, error) {
	return mc.getSubscription(bson.M{"slug": slug})
}

func (mc *MongoClient) SubscriptionByCategories(categories []string) ([]models.Subscription, error) {
	return mc.getSubscriptions(bson.M{"subscribedcategories": bson.M{"$in": categories}})
}

func (mc *MongoClient) SubscriptionByLabels(labels []string) ([]models.Subscription, error) {
	return mc.getSubscriptions(bson.M{"subscribedlabels": bson.M{"$in": labels}})
}

func (mc *MongoClient) SubscriptionByCategoriesLabels(categories []string, labels []string) ([]models.Subscription, error) {
	return mc.getSubscriptions(bson.M{"subscribedcategories": bson.M{"$in": categories}, "subscribedlabels": bson.M{"$in": labels}})
}

func (mc *MongoClient) SubscriptionByReceiver(receiver string) ([]models.Subscription, error) {
	return mc.getSubscriptions(bson.M{"receiver": receiver})
}

func (mc *MongoClient) SubscriptionById(id string) (models.Subscription, error) {
	if !bson.IsObjectIdHex(id) {
		return models.Subscription{}, db.ErrInvalidObjectId
	}
	return mc.getSubscription(bson.M{"_id": bson.ObjectIdHex(id)})
}

func (mc *MongoClient) AddSubscription(s *models.Subscription) (bson.ObjectId, error) {
	return mc.addSubscription(s)
}

func (mc *MongoClient) UpdateSubscription(s models.Subscription) error {
	return mc.updateSubscription(s)
}

func (mc *MongoClient) DeleteSubscriptionBySlug(slug string) error {
	ms, err := mc.SubscriptionBySlug(slug)
	if err == db.ErrNotFound {
		return db.ErrNotFound
	}
	return mc.deleteByObjectID(ms.ID, SUBSCRIPTION_COLLECTION)
}

// Return all the subscriptions
// UnexpectedError - failed to retrieve subscriptions from the database
func (mc *MongoClient) Subscriptions() ([]models.Subscription, error) {
	return mc.getSubscriptions(bson.M{})
}

// ******************************* TRANSMISSIONS  **********************************
// limits for transmissions here refer to resend counts
func (mc *MongoClient) AddTransmission(t *models.Transmission) (bson.ObjectId, error) {
	return mc.addTransmission(t)
}

func (mc *MongoClient) UpdateTransmission(t models.Transmission) error {
	return mc.updateTransmission(t)
}

func (mc *MongoClient) DeleteTransmission(age int64, status models.TransmissionStatus) error {
	currentTime := db.MakeTimestamp()
	end := currentTime - age
	query := bson.M{"modified": bson.M{"$lt": end}, "status": status}
	return mc.deleteAll(query, TRANSMISSION_COLLECTION)
}

func (mc *MongoClient) TransmissionsByNotificationSlug(slug string, resendLimit int) ([]models.Transmission, error) {
	return mc.getTransmissionsLimit(bson.M{"resendcount": bson.M{"$lt": resendLimit}, "notification.slug": slug})
}

func (mc *MongoClient) TransmissionsByStartEnd(start int64, end int64, resendLimit int) ([]models.Transmission, error) {
	return mc.getTransmissionsLimit(bson.M{"resendcount": bson.M{"$lt": resendLimit}, "created": bson.M{"$gt": start, "$lt": end}})
}

func (mc *MongoClient) TransmissionsByStart(start int64, resendLimit int) ([]models.Transmission, error) {
	return mc.getTransmissionsLimit(bson.M{"resendcount": bson.M{"$lt": resendLimit}, "created": bson.M{"$gt": start}})
}

func (mc *MongoClient) TransmissionsByEnd(end int64, resendLimit int) ([]models.Transmission, error) {
	return mc.getTransmissionsLimit(bson.M{"resendcount": bson.M{"$lt": resendLimit}, "created": bson.M{"$lt": end}})
}

func (mc *MongoClient) TransmissionsByStatus(resendLimit int, status models.TransmissionStatus) ([]models.Transmission, error) {
	return mc.getTransmissionsLimit(bson.M{"resendcount": bson.M{"$lt": resendLimit}, "status": status})
}

// ******************************* GENERAL CLEANUP **********************************

func (mc *MongoClient) Cleanup() error {
	return mc.CleanupOld(cleanupDefaultAge)
}

func (mc *MongoClient) CleanupOld(age int) error {
	currentTime := db.MakeTimestamp()
	end := int(currentTime) - age
	query := bson.M{"modified": bson.M{"$lt": end}}
	mns, err := mc.getNotifications(query)
	if err == db.ErrNotFound {
		return db.ErrNotFound
	}
	for _, mn := range mns {
		err = mc.deleteNotificationAndAssociatedTransmissions(mn)
		if err != nil {
			return err
		}
	}
	return err
}

// ******************************* NOTIFICATIONS **********************************

func (mc *MongoClient) deleteNotificationAndAssociatedTransmissions(n models.Notification) error {

	err := mc.deleteAll(bson.M{"notification.slug": n.Slug}, TRANSMISSION_COLLECTION)

	if err != nil {
		return err
	}
	return mc.deleteByObjectID(n.ID, NOTIFICATION_COLLECTION)

}

func (mc *MongoClient) addNotification(n *models.Notification) (bson.ObjectId, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	n.Created = db.MakeTimestamp()
	n.ID = bson.NewObjectId()

	// Handle DBRefs
	mn := MongoNotification{Notification: *n}

	err := mc.checkNotificationSlugIntegrity(mn.Slug)
	if err != nil {
		return n.ID, err
	}

	err = s.DB(mc.database.Name).C(NOTIFICATION_COLLECTION).Insert(mn)
	if err != nil {
		return n.ID, err
	}

	return n.ID, err
}

func (mc *MongoClient) checkNotificationSlugIntegrity(slug string) error {
	if slug == "" {
		return db.ErrSlugEmpty
	}
	_, err := mc.getNotificaiton(bson.M{"slug": slug})
	if err == nil {
		return db.ErrNotUnique
	}
	return nil
}

func (mc *MongoClient) updateNotification(n models.Notification) error {
	s := mc.getSessionCopy()
	defer s.Close()

	n.Modified = db.MakeTimestamp()

	// Handle DBRef
	mn := MongoNotification{Notification: n}

	err := s.DB(mc.database.Name).C(NOTIFICATION_COLLECTION).UpdateId(mn.ID, mn)
	if err == db.ErrNotFound {
		return db.ErrNotFound
	}

	return err
}

func (mc *MongoClient) getNotificaiton(q bson.M) (models.Notification, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Handle DBRef
	var mn MongoNotification
	err := s.DB(mc.database.Name).C(NOTIFICATION_COLLECTION).Find(q).One(&mn)
	err = errorMap(err)
	if err == db.ErrNotFound {
		return mn.Notification, db.ErrNotFound
	}

	return mn.Notification, err
}

func (mc *MongoClient) getNotifications(q bson.M) ([]models.Notification, error) {
	return mc.getNotificationsLimit(q, currentReadMaxLimit)
}

func (mc *MongoClient) getNotificationsLimit(q bson.M, limit int) ([]models.Notification, error) {

	s := mc.getSessionCopy()
	defer s.Close()

	// Handle DBRefs
	var mn []MongoNotification
	notifications := []models.Notification{}
	// Check if limit is 0
	if limit == 0 {
		return notifications, nil
	}
	err := s.DB(mc.database.Name).C(NOTIFICATION_COLLECTION).Find(q).Limit(limit).All(&mn)
	if err != nil {
		return notifications, err
	}

	for _, n := range mn {
		notifications = append(notifications, n.Notification)
	}

	return notifications, nil
}

// ******************************* SUBSCRIPTIONS **********************************

func (mc *MongoClient) addSubscription(sub *models.Subscription) (bson.ObjectId, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	sub.Created = db.MakeTimestamp()
	sub.ID = bson.NewObjectId()

	// Handle DBRefs
	ms := MongoSubscription{Subscription: *sub}

	err := mc.checkSubscriptionSlugIntegrity(ms.Slug)
	if err != nil {
		return sub.ID, err
	}

	err = s.DB(mc.database.Name).C(SUBSCRIPTION_COLLECTION).Insert(ms)
	if err != nil {
		return sub.ID, err
	}

	return sub.ID, err
}

func (mc *MongoClient) checkSubscriptionSlugIntegrity(slug string) error {
	if slug == "" {
		return db.ErrSlugEmpty
	}
	_, err := mc.getSubscription(bson.M{"slug": slug})
	if err == nil {
		return db.ErrNotUnique
	}
	return nil
}

func (mc *MongoClient) updateSubscription(sub models.Subscription) error {
	s := mc.getSessionCopy()
	defer s.Close()

	sub.Modified = db.MakeTimestamp()

	// Handle DBRef
	ms := MongoSubscription{Subscription: sub}

	err := s.DB(mc.database.Name).C(SUBSCRIPTION_COLLECTION).UpdateId(ms.ID, ms)
	if err == db.ErrNotFound {
		return db.ErrNotFound
	}

	return err
}

func (mc *MongoClient) getSubscription(q bson.M) (models.Subscription, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Handle DBRef
	var ms MongoSubscription
	err := s.DB(mc.database.Name).C(SUBSCRIPTION_COLLECTION).Find(q).One(&ms)
	err = errorMap(err)
	if err == db.ErrNotFound {
		return ms.Subscription, db.ErrNotFound
	}

	return ms.Subscription, err
}

func (mc *MongoClient) getSubscriptions(q bson.M) ([]models.Subscription, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Handle DBRefs
	var ms []MongoSubscription
	subscriptions := []models.Subscription{}
	err := s.DB(mc.database.Name).C(SUBSCRIPTION_COLLECTION).Find(q).All(&ms)
	if err != nil {
		return subscriptions, err
	}

	// Append all the subscriptions
	for _, subs := range ms {
		subscriptions = append(subscriptions, subs.Subscription)
	}

	return subscriptions, nil
}

func (mc *MongoClient) getSubscriptionsLimit(q bson.M, limit int) ([]models.Subscription, error) {

	s := mc.getSessionCopy()
	defer s.Close()

	// Handle DBRefs
	var ms []MongoSubscription
	subscriptions := []models.Subscription{}
	// Check if limit is 0
	if limit == 0 {
		return subscriptions, nil
	}
	err := s.DB(mc.database.Name).C(SUBSCRIPTION_COLLECTION).Find(q).Limit(limit).All(&ms)
	if err != nil {
		return subscriptions, err
	}

	for _, sub := range ms {
		subscriptions = append(subscriptions, sub.Subscription)
	}

	return subscriptions, nil
}

// ******************************* TRANSMISSIONS **********************************

func (mc *MongoClient) addTransmission(tran *models.Transmission) (bson.ObjectId, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	tran.Created = db.MakeTimestamp()
	tran.ID = bson.NewObjectId()

	// Handle DBRefs
	mt := MongoTransmission{Transmission: *tran}

	err := s.DB(mc.database.Name).C(TRANSMISSION_COLLECTION).Insert(mt)
	if err != nil {
		return tran.ID, err
	}

	return tran.ID, err
}

func (mc *MongoClient) updateTransmission(tran models.Transmission) error {
	s := mc.getSessionCopy()
	defer s.Close()

	tran.Modified = db.MakeTimestamp()

	// Handle DBRef
	mt := MongoTransmission{Transmission: tran}

	err := s.DB(mc.database.Name).C(TRANSMISSION_COLLECTION).UpdateId(mt.ID, mt)
	if err == db.ErrNotFound {
		return db.ErrNotFound
	}

	return err
}

func (mc *MongoClient) getTransmission(q bson.M) (models.Transmission, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Handle DBRef
	var mt MongoTransmission
	err := s.DB(mc.database.Name).C(TRANSMISSION_COLLECTION).Find(q).One(&mt)
	if err == db.ErrNotFound {
		return mt.Transmission, db.ErrNotFound
	}

	return mt.Transmission, err
}

func (mc *MongoClient) getTransmissionsLimit(q bson.M) ([]models.Transmission, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Handle DBRefs
	var mt []MongoTransmission
	trans := []models.Transmission{}
	err := s.DB(mc.database.Name).C(TRANSMISSION_COLLECTION).Find(q).All(&mt)
	//err := s.DB(mc.database.Name).C(NOTIFICATION_COLLECTION).Find(q).Limit(limit).All(&mt)
	if err != nil {
		return trans, err
	}

	for _, t := range mt {
		trans = append(trans, t.Transmission)
	}

	return trans, nil
}

/////////////////////////////////////// General delete functions ////////////////////////////////////////////

func (mc *MongoClient) deleteByID(id string, col string) error {

	if !bson.IsObjectIdHex(id) {
		return db.ErrInvalidObjectId
	}

	return mc.deleteByObjectID(bson.ObjectIdHex(id), col)

}

func (mc *MongoClient) deleteByObjectID(id bson.ObjectId, col string) error {
	s := mc.getSessionCopy()
	defer s.Close()

	err := s.DB(mc.database.Name).C(col).RemoveId(id)
	if err == db.ErrNotFound {
		return db.ErrNotFound
	}
	return err
}

func (mc *MongoClient) deleteAll(q bson.M, col string) error {
	s := mc.getSessionCopy()
	defer s.Close()

	_, err := s.DB(mc.database.Name).C(col).RemoveAll(q)
	if err == db.ErrNotFound {
		return db.ErrNotFound
	}
	return err
}
