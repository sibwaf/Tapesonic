package handlers

import (
	"encoding/json"
	"net/http"
	"tapesonic/logic"
	"time"
)

type LastFmSessionRs struct {
	Username  string
	UpdatedAt time.Time
}

type CreateLastFmSessionRq struct {
	Token string
}

type settingsLastFmAuthHandler struct {
	lastfm *logic.LastFmService
}

func NewSettingsLastFmAuthHandler(
	lastfm *logic.LastFmService,
) *settingsLastFmAuthHandler {
	return &settingsLastFmAuthHandler{
		lastfm: lastfm,
	}
}

func (h *settingsLastFmAuthHandler) Methods() []string {
	return []string{http.MethodGet, http.MethodPost}
}

func (h *settingsLastFmAuthHandler) Handle(r *http.Request) (any, error) {
	switch r.Method {
	case http.MethodGet:
		session, err := h.lastfm.GetCurrentSession()
		if err != nil {
			return nil, err
		}

		if session == nil {
			return nil, nil
		}

		return LastFmSessionRs{
			Username:  session.Username,
			UpdatedAt: session.UpdatedAt,
		}, nil
	case http.MethodPost:
		request := CreateLastFmSessionRq{}
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			return nil, err
		}

		session, err := h.lastfm.CreateSession(request.Token)
		if err != nil {
			return nil, err
		}

		return LastFmSessionRs{
			Username:  session.Username,
			UpdatedAt: session.UpdatedAt,
		}, nil
	default:
		return nil, http.ErrNotSupported
	}
}
