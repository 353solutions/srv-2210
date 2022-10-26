package client

import (
	"context"
	"fmt"
	"net/http"
)

type Client struct {
	BaseURL string
	client  http.Client
}

func New(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

func (c *Client) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status - %s", resp.Status)
	}

	return nil
}
