package subsonic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SubsonicClient struct {
	baseUrl string
}

func NewSubsonicClient(
	baseUrl string,
) *SubsonicClient {
	return &SubsonicClient{
		baseUrl: baseUrl,
	}
}

func (c *SubsonicClient) Ping(auth AuthMethod) (Response, error) {
	return c.doParsedQuery(auth, "/rest/ping", map[string]string{})
}

func (c *SubsonicClient) Search3(
	auth AuthMethod,
	query string,
	artistCount int,
	artistOffset int,
	albumCount int,
	albumOffset int,
	songCount int,
	songOffset int,
) (SearchResult3, error) {
	res, err := c.doParsedQuery(
		auth,
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
		return SearchResult3{}, err
	}

	if res.SearchResult3 != nil {
		return *res.SearchResult3, nil
	} else {
		return SearchResult3{}, ErrInvalidResponse
	}
}

func (c *SubsonicClient) Scrobble(auth AuthMethod, id string, time_ time.Time, submission bool) error {
	_, err := c.doParsedQuery(auth, "/rest/scrobble", map[string]string{
		"id":         id,
		"time":       fmt.Sprint(time_.UnixMilli()),
		"submission": fmt.Sprint(submission),
	})
	return err
}

func (c *SubsonicClient) GetCoverArt(ctx context.Context, auth AuthMethod, id string, headers http.Header) (*http.Response, error) {
	return c.doRawQuery(ctx, auth, "/rest/getCoverArt", map[string]string{"id": id}, headers)
}

func (c *SubsonicClient) Stream(ctx context.Context, auth AuthMethod, id string, headers http.Header) (*http.Response, error) {
	return c.doRawQuery(ctx, auth, "/rest/stream", map[string]string{"id": id}, headers)
}

func (c *SubsonicClient) doParsedQuery(auth AuthMethod, path string, params map[string]string) (Response, error) {
	res, err := c.doRawQuery(context.Background(), auth, path, params, http.Header{})
	if err != nil {
		return Response{}, ErrUnreachable
	}

	defer res.Body.Close()

	var response ResponseWrapper
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return Response{}, ErrInvalidResponse
	}

	if response.SubsonicResponse.Status != STATUS_OK {
		errorRs := response.SubsonicResponse.Error
		if errorRs == nil {
			return response.SubsonicResponse, ErrInvalidResponse
		} else {
			return response.SubsonicResponse, NewSubsonicError(errorRs.Code, errorRs.Message)
		}
	}

	return response.SubsonicResponse, nil
}

func (c *SubsonicClient) doRawQuery(ctx context.Context, auth AuthMethod, path string, params map[string]string, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("f", "json")
	query.Add("v", "1.16.1")
	query.Add("c", "tapesonic")

	for paramName, paramValue := range params {
		query.Add(paramName, paramValue)
	}

	req.URL.RawQuery = query.Encode()

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	auth.ApplyTo(req)

	return http.DefaultClient.Do(req)
}
