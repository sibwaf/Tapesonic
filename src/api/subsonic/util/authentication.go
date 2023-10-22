package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"tapesonic/api/subsonic/responses"
	"tapesonic/config"
)

func Authenticated(handler http.HandlerFunc, config *config.TapesonicConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		username := query.Get(SUBSONIC_QUERY_USERNAME)
		if username == "" {
			msg := fmt.Sprintf("Query parameter `%s` is missing", SUBSONIC_QUERY_USERNAME)
			LogInfo(r, msg)
			writeResponse(w, r, responses.NewFailureSubsonicResponse(responses.ERROR_CODE_PARAMETER_MISSING, msg))
			return
		}

		password := query.Get(SUBSONIC_QUERY_PASSWORD)

		token := query.Get(SUBSONIC_QUERY_TOKEN)
		salt := query.Get(SUBSONIC_QUERY_SALT)

		var isAuthenticated bool
		if token != "" {
			if salt == "" {
				msg := fmt.Sprintf("Query parameter `%s` is missing", SUBSONIC_QUERY_SALT)
				LogInfo(r, msg)
				writeResponse(w, r, responses.NewFailureSubsonicResponse(responses.ERROR_CODE_PARAMETER_MISSING, msg))
				return
			}

			isAuthenticated = authenticateByToken(
				username,
				token,
				salt,
				config.Username,
				config.Password,
			)
		} else if password != "" {
			isAuthenticated = authenticateByPassword(
				username,
				password,
				config.Username,
				config.Password,
			)
		} else {
			msg := fmt.Sprintf("Query parameters `%s` or `%s`+`%s` are missing", SUBSONIC_QUERY_PASSWORD, SUBSONIC_QUERY_TOKEN, SUBSONIC_QUERY_SALT)
			LogInfo(r, msg)
			writeResponse(w, r, responses.NewFailureSubsonicResponse(responses.ERROR_CODE_PARAMETER_MISSING, msg))
			return
		}

		if !isAuthenticated {
			LogWarning(r, "Request failed authentication")
			writeResponse(w, r, responses.NewFailureSubsonicResponse(responses.ERROR_CODE_UNAUTHENTICATED, "Wrong username/password or username/token"))
			return
		}

		handler(w, r)
	}
}

func authenticateByPassword(username string, password string, usernameConfig string, passwordConfig string) bool {
	if username != usernameConfig {
		return false
	}

	if strings.HasPrefix(password, "enc:") {
		return password == encodePassword(passwordConfig)
	} else {
		return password == passwordConfig
	}
}

func encodePassword(password string) string {
	parts := []string{}

	for _, char := range []byte(password) {
		parts = append(parts, fmt.Sprintf("%x", char))
	}

	return "enc:" + strings.Join(parts, "")
}

func authenticateByToken(username string, token string, salt string, usernameConfig string, passwordConfig string) bool {
	if username != usernameConfig {
		return false
	}

	expectedHash := md5.Sum([]byte(passwordConfig + salt))
	return hex.EncodeToString(expectedHash[:]) == token
}
