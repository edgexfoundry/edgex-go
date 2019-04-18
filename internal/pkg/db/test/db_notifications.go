//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0
//

package test

import (
	"fmt"
	"testing"

	dbp "github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/interfaces"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func TestNotificationsDB(t *testing.T, db interfaces.DBClient) {
	// Remove previous notification and transmission
	err := db.Cleanup()
	if err != nil {
		t.Fatalf("Error clean db: %v\n", err)
	}
	cleanUpAllSubscription(db)

	testDBNotification(t, db)
	testDBSubscription(t, db)
	testDBTransmission(t, db)

	defer db.CloseSession()
	// Calling CloseSession twice to test that there is no panic when closing an
	// already closed db
	defer db.CloseSession()
}

func testDBNotification(t *testing.T, db interfaces.DBClient) {
	// Prepare test data
	beforeTime := dbp.MakeTimestamp()
	err := populateNotification(db, 0, 10, contract.New)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}
	err = populateNotification(db, 10, 20, contract.Escalated)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}
	err = populateNotification(db, 20, 30, contract.Processed)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}
	afterTime := dbp.MakeTimestamp()

	// Test GetNotifications
	notifications, err := db.GetNotifications()
	if err != nil {
		t.Fatalf("Error getting notifications %v", err)
	}
	if len(notifications) != 30 {
		t.Fatalf("There should be 30 notifications instead of %d", len(notifications))
	}

	// Test GetNotificationById
	slugName := "slug-test"
	notification := getNotification(slugName, contract.New)
	notification.ID, err = db.AddNotification(notification)
	if err != nil {
		t.Fatalf("Fail to add notification")
	}
	n, err := db.GetNotificationById(notification.ID)
	if err != nil {
		t.Fatalf("Fail to getting notification by id %v", err)
	}
	if n.ID != notification.ID {
		t.Fatalf("ID does not match %s - %s", n.ID, notification.ID)
	}

	// Test GetNotificationBySlug
	n, err = db.GetNotificationBySlug(slugName)
	if err != nil {
		t.Fatalf("Error getting notification by slug %v", err)
	}
	if n.Slug != slugName {
		t.Fatalf("Slug does not match %s - %s", n.Slug, slugName)
	}

	// Test GetNotificationBySender
	sender := "test-sender"
	notifications, err = db.GetNotificationBySender(sender, 5)
	if err != nil {
		t.Fatalf("Error getting notifications by sender: %v", err)
	}
	if len(notifications) == 0 {
		t.Fatalf("There should be at least one notification")
	}
	if notifications[0].Sender != sender {
		t.Fatalf("Sender does not match %s - %s", n.Sender, sender)
	}

	// Test GetNotificationsByLabels
	labels := []string{"labelA", "labelB"}
	notifications, err = db.GetNotificationsByLabels(labels, 5)
	if err != nil {
		t.Fatalf("Error getting notifications %v", err)
	}
	if len(notifications) == 0 {
		t.Fatalf("There should be at least one notification")
	}

	// Test GetNotificationsByStartEnd
	notifications, err = db.GetNotificationsByStartEnd(beforeTime, afterTime, 5)
	if err != nil {
		t.Fatalf("Error getting notifications %v", err)
	}
	if len(notifications) != 5 {
		t.Fatalf("There should be five notifications")
	}

	// Test GetNotificationsByStart
	notifications, err = db.GetNotificationsByStart(beforeTime, 5)
	if err != nil {
		t.Fatalf("Error getting notifications %v", err)
	}
	if len(notifications) != 5 {
		t.Fatalf("There should be five notifications")
	}

	// Test GetNotificationsByEnd
	notifications, err = db.GetNotificationsByEnd(afterTime, 5)
	if err != nil {
		t.Fatalf("Error getting notifications %v", err)
	}
	if len(notifications) != 5 {
		t.Fatalf("There should be five notifications")
	}

	// Test GetNewNotifications
	notifications, err = db.GetNewNotifications(5)
	if err != nil {
		t.Fatalf("Error getting notifications %v", err)
	}
	if len(notifications) != 5 {
		t.Fatalf("There should be five notifications")
	}

	// Test GetNewNormalNotifications
	notifications, err = db.GetNewNormalNotifications(5)
	if err != nil {
		t.Fatalf("Error getting notifications %v", err)
	}
	if len(notifications) != 5 {
		t.Fatalf("There should be five notifications")
	}

	// Test MarkNotificationProcessed
	err = db.MarkNotificationProcessed(notification)
	if err != nil {
		t.Fatalf("Fail to mark notification to processed status, %v", err)
	}
	n, err = db.GetNotificationBySlug(slugName)
	if err != nil {
		t.Fatalf("Error getting notification by slug %v", err)
	}
	if n.Status != contract.Processed {
		t.Fatalf("Notification status should be %v ", contract.Processed)
	}

	// Test DeleteNotificationBySlug
	err = db.DeleteNotificationBySlug(notification.Slug)
	if err != nil {
		t.Fatalf("Fail to delete notification by slug '%v'", notification.Slug)
	}

	// Test DeleteNotificationsOld
	err = db.DeleteNotificationsOld(0)
	if err != nil {
		t.Fatalf("Fail to delete old notifications, '%v'", err)
	}

}

