package models

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/google/uuid"
	"github.com/globalsign/mgo/bson"
)

type Reading struct {
	Id       bson.ObjectId `bson:"_id"`
	Uuid     string        `bson:"uuid"`
	Pushed   int64         `bson:"pushed"`  // When the data was pushed out of EdgeX (0 - not pushed yet)
	Created  int64         `bson:"created"` // When the reading was created
	Origin   int64         `bson:"origin"`
	Modified int64         `bson:"modified"`
	Device   string        `bson:"device"`
	Name     string        `bson:"name"`
	Value    string        `bson:"value"` // Device sensor data value
}

func (r Reading) ToContract() contract.Reading {
	// Always hand back the UUID as the contract event ID unless it's blank (an old event, for example blackbox test scripts)
	id := r.Uuid
	if id == "" {
		id = r.Id.Hex()
	}
	to := contract.Reading{
		Id:       id,
		Pushed:   r.Pushed,
		Created:  r.Created,
		Origin:   r.Origin,
		Modified: r.Modified,
		Device:   r.Device,
		Name:     r.Name,
		Value:    r.Value,
	}
	return to
}

func (r *Reading) FromContract(from contract.Reading) error {
	// In this first case, ID is empty so this must be an add.
	// Generate new BSON/UUIDs
	if from.Id == "" {
		r.Id = bson.NewObjectId()
		r.Uuid = uuid.New().String()
	} else {
		// In this case, we're dealing with an existing event
		if !bson.IsObjectIdHex(from.Id) {
			// EventID is not a BSON ID. Is it a UUID?
			_, err := uuid.Parse(from.Id)
			if err != nil { // It is some unsupported type of string
				return db.ErrInvalidObjectId
			}
			// Leave model's ID blank for now. We will be querying based on the UUID.
			r.Uuid = from.Id
		} else {
			// ID of pre-existing event is a BSON ID. We will query using the BSON ID.
			r.Id = bson.ObjectIdHex(from.Id)
		}
	}

	r.Pushed = from.Pushed
	r.Created = from.Created
	r.Origin = from.Origin
	r.Modified = from.Modified
	r.Device = from.Device
	r.Name = from.Name
	r.Value = from.Value

	if r.Created == 0 {
		r.Created = db.MakeTimestamp()
	}

	return nil
}
