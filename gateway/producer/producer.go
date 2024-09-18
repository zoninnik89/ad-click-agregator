package producer

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	common "github.com/zoninnik89/commons"
	"log"
	"math"
	"time"
)

var (
	KafkaServerAddress = common.EnvString("KAFKA_SERVER_ADDRESS", "localhost:9092")
	KafkaTopic         = "clicks"
)

type Producer struct {
	producer *kafka.Producer
}

func NewKafkaProducer() *Producer {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers":   KafkaServerAddress,
		"delivery.timeout.ms": "1",
		"acks":                "all", //0-no ack, 1-leader, all,
		"enable.idempotence":  "true",
	}
	p, err := kafka.NewProducer(configMap)

	if err != nil {
		log.Println(err.Error())
	}

	return &Producer{
		producer: p,
	}
}

func (p *Producer) Publish(msg string, topic string, key []byte, deliveryChan chan kafka.Event) error {
	message := &kafka.Message{
		Value:          []byte(msg),
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            key,
	}

	err := p.producer.Produce(message, deliveryChan)

	if err != nil {
		return err
	}

	return nil
}

func (p *Producer) DeliveryReport(deliveryChan chan kafka.Event) {
	for e := range deliveryChan {
		switch e.(type) {
		case *kafka.Message:
			e := <-deliveryChan
			msg := e.(*kafka.Message)

			if msg.TopicPartition.Error != nil {
				log.Println("Message was not published")
			} else {
				log.Println("Message published", msg.TopicPartition)
			}
		}
	}
}

func (p *Producer) Flush(timeoutMs int) int {
	termChan := make(chan bool)

	d, _ := time.ParseDuration(fmt.Sprintf("%dms", timeoutMs))
	tEnd := time.Now().Add(d)
	for p.Len() > 0 {
		remain := tEnd.Sub(time.Now()).Seconds()
		if remain <= 0.0 {
			return p.Len()
		}

		p.handle.eventPoll(p.events,
			int(math.Min(100, remain*1000)), 1000, termChan)
	}

	return 0
}