func testDBSubscription(t *testing.T, db interfaces.DBClient) {
	// Prepare test data
	err := populateSubscription(db, 30)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	// Test GetSubscriptions
	subscriptions, err := db.GetSubscriptions()
	if err != nil {
		t.Fatalf("Error getting notifications %v", err)
	}
	if len(subscriptions) != 30 {
		t.Fatalf("There should be 30 notifications instead of %d", len(subscriptions))
	}

	// Test GetSubscriptionById
	slugName := "slug-test"
	subscription := getSubscription(slugName)
	subscription.ID, err = db.AddSubscription(subscription)
	if err != nil {
		t.Fatalf("Fail to add subscription, %v", err)
	}
	s, err := db.GetSubscriptionById(subscription.ID)
	if err != nil {
		t.Fatalf("Fail to get subscription by ID, %v", err)
	}
	if s.Slug != slugName {
		t.Fatalf("Unexpect test result, slug '%v' not match '%v'", s.Slug, slugName)
	}

	// Test GetSubscriptionBySlug
	s, err = db.GetSubscriptionBySlug(slugName)
	if err != nil {
		t.Fatalf("Fail to get subscription by slug, %v", err)
	}
	if s.Slug != slugName {
		t.Fatalf("Unexpect test result, slug '%v' not match '%v'", s.Slug, slugName)
	}

	// Test GetSubscriptionByReceiver
	receiverName := "test-receiver"
	subscriptions, err = db.GetSubscriptionByReceiver(receiverName)
	if err != nil {
		t.Fatalf("Fail to get subscription by receiver, %v", err)
	}
	if subscriptions[0].Receiver != receiverName {
		t.Fatalf("Unexpect test result, receiver '%v' not match '%v'", subscriptions[0].Receiver, receiverName)
	}

	// Test GetSubscriptionByCategories
	categories := []string{contract.Security, contract.Hwhealth}
	_, err = db.GetSubscriptionByCategories(categories)
	if err != nil {
		t.Fatalf("Fail to get subscription by categories, %v", err)
	}

	// Test GetSubscriptionByLabels
	labels := []string{"labelA", "labelB"}
	_, err = db.GetSubscriptionByLabels(labels)
	if err != nil {
		t.Fatalf("Fail to get subscription by labels, %v", err)
	}

	// Test GetSubscriptionByCategoriesLabels
	categories = []string{contract.Hwhealth, contract.Swhealth}
	labels = []string{"labelA"}
	_, err = db.GetSubscriptionByCategoriesLabels(categories, labels)
	if err != nil {
		t.Fatalf("Fail to get subscription by categories and labels, %v", err)
	}
}

