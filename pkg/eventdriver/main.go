package eventdriver

import (
	"context"

	"github.com/CTO2BPublic/passage-server/pkg/dbdriver"
	"github.com/CTO2BPublic/passage-server/pkg/kafkadriver"
	"github.com/CTO2BPublic/passage-server/pkg/models"
	"github.com/CTO2BPublic/passage-server/pkg/shared"

	"github.com/CTO2BPublic/passage-server/pkg/config"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
)

type Events struct {
}

var Driver = new(Events)
var Config = config.GetConfig()
var Kafka = kafkadriver.GetDriver()
var Tracer = otel.Tracer("eventsdriver")
var Db = dbdriver.GetDriver()

// Setup Initialize Kafka client
func (e *Events) NewDriver() *Events {

	if Config.Events.Kafka.Enabled {
		Kafka.NewClient()
	}

	return e
}

func GetDriver() *Events {
	return Driver
}

func (e *Events) handleEvent(ctx context.Context, data interface{}) error {

	txid, _ := shared.GetTransactionID(ctx)
	event, _ := data.(models.Event)

	if Config.Events.Console.Enabled {
		log.Info().Msgf("[Event] Stdout: txid: %s message: %s", txid, event.Message)
	}

	if Config.Events.Kafka.Enabled {
		go Kafka.WriteMessage(data, txid)
	}

	if Config.Events.Database.Enabled {
		Db.InsertEvent(ctx, event)
	}
	log.Debug().Msg("handleEvent called end") // 👈

	return nil
}
