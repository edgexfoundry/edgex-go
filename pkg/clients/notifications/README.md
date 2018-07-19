# README #
This package contains the notifications client written in the Go programming language.  The notifications client is used by Go services or other Go code to communicate with the EdgeX support-notifications microservice (regardless of underlying implemenation type) by sending REST requests to the service's API endpoints.

### How To Use ###
To use the support-notifications client package you first need to import the library into your project:
```
import "github.com/edgexfoundry/edgex-go/pkg/clients/notifications"
```
To send a notification you first need to get a NotificationsClient and then send a Notification struct:
```
		notification := notifications.Notification{
			Slug:        configuration.NotificationsSlug + strconv.FormatInt((time.Now().UnixNano()/int64(time.Millisecond)), 10),
			Content:     configuration.NotificationContent + name + "-" + string(action),
			Category:    notifications.SW_HEALTH,
			Description: configuration.NotificationDescription,
			Labels:      []string{configuration.NotificationLabel},
			Sender:      configuration.NotificationSender,
			Severity:    notifications.NORMAL,
		}

		notifications.GetNotificationsClient().SendNotification(notification)
```
This will send the notification to the notifications service.