func testDBTransmission(t *testing.T, db interfaces.DBClient) {
	// Prepare test data
	err := populateTransmission(db, 30, 10)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}

	// Test UpdateTransmission
	slugName := "slug-test"
	transmission := getTransmission(slugName, 10)
	transmission.ID, err = db.AddTransmission(transmission)
	if err != nil {
		t.Fatalf("Fail to add subscription, %v", err)
	}
	transmission.Status = contract.Failed
	err = db.UpdateTransmission(transmission)
	if err != nil {
		t.Fatalf("Fail to update transmission, %v", err)
	}
	transmissions, err := db.GetTransmissionsByStatus(0, contract.Failed)
	if transmissions[0].Status != contract.Failed {
		t.Fatalf("Unexpect test result. Transmission status '%s' not match %s", transmissions[0].Status, contract.Failed)
	}

	// Test GetTransmissionsByNotificationSlug
	transmissions, err = db.GetTransmissionsByNotificationSlug(slugName, 10)
	if err != nil {
		t.Fatalf("Fail to get transmission by notification slug, %v", err)
	}

	if transmissions[0].Notification.Slug != slugName {
		t.Fatalf("Unexpect test result. Slug '%v' not match '%v'", transmissions[0].Notification.Slug, slugName)
	}

	// Test GetTransmissionsByStartEnd
	resendCount := 2
	amount := 10
	beforeTime := dbp.MakeTimestamp()
	err = populateTransmission(db, amount, resendCount)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}
	afterTime := dbp.MakeTimestamp()

	transmissions, err = db.GetTransmissionsByStartEnd(beforeTime, afterTime, resendCount)
	if err != nil {
		t.Fatalf("Fail to get transmission by start time and end time, %v", err)
	}
	if len(transmissions) != amount {
		t.Fatalf("Unexpect result. The amount of transmissions should be %v, but actually is %v", amount, len(transmissions))
	}

	// Test GetTransmissionsByStart
	transmissions, err = db.GetTransmissionsByStart(beforeTime, resendCount)
	if err != nil {
		t.Fatalf("Fail to get transmission by start time, %v", err)
	}
	if len(transmissions) != amount {
		t.Fatalf("Unexpect result. The amount of transmissions should be %v, but actually is %v", amount, len(transmissions))
	}

	// Test GetTransmissionsByEnd
	transmissions, err = db.GetTransmissionsByEnd(afterTime, resendCount)
	if err != nil {
		t.Fatalf("Fail to get transmission by start time, %v", err)
	}
	if len(transmissions) != amount {
		t.Fatalf("Unexpect result. The amount of transmissions should be %v, but actually is %v", amount, len(transmissions))
	}

	// Test DeleteTransmission
	beforeTime = dbp.MakeTimestamp()
	err = populateTransmission(db, 5, 1)
	if err != nil {
		t.Fatalf("Error populating db: %v\n", err)
	}
	afterTime = dbp.MakeTimestamp()
	err = db.DeleteTransmission(afterTime-beforeTime, contract.Sent)
	if err != nil {
		t.Fatalf("Fail to delete old transmission, %v", err)
	}
}

func getNotification(slug string, status contract.NotificationsStatus) contract.Notification {
	n := contract.Notification{}
	n.Slug = slug
	n.Sender = "test-sender"
	n.Category = contract.Hwhealth
	n.Severity = contract.Normal
	n.Content = "The machine is running for 25 days."
	n.Labels = []string{"labelA", "labelB"}
	n.Status = status
	n.Description = "Notify running time"
	return n
}

func populateNotification(db interfaces.DBClient, from int, to int, status contract.NotificationsStatus) error {
	for i := from; i < to; i++ {
		n := getNotification(fmt.Sprintf("slug-%d", i), status)
		_, err := db.AddNotification(n)
		if err != nil {
			return err
		}
	}
	return nil
}

func getSubscription(slug string) contract.Subscription {
	s := contract.Subscription{}
	s.Slug = slug
	s.Receiver = "test-receiver"
	s.Description = "Subscription test"
	s.SubscribedCategories = []contract.NotificationsCategory{contract.Security, contract.Hwhealth, contract.Swhealth}
	s.SubscribedLabels = []string{"labelA", "labelB"}
	return s
}

func cleanUpAllSubscription(db interfaces.DBClient) error {
	subscriptions, err := db.GetSubscriptions()
	if err != nil {
		return err
	}
	for _, s := range subscriptions {
		err = db.DeleteSubscriptionBySlug(s.Slug)
		if err != nil {
			return err
		}
	}
	return nil
}

func populateSubscription(db interfaces.DBClient, count int) error {
	for i := 0; i < count; i++ {
		s := getSubscription(fmt.Sprintf("slug-%d", i))
		_, err := db.AddSubscription(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func getTransmission(slug string, resendCount int) contract.Transmission {
	t := contract.Transmission{}
	t.Notification = contract.Notification{Slug: slug}
	t.Receiver = "test-receiver"
	t.Status = contract.Sent
	t.ResendCount = resendCount
	return t
}

func populateTransmission(db interfaces.DBClient, count int, resendCount int) error {
	for i := 0; i < count; i++ {
		t := getTransmission(fmt.Sprintf("slug-%d", i), resendCount)
		_, err := db.AddTransmission(t)
		if err != nil {
			return err
		}
	}
	return nil
}
