package client

import (
	"encoding/json"
	"fmt"

	"github.com/CTO2BPublic/passage-server/pkg/models"
)

type ClientRequest struct {
	Body        interface{}
	ApiEndpoint string
	Method      string
}

type ClientResponse struct {
	Status int    `example:"201"`
	Title  string `example:"Record successfully created"`
	Type   string `example:"/status/success"`
	Error  string `example:"Missing required fields"`
}

type GetAccessRequestsOpts struct {
	Status string
}

func (c *ApiClient) GetAccessRequests(opts GetAccessRequestsOpts) (response []models.AccessRequest, statusCode int, err error) {

	req := ClientRequest{
		ApiEndpoint: "/access/requests",
		Method:      "GET",
	}

	return processRequest[[]models.AccessRequest](c, req)

}

type ExpireAccessRequestOpts struct {
	Id string
}

func (c *ApiClient) ExpireAccessRequest(opts ExpireAccessRequestOpts) (response ClientResponse, statusCode int, err error) {

	req := ClientRequest{
		ApiEndpoint: fmt.Sprintf("/access/requests/%s/expire", opts.Id),
		Method:      "POST",
	}

	return processRequest[ClientResponse](c, req)

}

// Generic processRequest function
func processRequest[T any](c *ApiClient, req ClientRequest) (T, int, error) {
	data, statusCode, err := c.doRequest(req)
	if err != nil || statusCode != 200 {
		return *new(T), statusCode, fmt.Errorf("unexpected API response code: %d, body: %s err: %s", statusCode, string(data), err)
	}

	var result T
	err = json.Unmarshal(data, &result)
	if err != nil {
		return *new(T), statusCode, fmt.Errorf("failed to parse api response: %d, body: %s err: %s", statusCode, string(data), err)
	}

	return result, statusCode, nil
}
