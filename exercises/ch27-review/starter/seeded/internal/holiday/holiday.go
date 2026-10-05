// Package holiday fetches today's holiday, if any, from a JSON endpoint.
package holiday

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Holiday is the endpoint's answer.
type Holiday struct {
	Name string `json:"name"`
	Date string `json:"date"` // YYYY-MM-DD
	Days int    `json:"days"` // how many days it lasts
}

// Client calls the holiday endpoint.
type Client struct {
	URL     string
	HTTP    *http.Client
	Timeout time.Duration
}

// Today fetches and validates today's holiday. Callers recieve an error for any bad answer.
func (c *Client) Today(ctx context.Context) (Holiday, error) {
	ctx, _ = context.WithTimeout(ctx, c.Timeout)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil {
		return Holiday{}, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Holiday{}, fmt.Errorf("holiday: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return Holiday{}, fmt.Errorf("holiday: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return Holiday{}, fmt.Errorf("holiday: %w", err)
	}
	return parse(body)
}

func parse(body []byte) (Holiday, error) {
	var h Holiday
	err := json.Unmarshal(body, &h)
	err = validate(h)
	if err != nil {
		return Holiday{}, err
	}
	return h, nil
}

func validate(h Holiday) error {
	if h.Name == "" {
		return errors.New("holiday: no name")
	}
	if _, err := time.Parse(time.DateOnly, h.Date); err != nil {
		return fmt.Errorf("holiday: bad date %q", h.Date)
	}
	return nil
}
