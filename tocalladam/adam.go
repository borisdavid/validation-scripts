package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type SensitivityInput struct {
	Snapshot string `json:"snapshot"`
	Asset    string `json:"asset"`
	Metric   string `json:"metric"`
	Currency string `json:"currency"`
	Type     string `json:"type"`
}

type SensitivityOutput struct {
	Asset    string
	Currency string

	Rho           float64
	RhoToMaturity float64
	RhoToCall     *float64
	NPV           float64

	Perpetual bool
	Callable  bool
}

const (
	positionsURL = "http://adam-http.service.consul/sensitivity"
	priceURL     = "http://adam-http.service.consul/price"
	snapshot     = "2024-05-08T00:30:05Z"
)

func sensitivityAdam(ids []string) ([]SensitivityOutput, error) {
	ctx := context.Background()

	count := 0
	output := make([]SensitivityOutput, 0, len(ids))
	for _, id := range ids {
		count++
		if count%100 == 0 {
			log.Printf("Processed %d/%d bonds (%f%%)", count, len(ids), float64(count)/float64(len(ids))*100)
		}

		if id == "670641c2-ff6d-45e2-9326-e425c37b2a9d" {
			fmt.Println("careful")
		}

		value, currency, ok, err := senstivityCall(ctx, id, "ASSET_DEPENDENT")
		if err != nil {
			return nil, fmt.Errorf("could not call sensitivity endpoint: %w", err)
		}

		if !ok {
			continue
		}

		result := SensitivityOutput{
			Asset:    id,
			Currency: currency,
			Rho:      value,
		}

		value, _, ok, err = senstivityCall(ctx, id, "TO_CALL")
		if err != nil {
			return nil, fmt.Errorf("could not call sensitivity endpoint to-call: %w", err)
		}

		if ok {
			copyValue := value
			result.RhoToCall = &copyValue
		}

		value, _, _, err = senstivityCall(ctx, id, "TO_MATURITY")
		if err != nil {
			return nil, fmt.Errorf("could not call sensitivity endpoint to-maturity: %w", err)
		}

		result.RhoToMaturity = value

		value, ok, err = npv(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("could not call npv endpoint: %w", err)
		}

		if ok {
			result.NPV = value
		}

		output = append(output, result)
	}

	return output, nil
}

func senstivityCall(ctx context.Context, id, sensitivityType string) (float64, string, bool, error) {
	input := SensitivityInput{
		Snapshot: snapshot,
		Asset:    id,
		Metric:   "RHO",
		Currency: "USD",
		Type:     sensitivityType,
	}

	body, err := json.Marshal(input)
	if err != nil {
		return 0.0, "", false, fmt.Errorf("could not marshal the yield request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, positionsURL, bytes.NewReader(body))
	if err != nil {
		return 0.0, "", false, fmt.Errorf("could not create the request: %w", err)
	}
	req.Header.Set("x-internal-service", "validation")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0.0, "", false, fmt.Errorf("could not send the request: %w", err)
	}

	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return 0.0, "", false, fmt.Errorf("could not read response body: %w", err)
	}

	var response struct {
		Result map[string]float64 `json:"result"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return 0.0, "", false, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
	}

	if s := res.StatusCode; s != http.StatusOK || len(response.Result) != 1 {
		log.Infof("failed with asset %s, status code %d, response %s", id, s, string(raw))

		return 0.0, "", false, nil
	}

	for currency, value := range response.Result {
		return value, currency, true, nil
	}

	return 0.0, "", false, nil
}

type NPVInput struct {
	Snapshot string `json:"snapshot"`
	Asset    string `json:"asset"`
	Metric   string `json:"metric"`
	Currency string `json:"currency"`
}

func npv(ctx context.Context, id string) (float64, bool, error) {
	input := NPVInput{
		Snapshot: snapshot,
		Asset:    id,
		Metric:   "NPV",
		Currency: "local",
	}

	body, err := json.Marshal(input)
	if err != nil {
		return 0.0, false, fmt.Errorf("could not marshal the yield request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, priceURL, bytes.NewReader(body))
	if err != nil {
		return 0.0, false, fmt.Errorf("could not create the request: %w", err)
	}
	req.Header.Set("x-internal-service", "validation")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0.0, false, fmt.Errorf("could not send the request: %w", err)
	}

	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return 0.0, false, fmt.Errorf("could not read response body: %w", err)
	}

	var response struct {
		Result float64 `json:"result"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return 0.0, false, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
	}

	if s := res.StatusCode; s != http.StatusOK {
		log.Infof("failed npv with asset %s, status code %d, response %s", id, s, string(raw))

		return 0.0, false, nil
	}

	return response.Result, true, nil
}
