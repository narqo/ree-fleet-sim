package vehicle

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type FleetStateClient struct {
	baseUrl string

	Client *http.Client
}

func NewFleetStateClient(baseUrl string) *FleetStateClient {
	// trim final slash from the path to simplify internal URL path management
	baseUrl = strings.TrimSuffix(baseUrl, "/")
	return &FleetStateClient{
		baseUrl: baseUrl,
		Client:  http.DefaultClient,
	}
}

func (c *FleetStateClient) UpdatePosition(ctx context.Context, vin VIN, lat, lon float64) error {
	v := url.Values{
		"lat": []string{strconv.FormatFloat(lat, 'f', -1, 64)},
		"lon": []string{strconv.FormatFloat(lon, 'f', -1, 64)},
	}
	surl := c.baseUrl + "/vehicle/" + string(vin)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, surl, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected response status %d", resp.StatusCode)
	}

	return nil
}
