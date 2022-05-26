package geocloud

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

func NewClient(rawBaseURL, apiKey string, opts ...ClientOpt) (*Client, error) {
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{baseURL, http.DefaultClient}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.httpClient.Jar, err = cookiejar.New(&cookiejar.Options{}); err != nil {
		return nil, err
	}
	c.httpClient.Jar.SetCookies(baseURL, []*http.Cookie{
		{
			Name:  APIKeyCookie,
			Value: apiKey,
		},
	})

	return c, nil
}

func (c *Client) get(url *url.URL, i interface{}) error {
	res, err := c.httpClient.Get(url.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 || res.StatusCode < 200 {
		return fmt.Errorf("http %d", res.StatusCode)
	}

	return json.NewDecoder(res.Body).Decode(i)
}

func (c *Client) post(url *url.URL, r io.Reader, contentType string, i interface{}) error {
	res, err := c.httpClient.Post(url.String(), contentType, r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 || res.StatusCode < 200 {
		return fmt.Errorf("http %d", res.StatusCode)
	}

	return json.NewDecoder(res.Body).Decode(i)
}
