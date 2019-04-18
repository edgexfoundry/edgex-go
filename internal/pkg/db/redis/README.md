# Redis Implementation Notes

## Core Data

### Events and Readings

`pkg/models/event` is stored as a marshalled value using the event id as the key.

Sorted sets are used to track events sorted by most recently added, created timestamp (both relative to all events and scoped to the source device), and the pushed timestamp. Readings are stored similarly in their own sorted sets.

Given

```go
var e Event
```

and

```go
var e models.Event = models.Event{
  ID: "57ba04a1189b95b8afcdafd7",
  Pushed: 1471806399999,
  Device: "123456789",
  Created: 1464039917100,
  Modified: 1474774741088,
  Origin: 1471806386919,
  Event: "",
  Reading: []Reading{
  Reading {
    ID: "57b9fe08189b95b8afcdafd4",
    Pushed: 0,
    Created: 1471806984866,
    Modified: 1471807130870,
    Origin: 1471806386919,
    Name: "temperature",
    Value: "39"
  },
  Reading {
    ID: "57e745efe4b0ca8e6d7116d7",
    Pushed: 0,
    Created: 1474774511737,
    Modified: 1474774511737,
    Origin: 1471806386919,
    Name: "power",
    Value: "38"
  },
  ...
  }
}
```

then

| Sorted Set Key           | Score         | Member                   |
| ------------------------ | ------------- | ------------------------ |
| event                    | 0             | 57ba04a1189b95b8afcdafd7 |
| event:created            | 1464039917100 | 57ba04a1189b95b8afcdafd7 |
| event:pushed             | 1471806399999 | 57ba04a1189b95b8afcdafd7 |
| event:device:123456789   | 1464039917100 | 57ba04a1189b95b8afcdafd7 |
| reading                  | 0             | 57b9fe08189b95b8afcdafd4 |
| reading:created          | 1471806984866 | 57b9fe08189b95b8afcdafd4 |
| reading:device:123456789 | 1471806984866 | 57b9fe08189b95b8afcdafd4 |
| reading:name:temperature | 1471806984866 | 57b9fe08189b95b8afcdafd4 |
| reading                  | 0             | 57e745efe4b0ca8e6d7116d7 |
| reading:created          | 1474774511737 | 57e745efe4b0ca8e6d7116d7 |
| reading:device:123456789 | 1474774511737 | 57e745efe4b0ca8e6d7116d7 |
| reading:name:power       | 1474774511737 | 57e745efe4b0ca8e6d7116d7 |

## Notification Service

Each of Notification, Subscription, and Transmission objects are stored as a key/value pair where the key is the id of the object.  The value is JSON marshalled string.

Given the migration away from BSON ids and toward UUID, all generated ids are UUIDs.

Changes:
* Add test at internal/pkg/db/test/db_notifications.go
* Create a separate folder at internal/pkg/db/redis_notification for migrating to master branch
    * Implement db interface function according to internal/support/notifications/interfaces/db.go

### Notifications

Notifications are queried by Slug, sender, time, labels, and status. To support those queries a sorted set with the index (e.g. slug) as the score and the value as the key of the notification object in question.

| Data Type   | Key                                | Value                  |  Score    |
|-------------|------------------------------------|------------------------|-----------|
| Sets        | Entity ID                          | Entity                 |           |  
| Sorted sets | "notification"                     | Entity ID              | 0         |
| Hashes      | "notification:slug"                | SlugName and Entity ID |           |
| Sorted sets | "notification:sender:{sender}"     | Entity ID              | 0         |
| Sorted sets | "notification:status:{status}"     | Entity ID              | 0         |
| Sorted sets | "notification:severity:{severity}" | Entity ID              | 0         |
| Sorted sets | "notification:created"             | Entity ID              | Timestamp |
| Sorted sets | "notification:modified"            | Entity ID              | Timestamp |

Where LABEL and STATUS are a specific label or status, respectively.

### Subscriptions

Subscriptions are queried by Slug, categories, labels, and receiver.

| Data Type   | Key                                | Value                  |  Score    |
|-------------|------------------------------------|------------------------|-----------|
| Sets        | Entity ID                          | Entity                 |           |  
| Sorted sets | "subscription"                     | Entity ID              | 0         |
| Hashes      | "subscription:slug"                | SlugName and Entity ID |           |
| Sorted sets | "subscription:receiver:{receiver}" | Entity ID              | 0         |
| Sets        | "subscription:label:{label}"       | Entity ID              |           |
| Sets        | "subscription:category:{category}" | Entity ID              |           |

Given the migration of EdgeX to UUID from BSON Id, UUID is used for the id in the Redis implementation of the Notification service.

### Transmission

| Data Type   | Key                                      | Value                  |  Score      |
|-------------|------------------------------------------|------------------------|-------------|
| Sets        | Entity ID                                | Entity                 |             |  
| Sorted sets | "transmission"                           | Entity ID              | 0           |
| Sorted sets | "transmission:slug:{slug}"               | Entity ID              | ResendCount |
| Sorted sets | "transmission:status:{status}"           | Entity ID              | ResendCount |
| Sorted sets | "transmission:resendcount:{resendcount}" | Entity ID              | ResendCount |
| Sorted sets | "transmission:created"                   | Entity ID              | Timestamp   |
| Sorted sets | "transmission:modified"                  | Entity ID              | Timestamp   |


