package aaapi

import (
	"errors"
)

var (
	// ErrStreamlistNotAvailable is returned by (*Network).StreamlistByKey() when there is no streamlist with given key
	ErrStreamlistNotAvailable = errors.New("streamlist not available")
)

// Network defines the parameters for an AudioAddict network such as di.fm
// It provides API methods to get channels and stream information
type Network struct {
	// Name is the human readable name for the network, e.g.: "RadioTunes"
	Name string

	// ListenURLBase is the base URL string to be used with certain API calls
	//++ TODO: unexport?
	ListenURLBase string

	// Key is to be used with certain API calls
	//++ TODO: unexport?
	Key string

	// streamlists is a slice of Streamlist's available on this network
	Streamlists []*Streamlist
	// best streamlists
	bestStreamlist        *Streamlist
	bestStreamlistPremium *Streamlist
}

func (n *Network) addStreamlist(s *Streamlist) {
	s.Network = n
	n.Streamlists = append(n.Streamlists, s)
	if s.Premium {
		if n.bestStreamlistPremium == nil || (n.bestStreamlistPremium.Bitrate < s.Bitrate) {
			n.bestStreamlistPremium = s
		}
	} else {
		if n.bestStreamlist == nil || (n.bestStreamlist.Bitrate < s.Bitrate) {
			n.bestStreamlist = s
		}
	}
}

// BestStreamlist returns the best quality Streamlist for the network.
// When premium is available (true), it will return the best premium Streamlist
func (n *Network) BestStreamlist(premium bool) *Streamlist {
	if premium {
		return n.bestStreamlistPremium
	}
	return n.bestStreamlist
}

// StreamlistByKey looks up the correct streamlist for given key.
// When none is found, ErrStreamlistNotAvailable is returned.
func (n *Network) StreamlistByKey(key string) (*Streamlist, error) {
	for _, sl := range n.Streamlists {
		if sl.Key == key {
			return sl, nil
		}
	}
	return nil, ErrStreamlistNotAvailable
}
