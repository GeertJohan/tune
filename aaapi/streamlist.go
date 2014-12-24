package aaapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Streamlist defines
type Streamlist struct {
	// Network is a reference to the network on which this streamlist resides.
	Network *Network

	// Key is the identifier to be used with API calls.
	Key string

	// Premium indicates wether this streamlist can only be used by premium accounts.
	Premium bool

	// Bitrate is the stream bitrate in kbit/s.
	Bitrate int

	// Encoding is the stream encoding.
	Encoding Encoding
}

// Name returns the human-readable name for this streamlist
func (sl *Streamlist) Name() string {
	var premium string
	if sl.Premium {
		premium += "Premium "
	}
	return fmt.Sprintf("%s%dkbit/s %s", premium, sl.Bitrate, sl.Encoding)
}

// Channels returns a list of channels available on this Streamlist
func (sl *Streamlist) Channels() ([]*Channel, error) {
	// fetch channels from server
	url := fmt.Sprintf("%s/%s", sl.Network.ListenURLBase, sl.Key)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// decode into channels slice
	channels := make([]*Channel, 0)
	err = json.NewDecoder(resp.Body).Decode(&channels)
	if err != nil {
		return nil, err
	}

	// setup references
	for _, ch := range channels {
		ch.Network = sl.Network
		ch.Streamlist = sl
	}

	// all done
	return channels, nil
}

// Encoding defines the type of audio encoding/compression that is used
type Encoding string

var (
	EncodingAAC = Encoding("AAC")
	EncodingMP3 = Encoding("MP3")
	EncodingWMA = Encoding("WMA")
)
