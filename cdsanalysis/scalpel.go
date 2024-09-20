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
	scalpelURL     = "http://scalpel.service.consul/credit/%s/%s/timeseries?from=%s&to=%s"
	scalpelURLPROD = "https://api.edgelab.ch/scalpel/credit/%s/%s/timeseries?from=%s&to=%s"

	fromDate = "2022-01-01"
	toDate   = "2024-09-19"
)

var tenors = []string{"M12", "Y7", "Y20", "Y50"}

// CreditCurve is a map of tenor to date to credit spread.
type CreditCurve map[string]TimeSeries

type TimeSeries map[string]float64

func requestCreditCurves(issuerIDs []string) (map[string]CreditCurve, error) {
	curves := make(map[string]CreditCurve, len(issuerIDs))
	for _, issuerID := range issuerIDs {
		curve, err := requestCreditCurve(issuerID)
		if err != nil {
			return nil, err
		}

		if curve == nil {
			continue
		}

		curves[issuerID] = curve
	}

	return curves, nil
}

func requestCreditCurve(issuerID string) (CreditCurve, error) {
	curve := make(CreditCurve, len(tenors))
	for _, tenor := range tenors {
		url := fmt.Sprintf(scalpelURL, issuerID, tenor, fromDate, toDate)

		// Fetch the credit curve from Scalpel.
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create the request: %w", err)
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
			return nil, fmt.Errorf("could not send the request: %w", err)
		}
		defer res.Body.Close()

		raw, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read response body: %w", err)
		}

		if res.StatusCode != http.StatusOK {
			log.Errorf("failed with issuer %s, status code %d, response %s", issuerID, res.StatusCode, string(raw))

			return nil, nil
		}

		var ts TimeSeries
		if err := json.Unmarshal(raw, &ts); err != nil {
			return nil, fmt.Errorf("could not unmarshal the response: %w", err)
		}

		curve[tenor] = ts
	}

	return curve, nil
}
