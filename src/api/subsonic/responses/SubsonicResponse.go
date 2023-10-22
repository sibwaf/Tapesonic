package responses

import (
	"encoding/xml"

	"tapesonic/build"
)

const (
	STATUS_OK     = "ok"
	STATUS_FAILED = "failed"
)

const (
	ERROR_CODE_GENERIC           = 0
	ERROR_CODE_PARAMETER_MISSING = 10
	ERROR_CODE_UNAUTHENTICATED   = 40
)

type SubsonicResponse struct {
	XMLName xml.Name `json:"-" xml:"http://subsonic.org/restapi subsonic-response"`

	Status        string         `json:"status" xml:"status,attr"`
	Version       string         `json:"version" xml:"version,attr"`
	Type          string         `json:"type" xml:"type,attr"`
	ServerVersion string         `json:"serverVersion" xml:"serverVersion,attr"`
	OpenSubsonic  string         `json:"openSubsonic" xml:"openSubsonic,attr"`
	Error         *subsonicError `json:"error,omitempty"`
}

type subsonicError struct {
	XMLName xml.Name `json:"-" xml:"error"`

	Code    int    `json:"code" xml:"code,attr"`
	Message string `json:"message" xml:"message,attr"`
}

func NewOkSubsonicResponse() *SubsonicResponse {
	return &SubsonicResponse{
		Status:        STATUS_OK,
		Version:       "1.16.1",
		Type:          "Tapesonic",
		ServerVersion: build.TAPESONIC_VERSION,
		OpenSubsonic:  "true",
	}
}

func NewFailureSubsonicResponse(code int, message string) *SubsonicResponse {
	response := NewOkSubsonicResponse()
	response.Status = STATUS_FAILED
	response.Error = &subsonicError{
		Code:    code,
		Message: message,
	}

	return response
}
