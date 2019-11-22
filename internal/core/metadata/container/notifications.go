package container

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/di"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"
)

// NotificationsClientName contains the name of the NotificationsClient's implementation in the DIC.
var NotificationsClientName = di.TypeInstanceToName((*notifications.NotificationsClient)(nil))

// NotificationsClientFrom helper function queries the DIC and returns the NotificationsClient's implementation.
func NotificationsClientFrom(get di.Get) notifications.NotificationsClient {
	return get(NotificationsClientName).(notifications.NotificationsClient)
}
