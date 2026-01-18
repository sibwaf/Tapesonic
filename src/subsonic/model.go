package subsonic

import (
	"encoding/xml"
	"fmt"
	"tapesonic/build"
	"time"
)

const (
	QUERY_USERNAME = "u"
	QUERY_PASSWORD = "p"
	QUERY_SALT     = "s"
	QUERY_TOKEN    = "t"
	QUERY_CLIENT   = "c"
	QUERY_FORMAT   = "f"
)

const (
	FORMAT_XML  = "xml"
	FORMAT_JSON = "json"
)

const (
	STATUS_OK     = "ok"
	STATUS_FAILED = "failed"
)

const (
	ERROR_CODE_GENERIC           = 0
	ERROR_CODE_PARAMETER_MISSING = 10
	ERROR_CODE_NOT_AUTHENTICATED = 40
	ERROR_CODE_NOT_FOUND         = 70
)

const (
	ALBUM_LIST_RANDOM    = "random"
	ALBUM_LIST_NEWEST    = "newest"
	ALBUM_LIST_HIGHEST   = "highest"
	ALBUM_LIST_FREQUENT  = "frequent"
	ALBUM_LIST_RECENT    = "recent"
	ALBUM_LIST_BY_NAME   = "alphabeticalByName"
	ALBUM_LIST_BY_ARTIST = "alphabeticalByArtist"
	ALBUM_LIST_STARRED   = "starred"
	ALBUM_LIST_BY_YEAR   = "byYear"
	ALBUM_LIST_BY_GENRE  = "byGenre"
)

type ResponseWrapper struct {
	XMLName          xml.Name `json:"-" xml:"subsonic-response"`
	SubsonicResponse Response `json:"subsonic-response"`
}

type Response struct {
	Status        string `json:"status" xml:"status,attr"`
	Version       string `json:"version" xml:"version,attr"`
	Type          string `json:"type" xml:"type,attr"`
	ServerVersion string `json:"serverVersion" xml:"serverVersion,attr"`
	OpenSubsonic  bool   `json:"openSubsonic" xml:"openSubsonic,attr"`

	Error *Error `json:"error,omitempty" xml:"error"`

	AlbumList2            *AlbumList2            `json:"albumList2,omitempty" xml:"albumList2"`
	Album                 *AlbumId3              `json:"album,omitempty" xml:"album"`
	Artists               *Artists               `json:"artists,omitempty" xml:"artists"`
	Artist                *ArtistId3             `json:"artist,omitempty" xml:"artist"`
	Genres                *Genres                `json:"genres,omitempty" xml:"genres"`
	InternetRadioStations *InternetRadioStations `json:"internetRadioStations,omitempty" xml:"internetRadioStations"`
	License               *License               `json:"license,omitempty" xml:"license"`
	MusicFolders          *MusicFolders          `json:"musicFolders,omitempty" xml:"musicFolders"`
	Podcasts              *Podcasts              `json:"podcasts,omitempty" xml:"podcasts"`
	NewestPodcasts        *NewestPodcasts        `json:"newestPodcasts,omitempty" xml:"newestPodcasts"`
	Playlists             *Playlists             `json:"playlists,omitempty" xml:"playlists"`
	Playlist              *Playlist              `json:"playlist,omitempty" xml:"playlist"`
	RandomSongs           *RandomSongs           `json:"randomSongs,omitempty" xml:"randomSongs"`
	ScanStatus            *ScanStatus            `json:"scanStatus,omitempty" xml:"scanStatus"`
	SearchResult3         *SearchResult3         `json:"searchResult3,omitempty" xml:"searchResult3"`
	Song                  *Child                 `json:"song,omitempty" xml:"song"`
	Starred2              *Starred2              `json:"starred2,omitempty" xml:"starred2"`
}

type Error struct {
	Code    int    `json:"code" xml:"code,attr"`
	Message string `json:"message" xml:"message,attr"`
}

func NewOkResponse() *Response {
	return &Response{
		Status:        STATUS_OK,
		Version:       "1.16.1",
		Type:          "Tapesonic",
		ServerVersion: build.TAPESONIC_VERSION,
		OpenSubsonic:  true,
	}
}

func NewFailedResponse(code int, message string) *Response {
	response := NewOkResponse()
	response.Status = STATUS_FAILED
	response.Error = &Error{
		Code:    code,
		Message: message,
	}

	return response
}

func NewParameterMissingResponse(name string) *Response {
	return NewFailedResponse(ERROR_CODE_PARAMETER_MISSING, fmt.Sprintf("Required parameter `%s` is missing", name))
}

func NewNotFoundResponse(what string) *Response {
	return NewFailedResponse(ERROR_CODE_NOT_FOUND, fmt.Sprintf("Not found: %s", what))
}

type MusicFolder struct {
	Id   string `json:"id" xml:"id,attr"`
	Name string `json:"name" xml:"name,attr"`
}

type MusicFolders struct {
	MusicFolder []MusicFolder `json:"musicFolder" xml:"musicFolder"`
}

type Child struct {
	Id    string `json:"id" xml:"id,attr"`
	IsDir bool   `json:"isDir" xml:"isDir,attr"`

	Artist   string `json:"artist" xml:"artist,attr"`
	ArtistId string `json:"artistId" xml:"artistId,attr"`
	Title    string `json:"title" xml:"title,attr"`
	Album    string `json:"album" xml:"album,attr"`
	AlbumId  string `json:"albumId" xml:"albumId,attr"`
	Track    int    `json:"track" xml:"track,attr"`

	CoverArt string `json:"coverArt" xml:"coverArt,attr"`

	Duration  int `json:"duration" xml:"duration,attr"`
	PlayCount int `json:"playCount" xml:"playCount,attr"`

	Played *time.Time `json:"played" xml:"played,attr,omitempty"`
}

