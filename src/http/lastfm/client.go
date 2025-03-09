package lastfm

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

type LastFmClient struct {
	apiKey    string
	apiSecret string
	baseUrl   string
}

func NewLastFmClient(apiKey string, apiSecret string) *LastFmClient {
	return &LastFmClient{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		baseUrl:   "https://ws.audioscrobbler.com/2.0/",
	}
}

func (c *LastFmClient) GetApiKey() string {
	return c.apiKey
}

func (c *LastFmClient) AuthGetToken() (TokenWrapper, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl, nil)
	if err != nil {
		return TokenWrapper{}, err
	}

	query := req.URL.Query()
	query.Add("method", "auth.getToken")
	query.Add("api_key", c.apiKey)
	query.Add("api_sig", c.createSignature(query))
	query.Add("format", "json")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return TokenWrapper{}, err
	}

	defer res.Body.Close()

	if err = extractError(res); err != nil {
		return TokenWrapper{}, err
	}

	var result TokenWrapper
	return result, json.NewDecoder(res.Body).Decode(&result)
}

func (c *LastFmClient) AuthGetSession(token string) (SessionWrapper, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseUrl, nil)
	if err != nil {
		return SessionWrapper{}, err
	}

	query := req.URL.Query()
	query.Add("method", "auth.getSession")
	query.Add("token", token)
	query.Add("api_key", c.apiKey)
	query.Add("api_sig", c.createSignature(query))
	query.Add("format", "json")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return SessionWrapper{}, err
	}

	defer res.Body.Close()

	if err = extractError(res); err != nil {
		return SessionWrapper{}, err
	}

	var result SessionWrapper
	return result, json.NewDecoder(res.Body).Decode(&result)
}

func (c *LastFmClient) UpdateNowPlaying(sessionKey string, request UpdateNowPlayingRq) error {
	params := url.Values{}
	params.Add("method", "track.updateNowPlaying")
	params.Add("artist", request.Artist)
	params.Add("track", request.Track)
	if request.Album != "" {
		params.Add("album", request.Album)
	}
	params.Add("api_key", c.apiKey)
	params.Add("sk", sessionKey)
	params.Add("api_sig", c.createSignature(params))
	params.Add("format", "json")

	res, err := http.DefaultClient.PostForm(c.baseUrl, params)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return extractError(res)
}

func (c *LastFmClient) Scrobble(sessionKey string, request ScrobbleRq) error {
	params := url.Values{}
	params.Add("method", "track.scrobble")
	params.Add("timestamp", fmt.Sprint(request.Timestamp))
	params.Add("artist", request.Artist)
	params.Add("track", request.Track)
	if request.Album != "" {
		params.Add("album", request.Album)
	}
	params.Add("api_key", c.apiKey)
	params.Add("sk", sessionKey)
	params.Add("api_sig", c.createSignature(params))
	params.Add("format", "json")

	res, err := http.DefaultClient.PostForm(c.baseUrl, params)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return extractError(res)
}

func (c *LastFmClient) createSignature(params url.Values) string {
	sortedKeys := make([]string, len(params))
	for key := range params {
		sortedKeys = append(sortedKeys, key)
	}
	slices.Sort(sortedKeys)

	signature := strings.Builder{}
	for _, key := range sortedKeys {
		values := params[key]
		if len(values) > 0 {
			signature.WriteString(key)
			signature.WriteString(values[0])
		}
	}

	signature.WriteString(c.apiSecret)

	md5 := md5.New()
	md5.Write([]byte(signature.String()))
	return hex.EncodeToString(md5.Sum(nil))
}

func extractError(res *http.Response) error {
	if res.StatusCode == http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(res.Body)
	if err == nil {
		return fmt.Errorf("last.fm http %d: %s", res.StatusCode, string(body))
	} else {
		return fmt.Errorf("last.fm http %d", res.StatusCode)
	}
}

func (c *LastFmClient) GetLibraryPlaylist(username string, page int) (PlaylistWrapper, error) {
	return c.getStationPlaylist(username, "library", page)
}

func (c *LastFmClient) GetMixPlaylist(username string, page int) (PlaylistWrapper, error) {
	return c.getStationPlaylist(username, "mix", page)
}

func (c *LastFmClient) GetRecommendedPlaylist(username string, page int) (PlaylistWrapper, error) {
	return c.getStationPlaylist(username, "recommended", page)
}

func (c *LastFmClient) getStationPlaylist(username string, station string, page int) (PlaylistWrapper, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://www.last.fm/player/station/user/%s/%s", username, station), nil)
	if err != nil {
		return PlaylistWrapper{}, err
	}

	query := req.URL.Query()
	query.Add("page", fmt.Sprint(page))
	query.Add("ajax", "1")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return PlaylistWrapper{}, err
	}

	defer res.Body.Close()

	var result PlaylistWrapper
	return result, json.NewDecoder(res.Body).Decode(&result)
}
