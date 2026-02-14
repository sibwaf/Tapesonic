package search

import "errors"

var (
	ErrMetadataMismatch = errors.New("actual metadata doesn't match expected")
)

type TrackForSearch struct {
	Artist string
	Title  string
	Album  string

	Urls []string
}
