package search

import (
	"fmt"
	"tapesonic/util"
)

type TrackForMatching struct {
	Artist string
	Title  string
}

type TrackMatcher struct {
}

func newTrackMatcher() *TrackMatcher {
	return &TrackMatcher{}
}

func (tm *TrackMatcher) Match(expected TrackForMatching, actual TrackForMatching) bool {
	if actual.Artist == "" {
		if util.MatchText(expected.Title, actual.Title) {
			return true
		}
		if util.MatchText(fmt.Sprintf("%s - %s", expected.Artist, expected.Title), actual.Title) {
			return true
		}
		if util.MatchText(fmt.Sprintf("%s - %s", expected.Title, expected.Artist), actual.Title) {
			return true
		}
	} else {
		if util.MatchText(expected.Artist, actual.Artist) && util.MatchText(expected.Title, actual.Title) {
			return true
		}
	}

	return false
}
