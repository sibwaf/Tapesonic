package logic

import (
	"fmt"
	"strings"
	"tapesonic/util"
)

type TrackForMatching struct {
	Artist string
	Title  string
}

type TrackMatcher struct {
}

func NewTrackMatcher() *TrackMatcher {
	return &TrackMatcher{}
}

func (tm *TrackMatcher) Match(expected TrackForMatching, actual TrackForMatching) bool {
	if actual.Artist == "" {
		if matchText(expected.Title, actual.Title) {
			return true
		}
		if matchText(fmt.Sprintf("%s - %s", expected.Artist, expected.Title), actual.Title) {
			return true
		}
		if matchText(fmt.Sprintf("%s - %s", expected.Title, expected.Artist), actual.Title) {
			return true
		}
	} else {
		if matchText(expected.Artist, actual.Artist) && matchText(expected.Title, actual.Title) {
			return true
		}
	}

	return false
}

func matchText(expected string, actual string) bool {
	expectedWords := util.SplitWords(expected)
	actualWords := util.SplitWords(actual)

	if len(expectedWords) != len(actualWords) {
		return false
	}

	for i, expectedWord := range expectedWords {
		if !strings.EqualFold(expectedWord, actualWords[i]) {
			return false
		}
	}

	return true
}
