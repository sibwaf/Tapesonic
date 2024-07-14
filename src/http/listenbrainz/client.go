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
