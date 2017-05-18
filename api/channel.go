package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrChannelRequiresPremium is returned by (*Channel).StreamURLs() when the channel is only for premium accounts and given account is not a premium account.
	ErrChannelRequiresPremium = errors.New("channel requires a premium account")

	// ErrNoTracklist is returned when the API returns a valid result, but the tracklist is empty (or only contains ads)
	ErrNoTracklist = errors.New("no tracklist found")
)

// Channel contains information about a AudioAddict channel
type Channel struct {
	// Network on which the channel resides
	Network *Network
	// Streamlist to which the channel belongs (audio quality)
	Streamlist *Streamlist
	// ID is the numerical reference to this channel. It seems to be unique per channel (quality-agnostic)
	ID int `json:"id"`
	// Key is the lowercase name for this channel
	Key string `json:"key"`
	// Name is the human readable name for this channel
	Name string `json:"name"`
	// Playlist is the playlist URL that can be used by the media player.
	Playlist string `json:"playlist"`
}

// StreamURLs returns a list of stream URL's
func (c *Channel) StreamURLs(acc *Account) ([]string, error) {
	if c.Streamlist.Premium && (acc == nil || !acc.Premium) {
		return nil, ErrChannelRequiresPremium
	}
	url := fmt.Sprintf("%s/%s/%s", c.Network.ListenURLBase, c.Streamlist.Key, c.Key) // e.g.: http://listen.di.fm/premium_high/dub?25*censor*cf51
	if acc != nil {
		url += `?` + acc.ListenKey
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	servers := make([]string, 0)
	err = json.NewDecoder(resp.Body).Decode(&servers)
	if err != nil {
		return nil, err
	}
	return servers, nil
}

type Tracklist []*Track

func (c *Channel) tracklist() (Tracklist, error) {
	url := fmt.Sprintf(`%s/%s/track_history/channel/%d`, APIBaseURL, c.Network.Key, c.ID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	tracklist := make(Tracklist, 0)
	err = json.NewDecoder(resp.Body).Decode(&tracklist)
	if err != nil {
		return nil, err
	}
	return tracklist, nil
}

func (c *Channel) Tracklist() (Tracklist, error) {
	tracklist, err := c.tracklist()
	if err != nil {
		return nil, err
	}
	//++ TODO: cleanup tracklist from ads
	return tracklist, nil
}

func (c *Channel) CurrentTrack() (*Track, error) {
	tracklist, err := c.tracklist()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(tracklist); i++ {
		track := tracklist[i]
		if track.Type == "track" {
			return track, nil
		}
	}
	return nil, ErrNoTracklist
}
