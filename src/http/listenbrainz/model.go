package listenbrainz

const (
	ListenTypeSingle     = "single"
	ListenTypePlayingNow = "playing_now"
)

type SubmitListensRequest struct {
	ListenType string                            `json:"listen_type"`
	Payload    []SubmitListensRequestPayloadItem `json:"payload"`
}

type SubmitListensRequestPayloadItem struct {
	ListenedAt    int64                                        `json:"listened_at,omitempty"`
	TrackMetadata SubmitListensRequestPayloadItemTrackMetadata `json:"track_metadata"`
}

type SubmitListensRequestPayloadItemTrackMetadata struct {
	ArtistName  string `json:"artist_name"`
	TrackName   string `json:"track_name"`
	ReleaseName string `json:"release_name"`
}
