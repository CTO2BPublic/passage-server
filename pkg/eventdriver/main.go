package eventdriver

import (
	"context"

	"github.com/CTO2BPublic/passage-server/pkg/kafkadriver"
	"github.com/CTO2BPublic/passage-server/pkg/shared"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
)

type Events struct {
}

var Driver = new(Events)
var Config = config.GetConfig().Events
var Kafka = kafkadriver.GetDriver()
var Tracer = otel.Tracer("eventsdriver")

// Setup Initialize Kafka client
func (e *Events) NewDriver() *Events {

	if Config.Kafka.Enabled {
		Kafka.NewClient()
	}

	return e
}

func GetDriver() *Events {
	return Driver
}

func (e *Events) handleEvent(ctx context.Context, data interface{}) error {

	txid, _ := shared.GetTransactionID(ctx)
	event, _ := data.(Event)

	if Config.Console.Enabled {
		log.Info().Msgf("[Event] Stdout: txid: %s message: %s", txid, event.Message)
	}

	if Config.Kafka.Enabled {
		go Kafka.WriteMessage(data, txid)
	}

	return nil
}
