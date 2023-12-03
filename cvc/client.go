package cvc

import "net/http"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{},
	}
}

func (c *Client) do(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", "cvcmonitor/0.1 (https://github.com/saucesteals/cvcmonitor)")

	return c.http.Do(r)
}
