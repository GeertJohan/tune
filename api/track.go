package api

type Track struct {
	Name          string            `json:"track"`
	Type          string            `json:"type"`
	Duration      int               `json:"duration"`
	Started       int               `json:"started"`
	ArtURL        string            `json:"art_url"`
	Artist        string            `json:"artist"`
	ChannelID     int               `json:"channel_id"`
	DisplayArtist string            `json:"display_artist"`
	DisplayTitle  string            `json:"display_title"`
	Images        map[string]string `json:"images"`
	Length        int               `json:"length"`
	NetworkID     int               `json:"network_id"`
	Release       string            `json:"release"`
	Title         string            `json:"title"`
	TrackID       int               `json:"track_id"`
}

func (t *Track) ArtURLHTTPS() string {
	return "https:" + t.ArtURL
}
