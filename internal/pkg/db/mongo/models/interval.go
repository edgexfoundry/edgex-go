package models

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
)

type Interval struct {
	Id        bson.ObjectId `bson:"_id"`
	Uuid      string        `bson:"uuid"`
	Created   int64         `bson:"created"`
	Modified  int64         `bson:"modified"`
	Origin    int64         `bson:"origin"`
	Name      string        `bson:"name"`
	Start     string        `bson:"start"`
	End       string        `bson:"end"`
	Frequency string        `bson:"frequency"`
	Cron      string        `bson:"cron"`
	RunOnce   bool          `bson:"runonce"`
}

func (in Interval) ToContract() contract.Interval {
	// Always hand back the UUID as the contract event ID unless it'in blank (an old event, for example blackbox test scripts
	id := in.Uuid
	if id == "" {
		id = in.Id.Hex()
	}

	to := contract.Interval{
		ID:        id,
		Created:   in.Created,
		Modified:  in.Modified,
		Origin:    in.Origin,
		Name:      in.Name,
		Start:     in.Start,
		End:       in.End,
		Frequency: in.Frequency,
		Cron:      in.Cron,
		RunOnce:   in.RunOnce,
	}
	return to
}

func (in *Interval) FromContract(from contract.Interval) error {

	var err error
	in.Id, in.Uuid, err = fromContractId(from.ID)
	if err != nil {
		return err
	}

	in.Created = from.Created
	in.Modified = from.Modified
	in.Origin = from.Origin
	in.Name = from.Name
	in.Start = from.Start
	in.End = from.End
	in.Frequency = from.Frequency
	in.RunOnce = from.RunOnce
	in.Cron = from.Cron

	// if not created
	if in.Created == 0 {
		in.Created = db.MakeTimestamp()
	}

	return nil
}

// Custom mongo marshalling
func (in Interval) GetBSON() (interface{}, error) {
	return struct {
		ID        bson.ObjectId `bson:"_id,omitempty"`
		Uuid      string        `bson:"uuid,omitempty"`
		Created   int64         `bson:"created"`
		Modified  int64         `bson:"modified"`
		Origin    int64         `bson:"origin"`
		Name      string        `bson:"name"`
		Start     string        `bson:"start"`
		End       string        `bson:"end"`
		Frequency string        `bson:"frequency"`
		Cron      string        `bson:"cron"`
		RunOnce   bool          `bson:"runonce"`
	}{
		ID:        in.Id,
		Uuid:      in.Uuid,
		Created:   in.Created,
		Modified:  in.Modified,
		Origin:    in.Origin,
		Name:      in.Name,
		Start:     in.Start,
		End:       in.End,
		Frequency: in.Frequency,
		Cron:      in.Cron,
		RunOnce:   in.RunOnce,
	}, nil
}

// Custom unmarshaling out of MongoDB
func (in *Interval) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		ID        bson.ObjectId `bson:"_id,omitempty"`
		Uuid      string        `bson:"uuid,omitempty"`
		Created   int64         `bson:"created"`
		Modified  int64         `bson:"modified"`
		Origin    int64         `bson:"origin"`
		Name      string        `bson:"name"`
		Start     string        `bson:"start"`
		End       string        `bson:"end"`
		Frequency string        `bson:"frequency"`
		Cron      string        `bson:"cron,omitempty"`
		RunOnce   bool          `bson:"runonce"`
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the fields
	in.Id = decoded.ID
	in.Uuid = decoded.Uuid
	in.Created = decoded.Created
	in.Modified = decoded.Modified
	in.Origin = decoded.Origin
	in.Name = decoded.Name
	in.Start = decoded.Start
	in.End = decoded.End
	in.Frequency = decoded.Frequency
	in.Cron = decoded.Cron
	in.RunOnce = decoded.RunOnce

	return nil
}

// no DBRefs so no custom mapping required
