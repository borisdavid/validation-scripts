package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	scalpelURL     = "http://scalpel.service.consul/credit/%s/curveassets?from=%s&to=%s"
	scalpelURLPROD = "https://api.edgelab.ch/scalpel/credit/%s/curveassets?from=%s&to=%s"

	fromDate = "2022-09-01"
	toDate   = "2024-10-02"
)

type IssuerCount struct {
	ID     string
	Count  int
	Assets []string
}

func CurveAssetsCount(issuersIDs []string) (chan IssuerCount, error) {
	issuerChan := make(chan string)

	go func() {
		defer close(issuerChan)

		for _, issuerID := range issuersIDs {
			issuerChan <- issuerID
		}
	}()

	issuerCountChan := make(chan IssuerCount)
	// Prepare the tasks and launch the workers.
	eg, _ := errgroup.WithContext(context.Background())

	for i := 0; i < 4; i++ {
		eg.Go(func() error {
			for issuerID := range issuerChan {
				log.Infof("Fetching curve assets for issuer %s", issuerID)

				issuerCount, err := CurveAssetsCountByIssuer(issuerID)
				if err != nil {
					return err
				}

				issuerCountChan <- issuerCount
			}

			return nil
		})
	}

	go func() {
		defer close(issuerCountChan)

		if err := eg.Wait(); err != nil {
			// Report the first non-nil error to the results channel.
			log.Errorf("error while fetching curve assets: %v", err)
		}
	}()

	return issuerCountChan, nil
}

func CurveAssetsCountByIssuer(issuerID string) (IssuerCount, error) {
	url := fmt.Sprintf(scalpelURL, issuerID, fromDate, toDate)
	if environment == "PROD" {
		url = fmt.Sprintf(scalpelURLPROD, issuerID, fromDate, toDate)
	}

	// Fetch the credit curve from Scalpel.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return IssuerCount{}, fmt.Errorf("could not create the request: %w", err)
	}
	req.Header.Set("x-internal-service", "QE-CDS-script")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if environment == "PROD" {
		req.Header.Set("Authorization", "Bearer "+tokenPROD)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return IssuerCount{}, fmt.Errorf("could not send the request: %w", err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return IssuerCount{}, fmt.Errorf("could not read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		log.Errorf("failed with issuer %s, status code %d, response %s", issuerID, res.StatusCode, string(raw))

		return IssuerCount{
			ID:    issuerID,
			Count: 0,
		}, nil
	}

	var output []struct {
		ID   string `json:"id"`
		Used bool   `json:"used"`
	}
	if err := json.Unmarshal(raw, &output); err != nil {
		return IssuerCount{}, fmt.Errorf("could not unmarshal the response: %w", err)
	}

	count := 0
	assets := make([]string, 0, len(output))
	for _, asset := range output {
		if asset.Used {
			count++
			assets = append(assets, asset.ID)
		}
	}

	return IssuerCount{
		ID:     issuerID,
		Count:  count,
		Assets: assets,
	}, nil
}