type AlbumId3 struct {
	Id string `json:"id" xml:"id,attr"`

	Name     string `json:"name" xml:"name,attr"`
	Artist   string `json:"artist" xml:"artist,attr"`
	ArtistId string `json:"artistId" xml:"artistId,attr"`

	CoverArt string `json:"coverArt" xml:"coverArt,attr"`

	SongCount int `json:"songCount" xml:"songCount,attr"`
	Duration  int `json:"duration" xml:"duration,attr"`
	PlayCount int `json:"playCount" xml:"playCount,attr"`

	Created time.Time  `json:"created" xml:"created,attr"`
	Starred *time.Time `json:"starred" xml:"starred,attr,omitempty"`
	Played  *time.Time `json:"played" xml:"played,attr,omitempty"`

	Year int `json:"year" xml:"year,attr"`

	ReleaseDate *ItemDate `json:"releaseDate" xml:"releaseDate"`

	Song []Child `json:"song,omitempty" xml:"song"`
}

type AlbumList2 struct {
	Album []AlbumId3 `json:"album" xml:"album"`
}

type Playlist struct {
	Id        string `json:"id" xml:"id,attr"`
	Name      string `json:"name" xml:"name,attr"`
	SongCount int    `json:"songCount" xml:"songCount,attr"`
	Duration  int    `json:"duration" xml:"duration,attr"`

	Created time.Time `json:"created" xml:"created,attr"`
	Changed time.Time `json:"changed" xml:"changed,attr"`

	CoverArt string `json:"coverArt" xml:"coverArt,attr"`
	Owner    string `json:"owner" xml:"owner,attr"`

	Entry []Child `json:"entry,omitempty" xml:"entry"`
}

type Playlists struct {
	Playlist []Playlist `json:"playlist" xml:"playlist"`
}

type Artists struct {
	IgnoredArticles string     `json:"ignoredArticles" xml:"ignoredArticles,attr"`
	Index           []IndexId3 `json:"index" xml:"index"`
}

type IndexId3 struct {
	Name   string      `json:"name" xml:"name,attr"`
	Artist []ArtistId3 `json:"artist" xml:"artist"`
}

type ArtistId3 struct {
	Id   string `json:"id" xml:"id,attr"`
	Name string `json:"name" xml:"name,attr"`

	CoverArt       string `json:"coverArt" xml:"coverArt,attr"`
	ArtistImageUrl string `json:"artistImageUrl" xml:"artistImageUrl,attr"`

	AlbumCount int `json:"albumCount" xml:"albumCount,attr"`

	Starred *time.Time `json:"starred" xml:"starred,attr,omitempty"`

	Album []AlbumId3 `json:"album,omitempty" xml:"album"`
}

type ItemDate struct {
	Year  int `json:"year" xml:"year,attr"`
	Month int `json:"month" xml:"month,attr"`
	Day   int `json:"day" xml:"day,attr"`
}

type SearchResult3 struct {
	Artist []ArtistId3 `json:"artist" xml:"artist"`
	Album  []AlbumId3  `json:"album" xml:"album"`
	Song   []Child     `json:"song" xml:"song"`
}

type License struct {
	Valid          bool       `json:"valid" xml:"valid,attr"`
	Email          string     `json:"email,omitempty" xml:"email,attr,omitempty"`
	LicenseExpires *time.Time `json:"licenseExpires" xml:"licenseExpires,attr,omitempty"`
	TrialExpires   *time.Time `json:"trialExpires" xml:"trialExpires,attr,omitempty"`
}

type RandomSongs struct {
	Song []Child `json:"song" xml:"song"`
}

type ScanStatus struct {
	Scanning bool `json:"scanning" xml:"scanning,attr"`
	Count    int  `json:"count" xml:"count,attr"`
}

type Genres struct {
	Genre []Genre `json:"genre" xml:"genre"`
}

type Genre struct {
	Value      string `json:"value" xml:"value,attr"`
	SongCount  int    `json:"songCount" xml:"songCount,attr"`
	AlbumCount int    `json:"albumCount" xml:"albumCount,attr"`
}

type InternetRadioStations struct {
	InternetRadioStation []InternetRadioStation `json:"internetRadioStation" xml:"internetRadioStation"`
}

type InternetRadioStation struct {
	Id          string `json:"id" xml:"id,attr"`
	Name        string `json:"name" xml:"name,attr"`
	StreamUrl   string `json:"streamUrl" xml:"streamUrl,attr"`
	HomePageUrl string `json:"homePageUrl,omitempty" xml:"homePageUrl,attr,omitempty"`
}

type PodcastChannel struct {
}

type PodcastEpisode struct {
}

type Podcasts struct {
	Channel []PodcastChannel `json:"channel" xml:"channel"`
}

type NewestPodcasts struct {
	Episode []PodcastEpisode `json:"episode" xml:"episode"`
}

type Starred2 struct {
	Artist []ArtistId3 `json:"artist" xml:"artist"`
	Album  []AlbumId3  `json:"album" xml:"album"`
	Song   []Child     `json:"song" xml:"song"`
}
