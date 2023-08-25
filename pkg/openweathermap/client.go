package openweathermap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sushkevichd/day-guide-telegram-bot/pkg/domain"
)

const baseURL = "https://api.openweathermap.org/data/2.5/weather"

type client struct {
	apiKey string
	hc     *http.Client
}

func NewClient(apiKey string) *client {
	return &client{
		apiKey: apiKey,
		hc:     &http.Client{},
	}
}

func (c *client) FetchCurrent(ctx context.Context, location domain.Location) (*domain.Weather, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing base url: %v", err)
	}

	q := u.Query()
	q.Set("q", string(location))
	q.Set("appid", c.apiKey)
	q.Set("units", "metric")
	q.Set("lang", "ru")

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %v", err)
	}

	var res weatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decoding response body: %v", err)
	}

	return &domain.Weather{
		Location:       res.Name,
		Temp:           res.Main.Temp,
		TempFeel:       res.Main.FeelsLike,
		Pressure:       res.Main.Pressure,
		Humidity:       res.Main.Humidity,
		Weather:        res.Weather[0].Main,
		WeatherVerbose: res.Weather[0].Description,
		WindSpeed:      res.Wind.Speed,
		WindDirection:  convertWindDirection(res.Wind.Deg),
	}, nil
}

func convertWindDirection(d int) string {
	if d == 0 {
		return "-"
	}

	switch {
	case 24 <= d && d <= 68:
		return "северо-восточный"
	case 69 <= d && d <= 113:
		return "восточный"
	case 114 <= d && d <= 158:
		return "юго-восточный"
	case 159 <= d && d <= 203:
		return "южный"
	case 204 <= d && d <= 248:
		return "юго-западный"
	case 249 <= d && d <= 293:
		return "западный"
	case 294 <= d && d <= 338:
		return "северо-западный"
	default:
		return "северный"
	}
}

type weatherAPIResponse struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		Id          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Rain struct {
		H float64 `json:"1h"`
	} `json:"rain"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		Id      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}
