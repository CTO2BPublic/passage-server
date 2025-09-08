package eventdriver

import (
	"context"
	"fmt"

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
		_, err := Kafka.NewClient()
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Kafka client")
		}
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
		go func() {
			if err := Kafka.WriteMessage(data, txid); err != nil {
				log.Error().Err(err).Msg("Failed to write event to Kafka")
			}
		}()
	}

	if Config.Events.Database.Enabled {
		if err := Db.InsertEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to insert event in database: %w", err)
		}

		activityLog, err := models.NewActivityLogFromEvent(event)
		if err == nil {
			if err := Db.InsertActivityLog(ctx, *activityLog); err != nil {
				return fmt.Errorf("failed to insert activity log in database: %w", err)
			}
		}
	}

	return nil
}
