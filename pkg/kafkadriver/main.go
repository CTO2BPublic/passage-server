package kafkadriver

import (
	"encoding/json"
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
)

type Kafka struct {
	Conn       sarama.SyncProducer
	AuditTopic string
}

var Driver = new(Kafka)
var Config = config.GetConfig().Events.Kafka
var Tracer = otel.Tracer("pkg/kafkadriver")

// Setup Initialize Kafka client
func (k *Kafka) NewClient() *Kafka {

	log.Info().Msg("[Kafka] Creating client")

	// Kafka client config
	k.AuditTopic = Config.Topic

	// Producer config
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.RequiredAcks = sarama.WaitForLocal
	producer, err := sarama.NewSyncProducer([]string{Config.Hostname}, producerConfig)
	if err != nil {
		log.Error().Msgf("[Kafka] Failed to create producer: %s", err.Error())
	}

	// Create topic
	log.Info().Msgf("[Kafka] Creating topic: %s", k.AuditTopic)

	broker := sarama.NewBroker(Config.Hostname)
	brokerConfig := sarama.NewConfig()

	broker.Open(brokerConfig)

	request := sarama.CreateTopicsRequest{
		Timeout: time.Second * 15,
		TopicDetails: map[string]*sarama.TopicDetail{
			k.AuditTopic: {
				NumPartitions:     int32(Config.NumPartitions),
				ReplicationFactor: int16(Config.ReplicationFactor),
			},
		},
	}

	// Send request to Broker
	resp, err := broker.CreateTopics(&request)
	if err != nil {
		log.Error().Msgf("[Kafka] Failed to create topics: %s", err.Error())
	}

	if len(resp.TopicErrors) > 0 {
		for _, e := range resp.TopicErrors {
			log.Error().Msg(e.Unwrap().Error())
		}
	}

	k.Conn = producer
	return k
}

func GetDriver() *Kafka {
	return Driver
}

func (k *Kafka) WriteMessage(data interface{}, key string) error {

	value, _ := json.Marshal(&data)

	msg := &sarama.ProducerMessage{
		Topic: k.AuditTopic,
		Value: sarama.StringEncoder(string(value)),
		Key:   sarama.StringEncoder(key),
	}

	if k.Conn != nil {
		partition, offset, err := k.Conn.SendMessage(msg)
		if err != nil {
			log.Error().Msgf("[Kafka] %s", err.Error())
			return err
		}

		log.Info().Msgf("[Event] Kafka: Message sent to partition: %d, offset: %d", partition, offset)
	}

	return nil
}
