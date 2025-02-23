package handlers

import (
	"net/http"
	"tapesonic/logic"
)

type LastFmAuthLinkRs struct {
	Url   string
	Token string
}

type settingsLastFmCreateAuthLinkHandler struct {
	lastfm *logic.LastFmService
}

func NewSettingsLastFmCreateAuthLinkHandler(
	lastfm *logic.LastFmService,
) *settingsLastFmCreateAuthLinkHandler {
	return &settingsLastFmCreateAuthLinkHandler{
		lastfm: lastfm,
	}
}

func (h *settingsLastFmCreateAuthLinkHandler) Methods() []string {
	return []string{http.MethodPost}
}

func (h *settingsLastFmCreateAuthLinkHandler) Handle(r *http.Request) (any, error) {
	switch r.Method {
	case http.MethodPost:
		link, err := h.lastfm.CreateAuthLink()
		if err != nil {
			return nil, err
		}

		return LastFmAuthLinkRs{
			Url:   link.Url,
			Token: link.Token,
		}, nil
	default:
		return nil, http.ErrNotSupported
	}
}
