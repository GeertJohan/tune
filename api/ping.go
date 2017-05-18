package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// PingInfo holds the information that is sent back by the audioaddict api server on a ping request.
type PingInfo struct {
	APIVersion float64
	Time       time.Time // NOTE: server time is converted to a time.Time value in the UTC timezone.
	IP         string
	Country    string
}

// Ping pings the server, this returns some information about the location of this client, and provides the client with the current servertime.
func Ping() (*PingInfo, error) {
	resp, err := http.Get(`http://api.audioaddict.com/v1/ping`)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var apiInfo struct {
		APIVersion float64 `json:"api_version"`
		Time       string  `json:"time"`
		IP         string  `json:"ip"`
		Country    string  `json:"country"`
	}
	err = json.NewDecoder(resp.Body).Decode(&apiInfo)
	if err != nil {
		return nil, err
	}
	t, err := time.Parse(`Mon, 2 Jan 2006 15:04:05 -0700`, apiInfo.Time)
	if err != nil {
		return nil, err
	}
	pi := &PingInfo{
		APIVersion: apiInfo.APIVersion,
		Time:       t.In(time.UTC),
		IP:         apiInfo.IP,
		Country:    apiInfo.Country,
	}
	return pi, nil
}
