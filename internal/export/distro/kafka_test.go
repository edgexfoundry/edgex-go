package distro

import (
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"testing"
)

func TestKafkaSend(t *testing.T) {
	const (
		msgStr = "this is a test msg from edgex"
	)
	addr := models.Addressable{
		Address: "127.0.0.1:9092",
		Port:    9092,
		Topic:   "edgex-kafka-sender-test",
	}
	sender := newKAFKASender(addr)
	var msg = []byte(msgStr)
	sender.Send(msg, nil)

}
