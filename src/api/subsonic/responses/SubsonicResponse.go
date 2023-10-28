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
)

type SubsonicResponseWrapper struct {
	XMLName xml.Name `json:"-" xml:"subsonic-response"`
	SubsonicResponse `json:"subsonic-response"`
}

type SubsonicResponse struct {
	Status        string `json:"status" xml:"status,attr"`
	Version       string `json:"version" xml:"version,attr"`
	Type          string `json:"type" xml:"type,attr"`
	ServerVersion string `json:"serverVersion" xml:"serverVersion,attr"`
	OpenSubsonic  string `json:"openSubsonic" xml:"openSubsonic,attr"`

	Error *subsonicError `json:"error,omitempty" xml:"error"`

	Playlists *SubsonicPlaylists `json:"playlists,omitempty" xml:"playlists"`
	Playlist  *SubsonicPlaylist  `json:"playlist,omitempty" xml:"playlist"`
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
		OpenSubsonic:  "true",
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

func NewParameterMissingResponse(name string) *SubsonicResponse {
	return NewFailedResponse(ERROR_CODE_PARAMETER_MISSING, fmt.Sprintf("Required parameter `%s` is missing", name))
}

func NewNotAuthenticatedResponse() *SubsonicResponse {
	return NewFailedResponse(ERROR_CODE_NOT_AUTHENTICATED, "Wrong username/password or username/token")
}
