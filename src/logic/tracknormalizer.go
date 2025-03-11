package logic

import (
	"fmt"
	"regexp"
	"strings"
	"tapesonic/util"
)

type TrackNormalizer struct {
}

func NewTrackNormalizer() *TrackNormalizer {
	return &TrackNormalizer{}
}

const (
	FORMAT_UNKNOWN = iota
	FORMAT_ARTIST_DASH_TITLE
)

type openClosePair struct {
	open  string
	close string
}

var parenthesesOptions = []openClosePair{
	{open: "(", close: ")"},
	{open: "[", close: "]"},
	{open: "「", close: "」"},
	{open: "【", close: "】"},
	{open: "（", close: "）"},
}

var removeJunkSuffixRegex = buildJunkSuffixRegex(
	"official lyric video",
	"official music video",
	"official audio",
	"official video",
	"official visualizer",
	"official visualiser",
	"english subtitles",
	"mv",
	"pv",
	"w/lyrics",
	"lyrics",
	"audio",
	"audio only",
	"subbed",
	"hd",
	"official",
	"360º",
	"music video",
	"full album stream",
	"lyric video",
)

func (normalizer *TrackNormalizer) Normalize(tracks []TrackProperties) ([]TrackProperties, error) {
	result := make([]TrackProperties, len(tracks))
	copy(result, tracks)

	requireGuessingIndices := []int{}
	guessingSamples := []string{}

	for i := range result {
		artist := strings.TrimSpace(util.Coalesce(result[i].Artist, result[i].AlbumArtist))
		title := strings.TrimSpace(util.Coalesce(result[i].Title, result[i].RawTitle))

		if artist != "" {
			removeArtistFromTitleRegex, err := regexp.Compile(fmt.Sprintf("^%s\\s+-\\s+(.+)", regexp.QuoteMeta(artist)))
			if err != nil {
				return []TrackProperties{}, fmt.Errorf("failed to compile regex for artist removal: %w", err)
			}

			if match := removeArtistFromTitleRegex.FindStringSubmatch(title); match != nil {
				title = match[1]
			}

			result[i].Artist = artist
			result[i].Title = title
		} else {
			requireGuessingIndices = append(requireGuessingIndices, i)
			guessingSamples = append(guessingSamples, title)
		}
	}

	if len(requireGuessingIndices) > 0 {
		format := guessTitleFormat(guessingSamples)
		for _, index := range requireGuessingIndices {
			artist, title := extractArtistAndTitle(result[index].RawTitle, format)

			result[index].Artist = artist
			result[index].Title = title
		}
	}

	for i := range result {
		titleWithoutJunkSuffix := strings.TrimSpace(removeJunkSuffixRegex.ReplaceAllString(result[i].Title, ""))
		if titleWithoutJunkSuffix != "" {
			result[i].Title = titleWithoutJunkSuffix
		}
	}

	return result, nil
}

func guessTitleFormat(samples []string) int {
	allContainDash := true
	for _, sample := range samples {
		if !strings.Contains(sample, " - ") {
			allContainDash = false
		}
	}

	if allContainDash {
		return FORMAT_ARTIST_DASH_TITLE
	} else {
		return FORMAT_UNKNOWN
	}
}

func extractArtistAndTitle(text string, format int) (string, string) {
	if format == FORMAT_ARTIST_DASH_TITLE {
		if artist, title, ok := strings.Cut(text, " - "); ok {
			return strings.TrimSpace(artist), strings.TrimSpace(title)
		}
	}

	return "", strings.TrimSpace(text)
}

func buildJunkSuffixRegex(suffixes ...string) *regexp.Regexp {
	suffixOptions := []string{}
	for _, suffix := range suffixes {
		suffixRegex := regexp.QuoteMeta(suffix)
		suffixRegex = strings.ReplaceAll(suffixRegex, " ", "\\s+?")

		for _, parentheses := range parenthesesOptions {
			openRegex := regexp.QuoteMeta(parentheses.open)
			closeRegex := regexp.QuoteMeta(parentheses.close)

			// non-capture group of: OPEN any-space-count SUFFIX any-space-count CLOSE
			suffixOptions = append(suffixOptions, fmt.Sprintf("(?:%s\\s*?%s\\s*?%s)", openRegex, suffixRegex, closeRegex))
		}
	}

	// case-insensitive: OPTION or OPTION ... any-space-count end-of-string
	return regexp.MustCompile(fmt.Sprintf("(?i)%s\\s*?$", strings.Join(suffixOptions, "|")))
}
