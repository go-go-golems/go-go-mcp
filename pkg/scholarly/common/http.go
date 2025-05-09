package common

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Response represents a standardized HTTP response
type Response struct {
	StatusCode int
	Body       []byte
	Error      error
}

// MakeHTTPRequest performs an HTTP request and returns a standardized response
func MakeHTTPRequest(req *http.Request) Response {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	response := Response{}

	if err != nil {
		log.Error().Err(err).Str("url", req.URL.String()).Msg("Error making HTTP request")
		response.Error = err
		return response
	}
	defer resp.Body.Close()

	response.StatusCode = resp.StatusCode
	log.Debug().Str("status", resp.Status).Int("statusCode", resp.StatusCode).Msg("HTTP response received")

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Error().Str("status", resp.Status).Bytes("body", body).Msg("HTTP request failed")
		response.Body = body
		return response
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error reading HTTP response body")
		response.Error = err
		return response
	}

	response.Body = body
	return response
}