package listenbrainz

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"tapesonic/config"
)

type ListenBrainzClient struct {
	token   string
	baseUrl string
}

func NewListenBrainzClient(token string) *ListenBrainzClient {
	return &ListenBrainzClient{
		token:   token,
		baseUrl: "https://api.listenbrainz.org",
	}
}

func (c *ListenBrainzClient) ValidateToken() (*ValidateTokenResponse, error) {
	slog.Log(context.Background(), config.LevelTrace, "Validating ListenBrainz token")

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/1/validate-token", c.baseUrl), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var result ValidateTokenResponse
	return &result, json.NewDecoder(res.Body).Decode(&result)
}

func (c *ListenBrainzClient) SubmitListens(request SubmitListensRequest) error {
	slog.Log(context.Background(), config.LevelTrace, fmt.Sprintf("Submitting listens to ListenBrainz: %+v", request))

	bodyContent, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/1/submit-listens", c.baseUrl), bytes.NewReader(bodyContent))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errorText := fmt.Sprintf("http %s", res.Status)

		bodyBytes, err := io.ReadAll(res.Body)
		if err == nil {
			errorText = fmt.Sprintf("%s: %s", errorText, string(bodyBytes))
		}

		return errors.New(errorText)
	}

	return nil
}

func (c *ListenBrainzClient) GetPlaylistsCreatedFor(username string, count int, offset int) (*PlaylistsResponse, error) {
	slog.Log(
		context.Background(),
		config.LevelTrace,
		fmt.Sprintf("Retrieving created-for playlists from ListenBrainz: username=%s, count=%d, offset=%d", username, count, offset),
	)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/1/user/%s/playlists/createdfor", c.baseUrl, username), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var result PlaylistsResponse
	return &result, json.NewDecoder(res.Body).Decode(&result)
}

func (c *ListenBrainzClient) GetPlaylist(id string) (*PlaylistResponse, error) {
	slog.Log(
		context.Background(),
		config.LevelTrace,
		fmt.Sprintf("Retrieving playlist from ListenBrainz: id=%s", id),
	)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/1/playlist/%s", c.baseUrl, id), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("playlist with id `%s` was not found", id)
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http code %d when requesting %s", res.StatusCode, req.URL)
	}

	var result PlaylistResponseWrapper
	return &result.Playlist, json.NewDecoder(res.Body).Decode(&result)
}
