package distro

import (
	"github.com/Shopify/sarama"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"strconv"
	"strings"
	"sync"
)

type kafkaPublisher struct {
	publisher sarama.SyncProducer
	topic     string
	mux       sync.Mutex
}

func newKAFKASender(addr models.Addressable) sender {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	kafkaAddrs := strings.Split(addr.Address, ",")
	newPublisher, err := sarama.NewSyncProducer(kafkaAddrs, config)
	if err != nil {
		LoggingClient.Warn("newKAFKASender failed " + addr.Address + addr.Topic)
	}
	sender := &kafkaPublisher{
		publisher: newPublisher,
		topic:     addr.Topic,
	}
	LoggingClient.Info("newKAFKASender success")
	return sender
}

func (sender *kafkaPublisher) Send(data []byte, event *models.Event) bool {
	const (
		edgeKafkaKey = "edge-kafka-key"
	)
	sender.mux.Lock()
	defer sender.mux.Unlock()
	msg := &sarama.ProducerMessage{
		Topic:     sender.topic,
		Partition: int32(-1),
		Key:       sarama.StringEncoder(edgeKafkaKey),
		Value:     sarama.ByteEncoder(data),
	}
	paritition, offset, err := sender.publisher.SendMessage(msg)

	if err != nil {
		LoggingClient.Warn("Send Message Fail Topic: " + sender.topic)
		return false
	}

	LoggingClient.Info("Partion = " + strconv.Itoa(int(paritition)) + "offset = " + strconv.Itoa(int(offset)))
	return true
}
