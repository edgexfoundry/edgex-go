# Redis Implementation Notes

`pkg/models/event` is stored as a marshalled value using the event id as the key. Sorted sets are used to track events sorted by most recently added, created timestamp (both relative to all events and scoped to the source device), and the pushed timestamp. Readings are stored are stored similarly in their own sorted sets.

Given

```
var e Event
```
and

```
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

| Sorted Set Key | Score  | Member |
|---|---|---|
| event | 0 | 57ba04a1189b95b8afcdafd7 |
| event:created | 1464039917100 | 57ba04a1189b95b8afcdafd7 |
| event:pushed | 1471806399999 | 57ba04a1189b95b8afcdafd7 |
| event:device:123456789 | 1464039917100 | 57ba04a1189b95b8afcdafd7 |
| reading | 0 | 57b9fe08189b95b8afcdafd4 |
| reading:created | 1471806984866 | 57b9fe08189b95b8afcdafd4 |
| reading:device:123456789 | 1471806984866 | 57b9fe08189b95b8afcdafd4 |
| reading:name:temperature | 1471806984866 | 57b9fe08189b95b8afcdafd4 |
| reading | 0 | 57e745efe4b0ca8e6d7116d7 |
| reading:created | 1474774511737 | 57e745efe4b0ca8e6d7116d7 |
| reading:device:123456789 | 1474774511737 | 57e745efe4b0ca8e6d7116d7 |
| reading:name:power | 1474774511737 | 57e745efe4b0ca8e6d7116d7 |
