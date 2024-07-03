package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type liquidityOutput struct {
	id      string
	horizon int
}

func requestMarketdata(assetIDs []string) ([]liquidityOutput, error) {
	maxNumber := len(assetIDs)
	liquidityHorizons := make([]liquidityOutput, 0, maxNumber)
	counter := 0
	for _, id := range assetIDs {

		horizon, err := liquidityHorizon(id)
		if err != nil {
			return nil, fmt.Errorf("could not check liquidity horizon: %w", err)
		}

		liquidityHorizons = append(liquidityHorizons, liquidityOutput{
			id:      id,
			horizon: horizon,
		})

		counter++
		if counter%100 == 0 {
			log.Printf("Processed %d/%d assets (%f%%)", counter, maxNumber, float64(counter)/float64(maxNumber)*100)
		}
	}

	return liquidityHorizons, nil
}

func liquidityHorizon(id string) (int, error) {
	url := fmt.Sprintf("http://marketdata.service.consul/assets/id/%s?view=full", id)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("could not create the request: %w", err)
	}
	req.Header.Set("x-internal-service", "validation")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	// req.Header.Set("Authorization", "Bearer "+edgelabBearerToken)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("could not send the request: %w", err)
	}

	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("could not read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		log.Infof("failed with asset %s, status code %d, response %s", id, res.StatusCode, string(raw))
		return 0, nil
	}

	var response struct {
		Horizon int `json:"liquidityHorizon"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return 0, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
	}

	return response.Horizon, nil
}
