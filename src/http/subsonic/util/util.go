package util

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"tapesonic/http/subsonic/responses"
)

const (
	SUBSONIC_QUERY_USERNAME = "u"
	SUBSONIC_QUERY_PASSWORD = "p"
	SUBSONIC_QUERY_SALT     = "s"
	SUBSONIC_QUERY_TOKEN    = "t"
	SUBSONIC_QUERY_CLIENT   = "c"
	SUBSONIC_QUERY_FORMAT   = "f"
)

const (
	SUBSONIC_FORMAT_XML  = "xml"
	SUBSONIC_FORMAT_JSON = "json"
)

func writeResponse(w http.ResponseWriter, r *http.Request, response *responses.SubsonicResponse) {
	wrappedResponse := responses.SubsonicResponseWrapper{
		SubsonicResponse: *response,
	}

	switch format := getFormat(r); format {
	case SUBSONIC_FORMAT_JSON:
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wrappedResponse)
	case SUBSONIC_FORMAT_XML, "":
		w.Header().Add("Content-Type", "application/xml")
		xml.NewEncoder(w).Encode(wrappedResponse)
	default:
		LogError(r, "Unsupported format", "format", format)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func GetClientName(r *http.Request) string {
	return r.URL.Query().Get(SUBSONIC_QUERY_CLIENT)
}

func getFormat(r *http.Request) string {
	return r.URL.Query().Get(SUBSONIC_QUERY_FORMAT)
}
