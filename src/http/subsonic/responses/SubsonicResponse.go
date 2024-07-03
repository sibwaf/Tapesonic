package responses

import (
	"encoding/xml"
	"fmt"
	"tapesonic/build"
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

type SubsonicResponseWrapper struct {
	XMLName          xml.Name `json:"-" xml:"subsonic-response"`
	SubsonicResponse `json:"subsonic-response"`
}

type SubsonicResponse struct {
	Status        string `json:"status" xml:"status,attr"`
	Version       string `json:"version" xml:"version,attr"`
	Type          string `json:"type" xml:"type,attr"`
	ServerVersion string `json:"serverVersion" xml:"serverVersion,attr"`
	OpenSubsonic  bool   `json:"openSubsonic" xml:"openSubsonic,attr"`

	Error *subsonicError `json:"error,omitempty" xml:"error"`

	AlbumList2            *AlbumList2            `json:"albumList2,omitempty" xml:"albumList2"`
	Album                 *AlbumId3              `json:"album,omitempty" xml:"album"`
	Artists               *Artists               `json:"artists,omitempty" xml:"artists"`
	Artist                *Artist                `json:"artist,omitempty" xml:"artist"`
	Genres                *Genres                `json:"genres,omitempty" xml:"genres"`
	InternetRadioStations *InternetRadioStations `json:"internetRadioStations,omitempty" xml:"internetRadioStations"`
	MusicFolders          *MusicFolders          `json:"musicFolders,omitempty" xml:"musicFolders"`
	NewestPodcasts        *NewestPodcasts        `json:"newestPodcasts,omitempty" xml:"newestPodcasts"`
	Playlists             *SubsonicPlaylists     `json:"playlists,omitempty" xml:"playlists"`
	Playlist              *SubsonicPlaylist      `json:"playlist,omitempty" xml:"playlist"`
	Podcasts              *Podcasts              `json:"podcasts,omitempty" xml:"podcasts"`
	RandomSongs           *RandomSongs           `json:"randomSongs,omitempty" xml:"randomSongs"`
	ScanStatus            *ScanStatus            `json:"scanStatus,omitempty" xml:"scanStatus"`
	SearchResult3         *SearchResult3         `json:"searchResult3,omitempty" xml:"searchResult3"`
	Song                  *SubsonicChild         `json:"song,omitempty" xml:"song"`
	Starred2              *Starred2              `json:"starred2,omitempty" xml:"starred2"`
}

type subsonicError struct {
	Code    int    `json:"code" xml:"code,attr"`
	Message string `json:"message" xml:"message,attr"`
}

func NewOkResponse() *SubsonicResponse {
	return &SubsonicResponse{
		Status:        STATUS_OK,
		Version:       "1.16.1",
		Type:          "Tapesonic",
		ServerVersion: build.TAPESONIC_VERSION,
		OpenSubsonic:  true,
	}
}

func NewFailedResponse(code int, message string) *SubsonicResponse {
	response := NewOkResponse()
	response.Status = STATUS_FAILED
	response.Error = &subsonicError{
		Code:    code,
		Message: message,
	}

	return response
}

func NewServerErrorResponse(message string) *SubsonicResponse {
	return NewFailedResponse(ERROR_CODE_GENERIC, message)
}

func NewParameterMissingResponse(name string) *SubsonicResponse {
	return NewFailedResponse(ERROR_CODE_PARAMETER_MISSING, fmt.Sprintf("Required parameter `%s` is missing", name))
}

func NewNotAuthenticatedResponse() *SubsonicResponse {
	return NewFailedResponse(ERROR_CODE_NOT_AUTHENTICATED, "Wrong username/password or username/token")
}

func NewNotFoundResponse(what string) *SubsonicResponse {
	return NewFailedResponse(ERROR_CODE_NOT_FOUND, fmt.Sprintf("Not found: %s", what))
}
