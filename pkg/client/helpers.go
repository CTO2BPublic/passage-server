package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog/log"
)

func (a *ApiClient) doRequest(Req ClientRequest) (ResponseData []byte, StatusCode int, Err error) {

	apiurl := a.Url + Req.ApiEndpoint
	bytearr, err := json.Marshal(Req.Body)
	if err != nil {
		log.Err(err).Msg("Err")
	}
	client := &http.Client{
		Transport: &http.Transport{},
	}
	req, err := http.NewRequest(Req.Method, apiurl, bytes.NewBuffer(bytearr))
	if err != nil {
		log.Err(err).Msg("Err")
	}

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", strings.Replace(a.Token, "\n", "", -1)))

	response, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer response.Body.Close()

	responseData, _ := io.ReadAll(response.Body)

	log.Debug().
		Str("Method", Req.Method).
		Str("Url", apiurl).
		Msgf("Response %s", responseData)

	return responseData, response.StatusCode, err
}

func NewStaticToken(id string) string {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    "passage-server",
		NotBefore: jwt.NewNumericDate(now),
		ID:        id,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(Config.SharedSecret))

	return signed
}
