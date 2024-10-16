package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	cerberusHost = "https://api.edgelab.ch/cerberus"
)

// IssuerList  lists all issuers IDs by querying cerberus.
func IssuerList() ([]string, error) {
	size := 1000
	issuers := make([]string, 0, size)
	url := fmt.Sprintf("%s/v2/issuers?size=%d", cerberusHost, size)

	for {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create http request for url %s: %w", url, err)
		}
		req.Header.Set("x-internal-service", "validation")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not send the request: %w", err)
		}

		defer res.Body.Close()

		raw, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read response body: %w", err)
		}

		var output struct {
			Data     []string `json:"data"`
			NextPage *string  `json:"next"`
		}
		if err := json.Unmarshal(raw, &output); err != nil {
			return nil, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
		}

		issuers = append(issuers, output.Data...)
		if output.NextPage == nil {
			break
		}

		url = fmt.Sprintf("%s%s", cerberusHost, *output.NextPage)

		log.Infof("Fetched %d issuers", len(issuers))
	}

	return issuers, nil
}

type IssuerDescription struct {
	ID                       string    `json:"id"`
	Name                     string    `json:"name"`
	ShortName                string    `json:"shortName"`
	Country                  Country   `json:"country"`
	StateProvince            string    `json:"stateProvince"`
	Sector                   Sector    `json:"sector"`
	CountryOfRisk            Country   `json:"countryOfRisk"`
	Industry                 Industry  `json:"industry"`
	CountryOfIncorporation   Country   `json:"countryOfIncorporation"`
	PrimaryExchangeCountries []Country `json:"primaryExchangeCountries"`
	MarketValue              *float64  `json:"marketValue"`
	MarketCap                *string   `json:"marketCap"`
}

type IssuerDescriptionMarketCap struct {
	ID                       string     `json:"id"`
	Name                     string     `json:"name"`
	ShortName                string     `json:"shortName"`
	Country                  Country    `json:"country"`
	StateProvince            string     `json:"stateProvince"`
	Sector                   Sector     `json:"sector"`
	CountryOfRisk            Country    `json:"countryOfRisk"`
	Industry                 Industry   `json:"industry"`
	CountryOfIncorporation   Country    `json:"countryOfIncorporation"`
	PrimaryExchangeCountries []Country  `json:"primaryExchangeCountries"`
	MarketValue              *float64   `json:"marketValue"`
	MarketCap                *MarketCap `json:"marketCap"`
}

type MarketCap struct {
	ID string `json:"id"`
}

type Country struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Sector struct {
	Name string `json:"name"`
}
type Industry struct {
	Name   string `json:"name"`
	Sector Sector `json:"sector"`
}

type SIC struct {
	Code        string `json:"sicCode"`
	Description string `json:"sicDescription"`
}

func IssuerDescriptions(issuers []string) ([]IssuerDescription, error) {
	descriptions := make([]IssuerDescription, 0, len(issuers))

	log.Infof("Fetching %d issuers", len(issuers))

	count := 0

	for _, issuerID := range issuers {
		url := fmt.Sprintf("%s/issuers/%s?view=full", cerberusHost, issuerID)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create the request: %w", err)
		}
		req.Header.Set("x-internal-service", "validation")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not send the request: %w", err)
		}

		defer res.Body.Close()

		raw, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read response body: %w", err)
		}

		if res.StatusCode != http.StatusOK {
			log.Errorf("failed with issuer %s, status code %d, response %s", issuerID, res.StatusCode, string(raw))
			continue
		}

		var description IssuerDescription
		if err := json.Unmarshal(raw, &description); err != nil {
			var descriptionMarketCap IssuerDescriptionMarketCap
			if err := json.Unmarshal(raw, &descriptionMarketCap); err != nil {
				log.Errorf("failed with issuer %s : %v", issuerID, err)
			}

			description.ID = descriptionMarketCap.ID
			description.Name = descriptionMarketCap.Name
			description.ShortName = descriptionMarketCap.ShortName
			description.Country = descriptionMarketCap.Country
			description.StateProvince = descriptionMarketCap.StateProvince
			description.Sector = descriptionMarketCap.Sector
			description.CountryOfRisk = descriptionMarketCap.CountryOfRisk
			description.Industry = descriptionMarketCap.Industry
			description.CountryOfIncorporation = descriptionMarketCap.CountryOfIncorporation
			description.PrimaryExchangeCountries = descriptionMarketCap.PrimaryExchangeCountries
			description.MarketValue = descriptionMarketCap.MarketValue
			if descriptionMarketCap.MarketCap != nil {
				description.MarketCap = &descriptionMarketCap.MarketCap.ID
			}
		}

		descriptions = append(descriptions, description)

		count++
		if count%1000 == 0 {
			log.Infof("Fetched descriptions of %d issuers (%.2f%%)", count, float64(count)/float64(len(issuers))*100)
		}
	}

	return descriptions, nil
}
