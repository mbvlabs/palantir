package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GeoResult struct {
	CountryCode string
	CountryName string
	City        string
	Region      string
}

type GeoResolver interface {
	Resolve(ip string) (GeoResult, error)
}

type IPAPIGeoResolver struct {
	client *http.Client
}

func NewIPAPIGeoResolver() *IPAPIGeoResolver {
	return &IPAPIGeoResolver{
		client: &http.Client{Timeout: 2 * time.Second},
	}
}

type ipAPIResponse struct {
	CountryCode string `json:"countryCode"`
	Country     string `json:"country"`
	City        string `json:"city"`
	RegionName  string `json:"regionName"`
}

func (r *IPAPIGeoResolver) Resolve(ip string) (GeoResult, error) {
	resp, err := r.client.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=countryCode,country,city,regionName", ip))
	if err != nil {
		return GeoResult{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GeoResult{}, nil
	}

	var data ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return GeoResult{}, nil
	}

	return GeoResult{
		CountryCode: data.CountryCode,
		CountryName: data.Country,
		City:        data.City,
		Region:      data.RegionName,
	}, nil
}
