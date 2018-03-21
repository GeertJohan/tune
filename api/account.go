package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCantAuthenticate   = errors.New("can't authenticate to server")
)

type Account struct {
	ID        int
	APIKey    string
	ListenKey string
	Firstname string
	Lastname  string
	Premium   bool
	Favorites []int
}

func (n *Network) AuthenticateUserPass(username, password string) (*Account, error) {
	return n.authenticate(url.Values{"username": {username}, "password": {password}})
}

func (n *Network) AuthenticateAPIKey(apikey string) (*Account, error) {
	return n.authenticate(url.Values{"api_key": {apikey}})
}

func (n *Network) authenticate(authValues url.Values) (*Account, error) {
	type favResult struct {
		ChannelID int `json:"channel_id"`
	}
	type subResult struct {
		Status string `json:"status"`
	}
	var authResult struct {
		Confirmed     bool        `json:"confirmed"`
		ID            int         `json:"id"`
		APIKey        string      `json:"api_key"`
		ListenKey     string      `json:"listen_key"`
		Firstname     string      `json:"first_name"`
		Lastname      string      `json:"last_name"`
		Favorites     []favResult `json:"network_favorite_channels"`
		Subscriptions []subResult `json:"subscriptions"`
	}

	apiURL := fmt.Sprintf(`%s/%s/members/authenticate`, APIBaseURL, n.Key)

	resp, err := http.PostForm(apiURL, authValues)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		return nil, ErrInvalidCredentials
	}
	if resp.StatusCode != 200 {
		fmt.Printf("statuscode: %d %s\n", resp.StatusCode, resp.Status)
		bts, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("%s\n", string(bts))
		return nil, ErrCantAuthenticate
	}
	err = json.NewDecoder(resp.Body).Decode(&authResult)
	if err != nil {
		return nil, err
	}

	a := &Account{
		ID:        authResult.ID,
		APIKey:    authResult.APIKey,
		ListenKey: authResult.ListenKey,
		Firstname: authResult.Firstname,
		Lastname:  authResult.Lastname,
	}
	for _, favRes := range authResult.Favorites {
		a.Favorites = append(a.Favorites, favRes.ChannelID)
	}
	for _, subRes := range authResult.Subscriptions {
		// TODO: the checks here should probably be more thourough.. i.e.: what services have status active!?
		if subRes.Status == "active" {
			a.Premium = true
			break
		}
	}

	return a, nil
}

// IsFavoriteChannel returns whether the channel (provided by id) is favorited in the account.
func (a *Account) IsFavoriteChannel(id int) bool {
	for _, favoriteID := range a.Favorites {
		if id == favoriteID {
			return true
		}
	}
	return false
}

// API return structure:
// {
// 	"api_key":"<your api key>",
// 	"confirmed":true,
// 	"email":"<your account email address>",
// 	"first_name":"Geert-Johan",
// 	"fraudulent":null,
// 	"id":66910,
// 	"last_name":"Riemer",
// 	"listen_key":"<your listen key>",
// 	"timezone":"Eastern Time (US & Canada)",
// 	"network_favorite_channels":[
// 		{
// 			"channel_id":56,
// 			"position":0
// 		},
// 		{
// 			"channel_id":3,
// 			"position":1
// 		},
// 		{
// 			"channel_id":177,
// 			"position":2
// 		},
// 		{
// 			"channel_id":184,
// 			"position":3
// 		},
// 		{
// 			"channel_id":105,
// 			"position":4
// 		},
// 		{
// 			"channel_id":182,
// 			"position":5
// 		},
// 		{
// 			"channel_id":198,
// 			"position":6
// 		},
// 		{
// 			"channel_id":174,
// 			"position":7
// 		},
// 		{
// 			"channel_id":64,
// 			"position":8
// 		},
// 		{
// 			"channel_id":36,
// 			"position":9
// 		},
// 		{
// 			"channel_id":181,
// 			"position":10
// 		},
// 		{
// 			"channel_id":13,
// 			"position":11
// 		},
// 		{
// 			"channel_id":123,
// 			"position":12
// 		},
// 		{
// 			"channel_id":12,
// 			"position":13
// 		},
// 		{
// 			"channel_id":59,
// 			"position":14
// 		},
// 		{
// 			"channel_id":57,
// 			"position":15
// 		},
// 		{
// 			"channel_id":117,
// 			"position":16
// 		},
// 		{
// 			"channel_id":53,
// 			"position":17
// 		},
// 		{
// 			"channel_id":10,
// 			"position":18
// 		},
// 		{
// 			"channel_id":1,
// 			"position":19
// 		}
// 	],
// 	"activated":true,
// 	"subscriptions":[
// 		{
// 			"auto_renew":true,
// 			"id":62141,
// 			"status":"active",
// 			"starts_on":"2011-03-29",
// 			"expires_on":"2015-06-03",
// 			"trial":false,
// 			"billable":true,
// 			"services":[
// 				{
// 					"id":1,
// 					"key":"di-premium",
// 					"name":"Digitally Imported Premium Radio"
// 				},
// 				{
// 					"id":2,
// 					"key":"sky-premium",
// 					"name":"SKY.FM Premium Radio"
// 				},
// 				{
// 					"id":3,
// 					"key":"jazzradio-premium",
// 					"name":"JAZZRADIO.com Premium Radio"
// 				},
// 				{
// 					"id":4,
// 					"key":"rockradio-premium",
// 					"name":"ROCKRADIO.com Premium Radio"
// 				},
// 				{
// 					"id":5,
// 					"key":"radiotunes-premium",
// 					"name":"RadioTunes.com Premium Radio"
// 				}
// 			],
// 			"plan":{
// 				"allow_trial":true,
// 				"availability_ends_at":"2014-03-01T00:00:00-05:00",
// 				"availability_starts_at":"2011-04-21T01:53:00-04:00",
// 				"created_at":"2011-04-21T01:53:48-04:00",
// 				"id":1,
// 				"key":"premium-pass",
// 				"name":"Premium Radio",
// 				"trial_duration_days":7,
// 				"updated_at":"2014-02-28T22:35:49-05:00"
// 			}
// 		}
// 	],
// 	"email_campaigns":[
// 		{
// 			"auto_opt_in":true,
// 			"created_at":"2013-08-07T00:09:59-04:00",
// 			"id":1,
// 			"key":"di_newsletter",
// 			"name":"Monthly Newsletter",
// 			"position":1,
// 			"sailthru_name":"di Monthly Newsletter",
// 			"updated_at":"2013-08-07T00:09:59-04:00",
// 			"network":{
// 				"key":"di"
// 			}
// 		},
// 		{
// 			"auto_opt_in":true,
// 			"created_at":"2013-08-07T00:09:59-04:00",
// 			"id":2,
// 			"key":"di_offers",
// 			"name":"Offers and Promotions",
// 			"position":2,
// 			"sailthru_name":"di Offers and Promotions",
// 			"updated_at":"2013-08-07T00:09:59-04:00",
// 			"network":{
// 				"key":"di"
// 			}
// 		},
// 		{
// 			"auto_opt_in":true,
// 			"created_at":"2013-08-07T00:09:59-04:00",
// 			"id":3,
// 			"key":"di_alerts",
// 			"name":"Special Event Alerts",
// 			"position":3,
// 			"sailthru_name":"di Special Event Alerts",
// 			"updated_at":"2013-08-07T00:09:59-04:00",
// 			"network":{
// 				"key":"di"
// 			}
// 		}
// 	],
// 	"feature_flags":[

// 	],
// 	"images":{

// 	}
// }
