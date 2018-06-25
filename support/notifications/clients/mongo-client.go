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

package clients

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/edgexfoundry/edgex-go/support/notifications/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	NOTIFICATION_COLLECTION = "notification"
	SUBSCRIPTION_COLLECTION = "subscription"
	TRANSMISSION_COLLECTION = "transmission"
)

var currentMongoClient *MongoClient // Singleton used so that MongoEvent can use it to de-reference readings
var currentReadMaxLimit int         // configuration read max limit
var currentResendLimit int          // configuration transmission resent count limit
var cleanupDefaultAge int64

type MongoClient struct {
	Session  *mgo.Session  // Mongo database session
	Database *mgo.Database // Mongo database
}

// Return a pointer to the MongoClient
func newMongoClient(config DBConfiguration) (*MongoClient, error) {
	// Create the dial info for the Mongo session
	connectionString := config.Host + ":" + strconv.Itoa(config.Port)
	fmt.Println("INFO: Connecting to mongo at: " + connectionString)
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{connectionString},
		Timeout:  time.Duration(config.Timeout) * time.Millisecond,
		Database: config.DatabaseName,
		Username: config.Username,
		Password: config.Password,
	}
	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		fmt.Println("Error dialing the mongo server: " + err.Error())
		return nil, err
	}

	mongoClient := &MongoClient{Session: session, Database: session.DB(config.DatabaseName)}
	currentMongoClient = mongoClient             // Set the singleton
	currentReadMaxLimit = config.ReadMax         // Set the read max
	currentResendLimit = config.ResendLimit      // Set the transmission resend count limit
	cleanupDefaultAge = config.CleanupDefaultAge //Set the default clean up age

	return mongoClient, nil
}

// Get the current Mongo Client
func getCurrentMongoClient() (*MongoClient, error) {
	if currentMongoClient == nil {
		return nil, errors.New("No current mongo client, please create a new client before requesting it")
	}

	return currentMongoClient, nil
}

// Get a copy of the session
func (mc *MongoClient) GetSessionCopy() *mgo.Session {
	return mc.Session.Copy()
}

// ******************************* NOTIFICATIONS **********************************

func (mc *MongoClient) Notifications() ([]models.Notification, error) {
	return mc.getNotifications(bson.M{})
}

func (mc *MongoClient) NotificationById(id string) (models.Notification, error) {
	if !bson.IsObjectIdHex(id) {
		return models.Notification{}, ErrInvalidObjectId
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
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}
	return mc.deleteNotificationAndAssociatedTransmissions(mn)
}

func (mc *MongoClient) DeleteNotificationBySlug(slug string) error {
	mn, err := mc.NotificationBySlug(slug)
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}
	return mc.deleteNotificationAndAssociatedTransmissions(mn)
}

func (mc *MongoClient) DeleteNotificationsOld(age int64) error {
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	end := currentTime - age
	query := bson.M{"modified": bson.M{
		"$lt": end}, "status": "PROCESSED"}
	mns, err := mc.getNotifications(query)
	if err == mgo.ErrNotFound {
		return ErrNotFound
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
		return models.Subscription{}, ErrInvalidObjectId
	}
	return mc.getSubscription(bson.M{"_id": bson.ObjectIdHex(id)})
}

func (mc *MongoClient) AddSubscription(s *models.Subscription) (bson.ObjectId, error) {
	return mc.addSubscription(s)
}

func (mc *MongoClient) DeleteSubscriptionBySlug(slug string) error {
	ms, err := mc.SubscriptionBySlug(slug)
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}
	return mc.deleteByObjectID(ms.ID, SUBSCRIPTION_COLLECTION)
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
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
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

func (mc *MongoClient) CleanupOld(age int64) error {
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	end := currentTime - age
	query := bson.M{"modified": bson.M{"$lt": end}}
	mns, err := mc.getNotifications(query)
	if err == mgo.ErrNotFound {
		return ErrNotFound
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
	s := mc.GetSessionCopy()
	defer s.Close()

	n.Created = time.Now().UnixNano() / int64(time.Millisecond)
	n.ID = bson.NewObjectId()

	// Handle DBRefs
	mn := MongoNotification{Notification: *n}

	err := mc.checkNotificationSlugIntegrity(mn.Slug)
	if err != nil {
		return n.ID, err
	}

	err = s.DB(mc.Database.Name).C(NOTIFICATION_COLLECTION).Insert(mn)
	if err != nil {
		return n.ID, err
	}

	return n.ID, err
}

func (mc *MongoClient) checkNotificationSlugIntegrity(slug string) error {
	if slug == "" {
		return ErrSlugEmpty
	}
	_, err := mc.getNotificaiton(bson.M{"slug": slug})
	if err == nil {
		return ErrNotUnique
	}
	return nil
}

func (mc *MongoClient) updateNotification(n models.Notification) error {
	s := mc.GetSessionCopy()
	defer s.Close()

	n.Modified = time.Now().UnixNano() / int64(time.Millisecond)

	// Handle DBRef
	mn := MongoNotification{Notification: n}

	err := s.DB(mc.Database.Name).C(NOTIFICATION_COLLECTION).UpdateId(mn.ID, mn)
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}

	return err
}

func (mc *MongoClient) getNotificaiton(q bson.M) (models.Notification, error) {
	s := mc.GetSessionCopy()
	defer s.Close()

	// Handle DBRef
	var mn MongoNotification
	err := s.DB(mc.Database.Name).C(NOTIFICATION_COLLECTION).Find(q).One(&mn)
	if err == mgo.ErrNotFound {
		return mn.Notification, ErrNotFound
	}

	return mn.Notification, err
}

