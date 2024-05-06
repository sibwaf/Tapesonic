package util

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"tapesonic/config"
	"tapesonic/http/subsonic/responses"
)

func Authenticated(handler http.HandlerFunc, config *config.TapesonicConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		username := query.Get(SUBSONIC_QUERY_USERNAME)
		if username == "" {
			response := responses.NewParameterMissingResponse(SUBSONIC_QUERY_USERNAME)
			LogDebug(r, response.Error.Message)
			writeResponse(w, r, response)
			return
		}

		password := query.Get(SUBSONIC_QUERY_PASSWORD)

		token := query.Get(SUBSONIC_QUERY_TOKEN)
		salt := query.Get(SUBSONIC_QUERY_SALT)

		var isAuthenticated bool
		if token != "" {
			if salt == "" {
				response := responses.NewParameterMissingResponse(SUBSONIC_QUERY_SALT)
				LogDebug(r, response.Error.Message)
				writeResponse(w, r, response)
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
			response := responses.NewParameterMissingResponse(SUBSONIC_QUERY_PASSWORD)
			LogDebug(r, response.Error.Message)
			writeResponse(w, r, response)
			return
		}

		if !isAuthenticated {
			LogWarning(r, "Request failed authentication")
			writeResponse(w, r, responses.NewNotAuthenticatedResponse())
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

	return GenerateToken(passwordConfig, salt) == token
}

func GenerateToken(password string, salt string) string {
	hash := md5.Sum([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}
