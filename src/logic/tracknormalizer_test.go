package logic_test

import (
	"tapesonic/logic"
	"testing"
)

type artistAndTitle struct {
	Artist string
	Title  string
}

func TestNormalize_YoutubePlaylistAlbum(t *testing.T) {
	svc := logic.NewTrackNormalizer()

	raw := []logic.TrackProperties{
		{RawTitle: "Artist 1 - Song 1", ParentTitle: "Artist 1 - Album Title", Uploader: "Uploader"},
		{RawTitle: "Artist 1 - Song 2", ParentTitle: "Artist 1 - Album Title", Uploader: "Uploader"},
		{RawTitle: "Artist 1 - Song 3", ParentTitle: "Artist 1 - Album Title", Uploader: "Uploader"},
		{RawTitle: "Artist 1 - Song 4", ParentTitle: "Artist 1 - Album Title", Uploader: "Uploader"},
	}
	expected := []artistAndTitle{
		{Artist: "Artist 1", Title: "Song 1"},
		{Artist: "Artist 1", Title: "Song 2"},
		{Artist: "Artist 1", Title: "Song 3"},
		{Artist: "Artist 1", Title: "Song 4"},
	}

	normalized, err := svc.Normalize(raw)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	compareTracks(normalized, expected, t)
}

func TestNormalize_YoutubePlaylistAlbum_WithMetadata(t *testing.T) {
	svc := logic.NewTrackNormalizer()

	raw := []logic.TrackProperties{
		{RawTitle: "Song 1", ParentTitle: "Album Name", Artist: "Artist 1", Title: "Song 1", Uploader: "Artist 1 - Topic"},
		{RawTitle: "Song 2", ParentTitle: "Album Name", Artist: "Artist 1", Title: "Song 2", Uploader: "Artist 1 - Topic"},
		{RawTitle: "Song 3", ParentTitle: "Album Name", Artist: "Artist 1", Title: "Song 3", Uploader: "Artist 1 - Topic"},
		{RawTitle: "Song 4", ParentTitle: "Album Name", Artist: "Artist 1", Title: "Song 4", Uploader: "Artist 1 - Topic"},
	}
	expected := []artistAndTitle{
		{Artist: "Artist 1", Title: "Song 1"},
		{Artist: "Artist 1", Title: "Song 2"},
		{Artist: "Artist 1", Title: "Song 3"},
		{Artist: "Artist 1", Title: "Song 4"},
	}

	normalized, err := svc.Normalize(raw)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	compareTracks(normalized, expected, t)
}

func TestNormalize_YoutubeVideoMixtape(t *testing.T) {
	svc := logic.NewTrackNormalizer()

	raw := []logic.TrackProperties{
		{RawTitle: "Artist 1 - Song 1", ParentTitle: "Mixtape Name", Uploader: "Uploader"},
		{RawTitle: "Artist 2 - Song 2", ParentTitle: "Mixtape Name", Uploader: "Uploader"},
		{RawTitle: "Artist 3 - Song 3", ParentTitle: "Mixtape Name", Uploader: "Uploader"},
		{RawTitle: "Artist 2 - Song 4", ParentTitle: "Mixtape Name", Uploader: "Uploader"},
	}
	expected := []artistAndTitle{
		{Artist: "Artist 1", Title: "Song 1"},
		{Artist: "Artist 2", Title: "Song 2"},
		{Artist: "Artist 3", Title: "Song 3"},
		{Artist: "Artist 2", Title: "Song 4"},
	}

	normalized, err := svc.Normalize(raw)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	compareTracks(normalized, expected, t)
}