func (mc *MongoClient) getNotifications(q bson.M) ([]models.Notification, error) {
	return mc.getNotificationsLimit(q, currentReadMaxLimit)
}

func (mc *MongoClient) getNotificationsLimit(q bson.M, limit int) ([]models.Notification, error) {

	s := mc.GetSessionCopy()
	defer s.Close()

	// Handle DBRefs
	var mn []MongoNotification
	notifications := []models.Notification{}
	// Check if limit is 0
	if limit == 0 {
		return notifications, nil
	}
	err := s.DB(mc.Database.Name).C(NOTIFICATION_COLLECTION).Find(q).Limit(limit).All(&mn)
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
	s := mc.GetSessionCopy()
	defer s.Close()

	sub.Created = time.Now().UnixNano() / int64(time.Millisecond)
	sub.ID = bson.NewObjectId()

	// Handle DBRefs
	ms := MongoSubscription{Subscription: *sub}

	err := mc.checkSubscriptionSlugIntegrity(ms.Slug)
	if err != nil {
		return sub.ID, err
	}

	err = s.DB(mc.Database.Name).C(SUBSCRIPTION_COLLECTION).Insert(ms)
	if err != nil {
		return sub.ID, err
	}

	return sub.ID, err
}

func (mc *MongoClient) checkSubscriptionSlugIntegrity(slug string) error {
	if slug == "" {
		return ErrSlugEmpty
	}
	_, err := mc.getSubscription(bson.M{"slug": slug})
	if err == nil {
		return ErrNotUnique
	}
	return nil
}

func (mc *MongoClient) updateSubscription(sub models.Subscription) error {
	s := mc.GetSessionCopy()
	defer s.Close()

	sub.Modified = time.Now().UnixNano() / int64(time.Millisecond)

	// Handle DBRef
	ms := MongoSubscription{Subscription: sub}

	err := s.DB(mc.Database.Name).C(SUBSCRIPTION_COLLECTION).UpdateId(ms.ID, ms)
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}

	return err
}

func (mc *MongoClient) getSubscription(q bson.M) (models.Subscription, error) {
	s := mc.GetSessionCopy()
	defer s.Close()

	// Handle DBRef
	var ms MongoSubscription
	err := s.DB(mc.Database.Name).C(SUBSCRIPTION_COLLECTION).Find(q).One(&ms)
	if err == mgo.ErrNotFound {
		return ms.Subscription, ErrNotFound
	}

	return ms.Subscription, err
}

func (mc *MongoClient) getSubscriptions(q bson.M) ([]models.Subscription, error) {
	return mc.getSubscriptionsLimit(q, currentReadMaxLimit)
}

func (mc *MongoClient) getSubscriptionsLimit(q bson.M, limit int) ([]models.Subscription, error) {

	s := mc.GetSessionCopy()
	defer s.Close()

	// Handle DBRefs
	var ms []MongoSubscription
	subscriptions := []models.Subscription{}
	// Check if limit is 0
	if limit == 0 {
		return subscriptions, nil
	}
	err := s.DB(mc.Database.Name).C(SUBSCRIPTION_COLLECTION).Find(q).Limit(limit).All(&ms)
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
	s := mc.GetSessionCopy()
	defer s.Close()

	tran.Created = time.Now().UnixNano() / int64(time.Millisecond)
	tran.ID = bson.NewObjectId()

	// Handle DBRefs
	mt := MongoTransmission{Transmission: *tran}

	err := s.DB(mc.Database.Name).C(TRANSMISSION_COLLECTION).Insert(mt)
	if err != nil {
		return tran.ID, err
	}

	return tran.ID, err
}

func (mc *MongoClient) updateTransmission(tran models.Transmission) error {
	s := mc.GetSessionCopy()
	defer s.Close()

	tran.Modified = time.Now().UnixNano() / int64(time.Millisecond)

	// Handle DBRef
	mt := MongoTransmission{Transmission: tran}

	err := s.DB(mc.Database.Name).C(TRANSMISSION_COLLECTION).UpdateId(mt.ID, mt)
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}

	return err
}

func (mc *MongoClient) getTransmission(q bson.M) (models.Transmission, error) {
	s := mc.GetSessionCopy()
	defer s.Close()

	// Handle DBRef
	var mt MongoTransmission
	err := s.DB(mc.Database.Name).C(TRANSMISSION_COLLECTION).Find(q).One(&mt)
	if err == mgo.ErrNotFound {
		return mt.Transmission, ErrNotFound
	}

	return mt.Transmission, err
}

func (mc *MongoClient) getTransmissionsLimit(q bson.M) ([]models.Transmission, error) {
	s := mc.GetSessionCopy()
	defer s.Close()

	// Handle DBRefs
	var mt []MongoTransmission
	trans := []models.Transmission{}
	err := s.DB(mc.Database.Name).C(TRANSMISSION_COLLECTION).Find(q).All(&mt)
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
		return ErrInvalidObjectId
	}

	return mc.deleteByObjectID(bson.ObjectIdHex(id), col)

}

func (mc *MongoClient) deleteByObjectID(id bson.ObjectId, col string) error {
	s := mc.GetSessionCopy()
	defer s.Close()

	err := s.DB(mc.Database.Name).C(col).RemoveId(id)
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}
	return err
}

func (mc *MongoClient) deleteAll(q bson.M, col string) error {
	s := mc.GetSessionCopy()
	defer s.Close()

	_, err := s.DB(mc.Database.Name).C(col).RemoveAll(q)
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}
	return err
}
