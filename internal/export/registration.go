//
// Copyright (c) 2017
// Mainflux
// IOTech
// Dell
//
// SPDX-License-Identifier: Apache-2.0
//

package export

import (
	"fmt"

	"github.com/edgexfoundry/edgex-go/pkg/models"

	"github.com/globalsign/mgo/bson"
)

// Export destination types
const (
	DestMQTT        = "MQTT_TOPIC"
	DestZMQ         = "ZMQ_TOPIC"
	DestIotCoreMQTT = "IOTCORE_TOPIC"
	DestAzureMQTT   = "AZURE_TOPIC"
	DestRest        = "REST_ENDPOINT"
	DestXMPP        = "XMPP_TOPIC"
	DestAWSMQTT     = "AWS_TOPIC"
	DestInfluxDB    = "INFLUXDB_ENDPOINT"
)

// Compression algorithm types
const (
	CompNone = "NONE"
	CompGzip = "GZIP"
	CompZip  = "ZIP"
)

// Data format types
const (
	FormatJSON            = "JSON"
	FormatXML             = "XML"
	FormatSerialized      = "SERIALIZED"
	FormatIoTCoreJSON     = "IOTCORE_JSON"
	FormatAzureJSON       = "AZURE_JSON"
	FormatAWSJSON         = "AWS_JSON"
	FormatCSV             = "CSV"
	FormatThingsBoardJSON = "THINGSBOARD_JSON"
	FormatNOOP            = "NOOP"
)

const (
	NotifyUpdateAdd    = "add"
	NotifyUpdateUpdate = "update"
	NotifyUpdateDelete = "delete"
)

// Registration - Defines the registration details
// on the part of north side export clients
type Registration struct {
	ID          bson.ObjectId      `bson:"_id,omitempty" json:"id,omitempty"`
	Created     int64              `json:"created"`
	Modified    int64              `json:"modified"`
	Origin      int64              `json:"origin"`
	Name        string             `json:"name,omitempty"`
	Addressable models.Addressable `json:"addressable,omitempty"`
	Format      string             `json:"format,omitempty"`
	Filter      Filter             `json:"filter,omitempty"`
	Encryption  EncryptionDetails  `json:"encryption,omitempty"`
	Compression string             `json:"compression,omitempty"`
	Enable      bool               `json:"enable"`
	Destination string             `json:"destination,omitempty"`
}

// Filter - Specifies the client filters on reading data
type Filter struct {
	DeviceIDs          []string `bson:"deviceIdentifiers,omitempty" json:"deviceIdentifiers,omitempty"`
	ValueDescriptorIDs []string `bson:"valueDescriptorIdentifiers,omitempty" json:"valueDescriptorIdentifiers,omitempty"`
}

func (reg Registration) Validate() (bool, error) {

	if reg.Name == "" {
		return false, fmt.Errorf("Name is required")
	}

	if reg.Compression == "" {
		reg.Compression = CompNone
	}

	if reg.Compression != CompNone &&
		reg.Compression != CompGzip &&
		reg.Compression != CompZip {
		return false, fmt.Errorf("Compression invalid: %s", reg.Compression)
	}

	if reg.Format != FormatJSON &&
		reg.Format != FormatXML &&
		reg.Format != FormatSerialized &&
		reg.Format != FormatIoTCoreJSON &&
		reg.Format != FormatAzureJSON &&
		reg.Format != FormatAWSJSON &&
		reg.Format != FormatCSV &&
		reg.Format != FormatThingsBoardJSON &&
		reg.Format != FormatNOOP {
		return false, fmt.Errorf("Format invalid: %s", reg.Format)
	}

	if reg.Destination != DestMQTT &&
		reg.Destination != DestZMQ &&
		reg.Destination != DestIotCoreMQTT &&
		reg.Destination != DestAzureMQTT &&
		reg.Destination != DestAWSMQTT &&
		reg.Destination != DestRest &&
		reg.Destination != DestInfluxDB {
		return false, fmt.Errorf("Destination invalid: %s", reg.Destination)
	}

	if reg.Encryption.Algo == "" {
		reg.Encryption.Algo = EncNone
	}

	if reg.Encryption.Algo != EncNone &&
		reg.Encryption.Algo != EncAes {
		return false, fmt.Errorf("Encryption invalid: %s", reg.Encryption.Algo)
	}

	return true, nil
}
