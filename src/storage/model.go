package storage

type Tape struct {
	Id       string `gorm:"primaryKey;not null"`
	Metadata string
	Url      string

	Name       string
	AuthorName string

	ThumbnailPath string

	Tracks []*TapeTrack
}

type TapeTrack struct {
	TapeId         string `gorm:"primaryKey;not null"`
	TapeTrackIndex int    `gorm:"primaryKey;not null"`

	FilePath string `gorm:"not null"`

	RawStartOffsetMs int
	StartOffsetMs    int
	RawEndOffsetMs   int
	EndOffsetMs      int

	Artist string
	Title  string

	Tape *Tape `gorm:"foreignKey:TapeId"`
}

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
