package crondriver

import (
	"time"

	"github.com/CTO2BPublic/passage-server/pkg/client"
	"github.com/CTO2BPublic/passage-server/pkg/models"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

func (c *Cron) Start() {

	apiClient := client.NewApiClient(client.ApiClientOpts{
		Url:   "http://localhost:8080",
		Token: client.NewStaticToken("internal-cron"),
	})

	cron := cron.New()

	_, err := cron.AddFunc("@every 60s", func() {
		Requests, _, err := apiClient.GetAccessRequests(client.GetAccessRequestsOpts{})
		if err != nil {
			log.Err(err).Msg("Failed to fetch access requests")
			return
		}

		log.Debug().Msgf("Fetched %d access requests", len(Requests))
		processAccessRequests(apiClient, Requests)
	})
	if err != nil {
		log.Error().Str("/errors/cron", "Failed to schedule cron entry").Msg(err.Error())
	}

	go cron.Start()
}

func processAccessRequests(apiClient *client.ApiClient, Requests []models.AccessRequest) {
	now := time.Now()
	for _, request := range Requests {
		expiration := request.Status.ExpiresAt
		if expiration != nil && now.After(*expiration) && request.Status.Status != models.AccessRequestExpired {

			log.Info().
				Str("Request", request.Id).
				Str("Role", request.RoleRef.Name).
				Str("Requester", request.Status.RequestedBy).
				Str("Expires", request.Status.ExpiresAt.Local().String()).
				Str("TTL", request.Details.TTL).
				Msg("Revoking access request")

			apiClient.ExpireAccessRequest(client.ExpireAccessRequestOpts{
				Id: request.Id,
			})
		}
	}
}
