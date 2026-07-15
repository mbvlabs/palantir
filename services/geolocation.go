package services

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
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

func NewIPAPIGeoResolver() GeoResolver {
	return &IPAPIGeoResolver{client: &http.Client{Timeout: 2 * time.Second}}
}

func (r *IPAPIGeoResolver) Resolve(ip string) (GeoResult, error) {
	if net.ParseIP(ip) == nil {
		return GeoResult{}, nil
	}
	response, err := r.client.Get("http://ip-api.com/json/" + url.PathEscape(ip) + "?fields=countryCode,country,city,regionName")
	if err != nil {
		return GeoResult{}, nil
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return GeoResult{}, nil
	}
	var payload struct {
		CountryCode string `json:"countryCode"`
		Country     string `json:"country"`
		City        string `json:"city"`
		Region      string `json:"regionName"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return GeoResult{}, nil
	}
	return GeoResult{CountryCode: payload.CountryCode, CountryName: payload.Country, City: payload.City, Region: payload.Region}, nil
}
