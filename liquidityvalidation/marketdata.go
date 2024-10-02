package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type liquidityOutput struct {
	id         string
	horizon    int
	marketCap  *float64
	pocHorizon *int
}

func requestMarketdata(assetIDs []string) ([]liquidityOutput, error) {
	maxNumber := len(assetIDs)
	liquidityHorizons := make([]liquidityOutput, 0, maxNumber)
	issuersMarketCap := make(map[string]*float64)
	counter := 0
	for _, id := range assetIDs {
		horizon, issuer, err := liquidityHorizon(id)
		if err != nil {
			return nil, fmt.Errorf("could not check liquidity horizon: %w", err)
		}

		output := liquidityOutput{
			id:      id,
			horizon: horizon,
		}

		if marketCap, ok := issuersMarketCap[issuer]; ok {
			if marketCap != nil {
				output.marketCap = marketCap

				pocHorizon := marketCapToHorizon(marketCap)
				output.pocHorizon = &pocHorizon
			}
		} else {
			marketCap, err := issuerMarketCap(issuer)
			if err != nil {
				return nil, fmt.Errorf("could not check issuer market cap: %w", err)
			}

			issuersMarketCap[issuer] = marketCap
			if marketCap != nil {
				output.marketCap = marketCap

				pocHorizon := marketCapToHorizon(marketCap)
				output.pocHorizon = &pocHorizon
			}
		}

		liquidityHorizons = append(liquidityHorizons, output)

		counter++
		if counter%100 == 0 {
			log.Printf("Processed %d/%d assets (%f%%)", counter, maxNumber, float64(counter)/float64(maxNumber)*100)
		}
	}

	return liquidityHorizons, nil
}

func liquidityHorizon(id string) (int, string, error) {
	url := fmt.Sprintf("http://marketdata.service.consul/assets/id/%s?view=full", id)
	if environment == "PROD" {
		url = fmt.Sprintf("https://api.edgelab.ch/cerberux/assets/id/%s?view=full", id)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return 0, "", fmt.Errorf("could not create the request: %w", err)
	}
	req.Header.Set("x-internal-service", "validation")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if environment == "PROD" {
		req.Header.Set("Authorization", "Bearer "+tokenPROD)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("could not send the request: %w", err)
	}

	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, "", fmt.Errorf("could not read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		log.Infof("failed with asset %s, status code %d, response %s", id, res.StatusCode, string(raw))
		return 0, "", nil
	}

	var response struct {
		Horizon int `json:"liquidityHorizon"`
		Issuer  struct {
			ID string `json:"id"`
		} `json:"issuer"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return 0, "", fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
	}

	return response.Horizon, response.Issuer.ID, nil
}

func issuerMarketCap(id string) (*float64, error) {
	url := fmt.Sprintf("http://marketdata.service.consul/issuers/%s?view=full", id)
	if environment == "PROD" {
		url = fmt.Sprintf("https://api.edgelab.ch/cerberus/issuers/%s?view=full", id)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create the request: %w", err)
	}
	req.Header.Set("x-internal-service", "validation")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if environment == "PROD" {
		req.Header.Set("Authorization", "Bearer "+tokenPROD)
	}
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
		log.Infof("failed with issuer %s, status code %d, response %s", id, res.StatusCode, string(raw))
		return nil, nil
	}

	var response struct {
		MarketValue *float64 `json:"marketValue"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
	}

	if response.MarketValue == nil {
		return nil, nil
	}

	value := *response.MarketValue * 1000000

	return &value, nil
}

func marketCapToHorizon(marketCap *float64) int {
	const (
		maxHorizon = 30

		//  Normalize between 0 and 1, kind of
		liquid   = 10.0
		illiquid = 8.0
	)

	if marketCap == nil {
		return maxHorizon / 2.0
	}

	if *marketCap <= 0 {
		return maxHorizon
	}

	// Usual definition:
	// Mega : cap > 200 B$ (2e11)
	// Large: cap > 10 B$ (1e10)
	// Med  : cap > 2 B$ (2e9)
	// Small: cap > 250 M$ (2.5e8)
	// Micro: cap > 50M$ (5e7)
	// Nano : cap otherwise

	mlog := (math.Log10(*marketCap) - illiquid) / (liquid - illiquid)

	return int(maxHorizon * (1.0 - min(max(mlog, 0.0), 1.0)))
}
