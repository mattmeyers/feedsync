package pocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	addLinkURL = "https://getpocket.com/v3/add"
)

// Client handles all interaction with the V3 Pocket API.
type Client struct {
	client *http.Client

	consumerKey string
	accessToken string
}

// NewClient produces a new Client with the provided credentials. If either of the input
// parameters is missing, then an error will be returned.
func NewClient(consumerKey, accessToken string) *Client {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		client:      client,
		consumerKey: consumerKey,
		accessToken: accessToken,
	}
}

func (c *Client) newRequest(method string, url string, body interface{}) (*http.Request, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Accept", "application/json")

	return req, nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type addLinkRequest struct {
	URL         string `json:"url"`
	ConsumerKey string `json:"consumer_key"`
	AccessToken string `json:"access_token"`
}

// AddLink sends a link to the Pocket API to be added to the user's list.
func (c *Client) AddLink(link string) error {
	body := addLinkRequest{
		URL:         link,
		ConsumerKey: c.consumerKey,
		AccessToken: c.accessToken,
	}

	req, err := c.newRequest(http.MethodPost, addLinkURL, body)
	if err != nil {
		return err
	}

	res, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"add link request failed with %d status code: %s",
			res.StatusCode,
			res.Header.Get("X-Error"),
		)
	}

	return nil
}