func TestNormalize_BandcampAlbumFromArtist(t *testing.T) {
	svc := logic.NewTrackNormalizer()

	raw := []logic.TrackProperties{
		{RawTitle: "Artist 1 - Song 1", ParentTitle: "Album Title", Artist: "Artist 1", Title: "Song 1", AlbumArtist: "Artist 1", Uploader: "Artist 1"},
		{RawTitle: "Artist 1 - Song 2", ParentTitle: "Album Title", Artist: "Artist 1", Title: "Song 2", AlbumArtist: "Artist 1", Uploader: "Artist 1"},
		{RawTitle: "Artist 1 - Song 3", ParentTitle: "Album Title", Artist: "Artist 1", Title: "Song 3", AlbumArtist: "Artist 1", Uploader: "Artist 1"},
		{RawTitle: "Artist 1 - Song 4", ParentTitle: "Album Title", Artist: "Artist 1", Title: "Song 4", AlbumArtist: "Artist 1", Uploader: "Artist 1"},
	}
	expected := []artistAndTitle{
		{Artist: "Artist 1", Title: "Song 1"},
		{Artist: "Artist 1", Title: "Song 2"},
		{Artist: "Artist 1", Title: "Song 3"},
		{Artist: "Artist 1", Title: "Song 4"},
	}

	normalized, err := svc.Normalize(raw)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	compareTracks(normalized, expected, t)
}

func TestNormalize_BandcampCompilationFromLabel(t *testing.T) {
	svc := logic.NewTrackNormalizer()

	raw := []logic.TrackProperties{
		{RawTitle: "Artist 1 - Artist 1 - Song 1", ParentTitle: "Album Title", Artist: "Artist 1", Title: "Artist 1 - Song 1", AlbumArtist: "Label Name", Uploader: "Artist 1"},
		{RawTitle: "Artist 2 - Artist 2 - Song 2", ParentTitle: "Album Title", Artist: "Artist 2", Title: "Artist 2 - Song 2", AlbumArtist: "Label Name", Uploader: "Artist 2"},
		{RawTitle: "Artist 3 - Artist 3 - Song 3", ParentTitle: "Album Title", Artist: "Artist 3", Title: "Artist 3 - Song 3", AlbumArtist: "Label Name", Uploader: "Artist 3"},
		{RawTitle: "Artist 1 - Artist 1 - Song 4", ParentTitle: "Album Title", Artist: "Artist 1", Title: "Artist 1 - Song 4", AlbumArtist: "Label Name", Uploader: "Artist 1"},
	}
	expected := []artistAndTitle{
		{Artist: "Artist 1", Title: "Song 1"},
		{Artist: "Artist 2", Title: "Song 2"},
		{Artist: "Artist 3", Title: "Song 3"},
		{Artist: "Artist 1", Title: "Song 4"},
	}

	normalized, err := svc.Normalize(raw)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	compareTracks(normalized, expected, t)
}

func TestNormalize_RemoveJunkPrefixForSingleTracks(t *testing.T) {
	svc := logic.NewTrackNormalizer()

	type inputAndOutput struct {
		input  logic.TrackProperties
		output artistAndTitle
	}

	samples := []inputAndOutput{
		{
			input:  logic.TrackProperties{RawTitle: "Artist 1 - 04 - Song 1"},
			output: artistAndTitle{Artist: "Artist 1", Title: "Song 1"},
		},
	}

	for _, sample := range samples {
		normalized, err := svc.Normalize([]logic.TrackProperties{sample.input})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		compareTracks(normalized, []artistAndTitle{sample.output}, t)
	}
}

