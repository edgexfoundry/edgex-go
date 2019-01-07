package models

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
)

type IntervalAction struct {
	Created    int64         `bson:"created"`
	Modified   int64         `bson:"modified"`
	Origin     int64         `bson:"origin"`
	Id         bson.ObjectId `bson:"_id,omitempty"`
	Uuid       string        `bson:"uuid,omitempty"`
	Name       string        `bson:"name"`
	Interval   string        `bson:"interval"`
	Parameters string        `bson:"parameters"`
	Target     string        `bson:"target"`
	Protocol   string        `bson:"protocol"`
	HTTPMethod string        `bson:"httpMethod"`
	Address    string        `bson:"address"`
	Port       int           `bson:"port"`
	Path       string        `bson:"path"`
	Publisher  string        `bson:"publisher"`
	User       string        `bson:"user"`
	Password   string        `bson:"password"`
	Topic      string        `bson:"topic"`
}

func (ia *IntervalAction) ToContract() (c contract.IntervalAction) {
	// Always hand back the UUID as the contract event ID unless it's blank (an old event, for example blackbox test scripts
	id := ia.Uuid
	if id == "" {
		id = ia.Id.Hex()
	}

	c.ID = id
	c.Created = ia.Created
	c.Modified = ia.Modified
	c.Origin = ia.Origin
	c.Name = ia.Name
	c.Interval = ia.Interval
	c.Parameters = ia.Parameters
	c.Target = ia.Target
	c.Protocol = ia.Protocol
	c.HTTPMethod = ia.HTTPMethod
	c.Address = ia.Address
	c.Port = ia.Port
	c.Path = ia.Path
	c.Publisher = ia.Publisher
	c.User = ia.User
	c.Password = ia.Password
	c.Topic = ia.Topic

	return
}

func (ia *IntervalAction) FromContract(from contract.IntervalAction) (id string, err error) {
	ia.Id, ia.Uuid, err = fromContractId(from.ID)
	if err != nil {
		return
	}

	ia.Created = from.Created
	ia.Modified = from.Modified
	ia.Origin = from.Origin
	ia.Name = from.Name
	ia.Interval = from.Interval
	ia.Parameters = from.Parameters
	ia.Target = from.Target
	ia.Protocol = from.Protocol
	ia.HTTPMethod = from.HTTPMethod
	ia.Address = from.Address
	ia.Port = from.Port
	ia.Path = from.Path
	ia.Publisher = from.Publisher
	ia.User = from.User
	ia.Password = from.Password
	ia.Topic = from.Topic

	id = toContractId(ia.Id, ia.Uuid)
	return
}

func (ia *IntervalAction) TimestampForUpdate() {
	ia.Modified = db.MakeTimestamp()
}

func (ia *IntervalAction) TimestampForAdd() {
	ia.TimestampForUpdate()
	ia.Created = ia.Modified
}
