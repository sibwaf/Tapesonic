package model

type TrackSourceDescriptor struct {
	LocalPath   string
	LocalFormat string
	LocalCodec  string

	RemoteUrl        string
	SourceDurationMs int64

	StartOffsetMs int64
	EndOffsetMs   int64
}
