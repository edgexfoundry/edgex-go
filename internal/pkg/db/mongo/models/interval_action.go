package models

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo/bson"
)

type IntervalAction struct {
	Id         bson.ObjectId `bson:"_id"`
	Uuid       string        `bson:"uuid"`
	Created    int64         `bson:"created"`
	Modified   int64         `bson:"modified"`
	Origin     int64         `bson:"origin"`
	Name       string        `bson:"bson"`
	Interval   string        `bson:"interval"`
	Parameters string        `bson:"parameters"`
	Target     string        `bson:"target"`
	Protocol   string        `bson:"protocol"`
	HTTPMethod string        `bson:"httpmethod"`
	Address    string        `bson:"address"`
	Port       int           `bson:"port"`
	Path       string        `bson:"path"`
	Publisher  string        `bson:"publisher"`
	User       string        `bson:"user"`
	Password   string        `bson:"password"`
	Topic      string        `bson:"topic"`
}

func (ia IntervalAction) ToContract() contract.IntervalAction {
	// Always hand back the UUID as the contract event ID unless it's blank (an old event, for example blackbox test scripts
	id := ia.Uuid
	if id == "" {
		id = ia.Id.Hex()
	}

	to := contract.IntervalAction{

		ID:         id,
		Created:    ia.Created,
		Modified:   ia.Modified,
		Origin:     ia.Origin,
		Name:       ia.Name,
		Interval:   ia.Interval,
		Parameters: ia.Parameters,
		Target:     ia.Target,
		Protocol:   ia.Protocol,
		HTTPMethod: ia.HTTPMethod,
		Address:    ia.Address,
		Port:       ia.Port,
		Path:       ia.Path,
		Publisher:  ia.Publisher,
		User:       ia.User,
		Password:   ia.Password,
		Topic:      ia.Topic,
	}
	return to
}

func (ia *IntervalAction) FromContract(from contract.IntervalAction) error {

	var err error
	ia.Id, ia.Uuid, err = fromContractId(from.ID)
	if err != nil {
		return err
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

	if ia.Created == 0 {
		ia.Created = db.MakeTimestamp()
	}

	return nil
}

func (ia IntervalAction) GetBSON() (interface{}, error) {
	return struct {
		ID         bson.ObjectId `bson:"_id,omitempty"`
		Uuid       string        `bson:"uuid,omitempty"`
		Created    int64         `bson:"created"`
		Modified   int64         `bson:"modified"`
		Origin     int64         `bson:"origin"`
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
	}{
		ID:         ia.Id,
		Uuid:       ia.Uuid,
		Created:    ia.Created,
		Modified:   ia.Modified,
		Origin:     ia.Origin,
		Name:       ia.Name,
		Interval:   ia.Interval,
		Parameters: ia.Parameters,
		Target:     ia.Target,
		Protocol:   ia.Protocol,
		HTTPMethod: ia.HTTPMethod,
		Address:    ia.Address,
		Port:       ia.Port,
		Path:       ia.Path,
		Publisher:  ia.Publisher,
		User:       ia.User,
		Password:   ia.Password,
		Topic:      ia.Topic,
	}, nil
}

func (ia *IntervalAction) SetBSON(raw bson.Raw) error {
	decoded := new(struct {
		ID         bson.ObjectId `bson:"_id,omitempty"`
		Uuid       string        `bson:"uuid,omitempty"`
		Created    int64         `bson:"created"`
		Modified   int64         `bson:"modified"`
		Origin     int64         `bson:"origin"`
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
	})

	bsonErr := raw.Unmarshal(decoded)
	if bsonErr != nil {
		return bsonErr
	}

	// Copy over the non-DBRef fields
	ia.Id = decoded.ID
	ia.Uuid = decoded.Uuid
	ia.Created = decoded.Created
	ia.Modified = decoded.Modified
	ia.Origin = decoded.Origin
	ia.Name = decoded.Name
	ia.Interval = decoded.Interval
	ia.Parameters = decoded.Parameters
	ia.Target = decoded.Target
	ia.Protocol = decoded.Protocol
	ia.HTTPMethod = decoded.HTTPMethod
	ia.Address = decoded.Address
	ia.Port = decoded.Port

	return nil
}
