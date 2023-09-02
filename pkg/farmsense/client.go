package farmsense

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const baseURL = "http://api.farmsense.net/v1/moonphases"

type client struct {
	hc *http.Client
}

func NewClient() *client {
	return &client{
		hc: &http.Client{},
	}
}

func (c *client) FetchCurrent(ctx context.Context) (*domain.MoonPhase, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base url: %v", err)
	}

	q := u.Query()

	now := time.Now().UnixMilli() / 1000
	q.Set("d", strconv.Itoa(int(now)))

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %v", err)
	}

	var res []moonPhasesResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding response body: %v", err)
	}

	if res[0].Error != 0 {
		return nil, fmt.Errorf("API error: %s", res[0].ErrorMsg)
	}

	return &domain.MoonPhase{
		Age:             res[0].Index,
		Names:           res[0].Moon,
		Phase:           res[0].Phase,
		DistanceToEarth: res[0].Distance,
		IlluminationPrc: int(res[0].Illumination * 100),
		DistanceToSun:   res[0].DistanceToSun,
	}, nil
}

type moonPhasesResponse struct {
	Error              int      `json:"Error"`
	ErrorMsg           string   `json:"ErrorMsg"`
	TargetDate         string   `json:"TargetDate"`
	Moon               []string `json:"Moon"`
	Index              int      `json:"Index"`
	Age                float64  `json:"Age"`
	Phase              string   `json:"Phase"`
	Distance           float64  `json:"Distance"`
	Illumination       float64  `json:"Illumination"`
	AngularDiameter    float64  `json:"AngularDiameter"`
	DistanceToSun      float64  `json:"DistanceToSun"`
	SunAngularDiameter float64  `json:"SunAngularDiameter"`
}
