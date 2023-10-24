package storage

type TapeInfo struct {
	Id     string
	Name   string
	Author string
	Tracks []TrackInfo
}

type TrackInfo struct {
	Name     string
	Index    int
	OffsetMs int
	LengthMs int
}

type StreamableTrack struct {
	Path  string
	Track TrackInfo
}

type Cover struct {
	Path string
}
