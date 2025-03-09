package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"tapesonic/http/subsonic/responses"
	subsonicUtil "tapesonic/http/subsonic/util"
	commonUtil "tapesonic/util"
	"time"
)

type SubsonicClient struct {
	baseUrl  string
	username string
	password string
}

func NewSubsonicClient(
	baseUrl string,
	username string,
	password string,
) *SubsonicClient {
	return &SubsonicClient{
		baseUrl:  baseUrl,
		username: username,
		password: password,
	}
}

func (c *SubsonicClient) Search3(
	query string,
	artistCount int,
	artistOffset int,
	albumCount int,
	albumOffset int,
	songCount int,
	songOffset int,
) (*responses.SearchResult3, error) {
	res, err := c.doParsedQuery(
		"/rest/search3",
		map[string]string{
			"query":        query,
			"artistCount":  fmt.Sprint(artistCount),
			"artistOffset": fmt.Sprint(artistOffset),
			"albumCount":   fmt.Sprint(albumCount),
			"albumOffset":  fmt.Sprint(albumOffset),
			"songCount":    fmt.Sprint(songCount),
			"songOffset":   fmt.Sprint(songOffset),
		},
	)
	if err != nil {
		return nil, err
	}

	return res.SearchResult3, nil
}

func (c *SubsonicClient) GetSong(id string) (*responses.SubsonicChild, error) {
	res, err := c.doParsedQuery("/rest/getSong", map[string]string{"id": id})
	if err != nil {
		return nil, err
	}

	return res.Song, nil
}

func (c *SubsonicClient) GetRandomSongs(size int, genre string, fromYear *int, toYear *int) (*responses.RandomSongs, error) {
	params := map[string]string{"size": fmt.Sprint(size)}
	if genre != "" {
		params["genre"] = genre
	}
	if fromYear != nil {
		params["fromYear"] = fmt.Sprint(*fromYear)
	}
	if toYear != nil {
		params["toYear"] = fmt.Sprint(*toYear)
	}

	res, err := c.doParsedQuery("/rest/getRandomSongs", params)
	if err != nil {
		return nil, err
	}

	return res.RandomSongs, nil
}

func (c *SubsonicClient) GetAlbum(id string) (*responses.AlbumId3, error) {
	res, err := c.doParsedQuery("/rest/getAlbum", map[string]string{"id": id})
	if err != nil {
		return nil, err
	}

	return res.Album, nil
}

func (c *SubsonicClient) GetAlbumList2(
	type_ string,
	size int,
	offset int,
	fromYear *int,
	toYear *int,
) (*responses.AlbumList2, error) {
	params := map[string]string{
		"type":   type_,
		"size":   fmt.Sprint(size),
		"offset": fmt.Sprint(offset),
	}
	if fromYear != nil {
		params["fromYear"] = fmt.Sprint(*fromYear)
	}
	if toYear != nil {
		params["toYear"] = fmt.Sprint(*toYear)
	}

	res, err := c.doParsedQuery("/rest/getAlbumList2", params)
	if err != nil {
		return nil, err
	}

	return res.AlbumList2, nil
}

func (c *SubsonicClient) GetPlaylist(id string) (*responses.SubsonicPlaylist, error) {
	res, err := c.doParsedQuery("/rest/getPlaylist", map[string]string{"id": id})
	if err != nil {
		return nil, err
	}

	return res.Playlist, nil
}

func (c *SubsonicClient) GetPlaylists() (*responses.SubsonicPlaylists, error) {
	res, err := c.doParsedQuery("/rest/getPlaylists", map[string]string{})
	if err != nil {
		return nil, err
	}

	return res.Playlists, nil
}

func (c *SubsonicClient) GetArtist(id string) (*responses.Artist, error) {
	res, err := c.doParsedQuery("/rest/getArtist", map[string]string{"id": id})
	if err != nil {
		return nil, err
	}

	return res.Artist, nil
}

func (c *SubsonicClient) Scrobble(id string, time_ time.Time, submission bool) error {
	_, err := c.doParsedQuery("/rest/scrobble", map[string]string{
		"id":         id,
		"time":       fmt.Sprint(time_.UnixMilli()),
		"submission": fmt.Sprint(submission),
	})
	return err
}

func (c *SubsonicClient) GetCoverArt(id string) (mime string, reader io.ReadCloser, err error) {
	return c.doRawQuery("/rest/getCoverArt", map[string]string{"id": id})
}

func (c *SubsonicClient) Stream(id string) (mime string, reader io.ReadCloser, err error) {
	return c.doRawQuery("/rest/stream", map[string]string{"id": id})
}

func (c *SubsonicClient) GetLicense() (*responses.License, error) {
	res, err := c.doParsedQuery("/rest/getLicense", map[string]string{})
	if err != nil {
		return nil, err
	}

	return res.License, nil
}

func (c *SubsonicClient) doParsedQuery(path string, params map[string]string) (*responses.SubsonicResponse, error) {
	_, body, err := c.doRawQuery(path, params)
	if err != nil {
		return nil, err
	}

	defer body.Close()

	var response responses.SubsonicResponseWrapper
	err = json.NewDecoder(body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		err = fmt.Errorf("subsonic error %d: %s", response.Error.Code, response.Error.Message)
		return nil, err
	}

	return &response.SubsonicResponse, nil
}

func (c *SubsonicClient) doRawQuery(path string, params map[string]string) (string, io.ReadCloser, error) {
	req, err := http.NewRequest("GET", c.baseUrl+path, nil)
	if err != nil {
		return "", nil, err
	}

	query := prepareQueryParams(*req.URL, c.username, c.password)
	for paramName, paramValue := range params {
		query.Add(paramName, paramValue)
	}
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}

	return res.Header.Get("Content-Type"), res.Body, nil
}

func prepareQueryParams(url url.URL, username string, password string) url.Values {
	salt := commonUtil.GenerateRandomString(8)
	token := subsonicUtil.GenerateToken(password, salt)

	query := url.Query()
	query.Add("f", "json")
	query.Add("v", "1.16.1")
	query.Add("c", "tapesonic")
	query.Add("u", username)
	query.Add("t", token)
	query.Add("s", salt)
	return query
}