func TestNormalize_RemoveJunkSuffixForSingleTracks(t *testing.T) {
	svc := logic.NewTrackNormalizer()

	type inputAndOutput struct {
		input  logic.TrackProperties
		output artistAndTitle
	}

	samples := []inputAndOutput{
		{
			input:  logic.TrackProperties{RawTitle: "Artist1「Song1」（Official Music Video）"},
			output: artistAndTitle{Title: "Artist1「Song1」"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Artist1「Song1」(Official Visualiser)"},
			output: artistAndTitle{Title: "Artist1「Song1」"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "\"Song 1\" by Artist 1 (official video)"},
			output: artistAndTitle{Title: "\"Song 1\" by Artist 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "“Song 1” (Official Audio)"},
			output: artistAndTitle{Title: "“Song 1”"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Artist1 Song1(Official Music Video)"},
			output: artistAndTitle{Title: "Artist1 Song1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Song 1 (OFFICIAL LYRIC VIDEO)"},
			output: artistAndTitle{Title: "Song 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Song 1 [OFFICIAL AUDIO]"},
			output: artistAndTitle{Title: "Song 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Song 1 (MV)"},
			output: artistAndTitle{Title: "Song 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "曲名(MV)"},
			output: artistAndTitle{Title: "曲名"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Artist 1 \"Song 1\" (Official Video)"},
			output: artistAndTitle{Title: "Artist 1 \"Song 1\""},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Song 1 (Official Music Video)"},
			output: artistAndTitle{Title: "Song 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "\"cold weather\" (Official Lyric Video)"},
			output: artistAndTitle{Title: "\"cold weather\""},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Artist 1-Song 1 (w/Lyrics)"},
			output: artistAndTitle{Title: "Artist 1-Song 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Song 1 (Audio)"},
			output: artistAndTitle{Title: "Song 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Song 1 (Lyrics)"},
			output: artistAndTitle{Title: "Song 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "\"Song 1\" (audio only)"},
			output: artistAndTitle{Title: "\"Song 1\""},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Artist 1 - Song 1 【Subbed】"},
			output: artistAndTitle{Artist: "Artist 1", Title: "Song 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "\"Song 1\" (Full Album Stream)"},
			output: artistAndTitle{Title: "\"Song 1\""},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Song 1 (HD)"},
			output: artistAndTitle{Title: "Song 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Song 1 (official)"},
			output: artistAndTitle{Title: "Song 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "Song 1 (Music video)"},
			output: artistAndTitle{Title: "Song 1"},
		},
		{
			input:  logic.TrackProperties{RawTitle: "\"Song 1\" (360º)"},
			output: artistAndTitle{Title: "\"Song 1\""},
		},
	}

	for _, sample := range samples {
		normalized, err := svc.Normalize([]logic.TrackProperties{sample.input})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		compareTracks(normalized, []artistAndTitle{sample.output}, t)
	}
}

func TestNormalize_KeepAllowedSuffixForSingleTracks(t *testing.T) {
	svc := logic.NewTrackNormalizer()

	type inputAndOutput struct {
		input  logic.TrackProperties
		output artistAndTitle
	}

	samples := []inputAndOutput{
		{ // non-matching parentheses
			input:  logic.TrackProperties{RawTitle: "Artist 1 - Song 1 [official audio)"},
			output: artistAndTitle{Artist: "Artist 1", Title: "Song 1 [official audio)"},
		},
		{ // keep "live" to differentiate from proper recordings
			input:  logic.TrackProperties{RawTitle: "Artist 1 - Song 1 (live)"},
			output: artistAndTitle{Artist: "Artist 1", Title: "Song 1 (live)"},
		},
		{ // not a suffix but an actual title
			input:  logic.TrackProperties{RawTitle: "Artist 3 - (official audio)"},
			output: artistAndTitle{Artist: "Artist 3", Title: "(official audio)"},
		},
	}

	for _, sample := range samples {
		normalized, err := svc.Normalize([]logic.TrackProperties{sample.input})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		compareTracks(normalized, []artistAndTitle{sample.output}, t)
	}
}

func compareTracks(actualTracks []logic.TrackProperties, expectedTracks []artistAndTitle, t *testing.T) {
	if len(actualTracks) != len(expectedTracks) {
		t.Fatalf("Expected %d tracks, but got %d tracks", len(expectedTracks), len(actualTracks))
	}

	for i, actualTrack := range actualTracks {
		expectedTrack := expectedTracks[i]

		if actualTrack.Artist != expectedTrack.Artist || actualTrack.Title != expectedTrack.Title {
			t.Errorf(
				"Bad artist/title: expected artist=`%s` title=`%s`, got artist=`%s`, title=`%s`",
				expectedTrack.Artist,
				expectedTrack.Title,
				actualTrack.Artist,
				actualTrack.Title,
			)
		}
	}
}
