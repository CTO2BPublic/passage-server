package client

import (
	"github.com/CTO2BPublic/passage-server/pkg/config"
)

var Config = config.GetConfig()

type ApiClient struct {
	Url   string
	Token string
}

type ApiClientOpts struct {
	Url   string
	Token string
}

func NewApiClient(opts ApiClientOpts) *ApiClient {
	Client := new(ApiClient)

	Client.Url = opts.Url
	Client.Token = opts.Token

	return Client
}
