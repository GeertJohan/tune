package api

var (
	// NetworkDI defines the network parameters for DI.fm radio.
	NetworkDI *Network

	// NetworkRadioTunes contains the network parameters for RadioTunes.com radio.
	NetworkRadioTunes *Network

	// NetworkList contains all networks defined by this package, it is possible to add/remove items from this list.
	NetworkList = []*Network{NetworkDI, NetworkRadioTunes}
)

func init() {
	NetworkDI = &Network{
		Name:          "di.fm",
		ListenURLBase: "http://listen.di.fm",
		Key:           "di",
	}
	NetworkDI.addStreamlist(&Streamlist{
		Key:      "public1",
		Bitrate:  64,
		Encoding: EncodingAAC,
	})
	NetworkDI.addStreamlist(&Streamlist{
		Key:      "public2",
		Bitrate:  40,
		Encoding: EncodingAAC,
	})
	NetworkDI.addStreamlist(&Streamlist{
		Key:      "public3",
		Bitrate:  96,
		Encoding: EncodingMP3,
	})
	NetworkDI.addStreamlist(&Streamlist{
		Key:      "premium_low",
		Premium:  true,
		Bitrate:  40,
		Encoding: EncodingAAC,
	})
	NetworkDI.addStreamlist(&Streamlist{
		Key:      "premium_medium",
		Premium:  true,
		Bitrate:  64,
		Encoding: EncodingAAC,
	})
	NetworkDI.addStreamlist(&Streamlist{
		Key:      "premium",
		Premium:  true,
		Bitrate:  128,
		Encoding: EncodingAAC,
	})
	NetworkDI.addStreamlist(&Streamlist{
		Key:      "premium_high",
		Premium:  true,
		Bitrate:  256,
		Encoding: EncodingMP3,
	})

	NetworkRadioTunes = &Network{
		Name:          "RadioTunes",
		ListenURLBase: "http://listen.radiotunes.com",
		Key:           "radiotunes",
		// Streamlists:   make(map[string]*Streamlist),
	}
	NetworkRadioTunes.addStreamlist(&Streamlist{
		Key:      "public1",
		Bitrate:  40,
		Encoding: EncodingAAC,
	})
	NetworkRadioTunes.addStreamlist(&Streamlist{
		Key:      "public5",
		Bitrate:  40,
		Encoding: EncodingWMA,
	})
	NetworkRadioTunes.addStreamlist(&Streamlist{
		Key:      "public3",
		Bitrate:  96,
		Encoding: EncodingMP3,
	})
	NetworkRadioTunes.addStreamlist(&Streamlist{
		Key:      "premium_low",
		Premium:  true,
		Bitrate:  40,
		Encoding: EncodingAAC,
	})
	NetworkRadioTunes.addStreamlist(&Streamlist{
		Key:      "premium_medium",
		Premium:  true,
		Bitrate:  64,
		Encoding: EncodingAAC,
	})
	NetworkRadioTunes.addStreamlist(&Streamlist{
		Key:      "premium",
		Premium:  true,
		Bitrate:  128,
		Encoding: EncodingAAC,
	})
	NetworkRadioTunes.addStreamlist(&Streamlist{
		Key:      "premium_high",
		Premium:  true,
		Bitrate:  256,
		Encoding: EncodingMP3,
	})
}
