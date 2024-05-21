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

const (
	requestURL    = "https://api.edgelab.ch/adam/bond-yield"
	requestNPVURL = "https://api.edgelab.ch/adam/price"
	snapshot      = "2024-05-20T19:30:05Z"
)

type AdamMetric string

const (
	ytm AdamMetric = "YIELD_METRIC_YIELD_TO_MATURITY"
	ytc AdamMetric = "YIELD_METRIC_YIELD_TO_CALLABILITY"
	ytp AdamMetric = "YIELD_METRIC_YIELD_TO_PUTABILITY"
	ytw AdamMetric = "YIELD_METRIC_YIELD_TO_WORST"
)

type RequestInput struct {
	Snapshot  string     `json:"snapshot"`
	ValueDate string     `json:"value_date"`
	Metric    AdamMetric `json:"metric"`
	Asset     string     `json:"asset"`
	Price     struct {
		Value    float64 `json:"value"`
		Currency string  `json:"currency"`
		Type     string  `json:"type"`
	} `json:"price"`
}

type Result struct {
	AssetID string   `json:"assetId"`
	YTM     float64  `json:"ytm"`
	YTC     *float64 `json:"ytc"`
	YTP     *float64 `json:"ytp"`
	YTW     float64  `json:"ytw"`
}

func requestAdam(ids []string) ([]Result, error) {
	ctx := context.Background()
	output := make([]Result, 0, len(ids))

	counter := 0
	maxNumber := len(ids)
	for _, id := range ids {
		// Fetch NPV.
		npv, err := requestSingleAdamNPV(ctx, id)
		if err != nil {
			log.Warnf("could not fetch NPV for %s: %v", id, err)

			continue
		}

		// Fetch YTM.
		ytm, err := requestSingleAdam(ctx, id, ytm, npv)
		if err != nil {
			log.Warnf("could not fetch YTM for %s: %v", id, err)

			continue
		}

		// Fetch YTW.
		ytw, err := requestSingleAdam(ctx, id, ytw, npv)
		if err != nil {
			log.Warnf("could not fetch YTW for %s: %v", id, err)

			continue
		}

		// Try to fetch YTC and YTP.
		ytc, err := requestSingleAdam(ctx, id, ytc, npv)
		if err != nil {
			log.Tracef("could not fetch YTC for %s: %v", id, err)
		}
		ytp, err := requestSingleAdam(ctx, id, ytp, npv)
		if err != nil {
			log.Tracef("could not fetch YTP for %s: %v", id, err)
		}

		output = append(output, Result{
			AssetID: id,
			YTM:     *ytm,
			YTC:     ytc,
			YTP:     ytp,
			YTW:     *ytw,
		})

		counter++
		if counter%100 == 0 {
			log.Printf("Processed %d/%d bonds (%f%%)", counter, maxNumber, float64(counter)/float64(maxNumber)*100)
		}
	}

	return output, nil
}

type RequestNPVInput struct {
	Snapshot string `json:"snapshot"`
	Asset    string `json:"asset"`
	Currency string `json:"currency"`
	Metric   string `json:"metric"`
}

func requestSingleAdamNPV(ctx context.Context, id string) (float64, error) {
	input := RequestNPVInput{
		Snapshot: snapshot,
		Metric:   "NPV",
		Asset:    id,
		Currency: "local",
	}

	body, err := json.Marshal(input)
	if err != nil {
		return 0.0, fmt.Errorf("could not marshal the yield request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestNPVURL, bytes.NewReader(body))
	if err != nil {
		return 0.0, fmt.Errorf("could not create the request: %w", err)
	}
	req.Header.Set("x-internal-service", "validation")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiJkODI0NWY0Yy02YmU2LTQ0MmMtYjQ0OC1jYTQ5ZDQyMzdkMGUiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzE2MjkxODk3LCJleHAiOjE3MTYzMjc4OTYsIm5iZiI6MTcxNjI5MTg5NywiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.GTFea7TI-gDdP3p4IJcv5RJpIE6HsMIdVgG1g9zlj4ZAqujXyGqV4RcY7lpqsaFYNoGk7vewFWdXgKcUtW_BEF1bWzwshyLOZoPnJ7KVJmmOR9KA42v3q9dHRKDwtBvrL9-l8UJc29rtgRoBnsg4wnCmmCmmb5boDQQIq5Ov9NbzMHstVJwA3zhQ34FpVWOqowosGN0Zn0AgcJQEeZXyW9-KkJkUJ5JobgGRRvo_3ToREM_P8oan6ILEiY4FCasGsF6aMxwJ7uQLA85UfOPMB4GnaAuFtdNJTkGT5xZMaHiawcQ-1tEUK8yak77e1Q03TYr27xUf7h1vSOXy7gfTxw")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0.0, fmt.Errorf("could not send the request: %w", err)
	}

	defer res.Body.Close()

	if s := res.StatusCode; s != http.StatusOK {
		return 0.0, fmt.Errorf("received non-200 status code: %d", s)
	}

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return 0.0, fmt.Errorf("could not read response body: %w", err)
	}

	var response struct {
		Result float64 `json:"result"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return 0.0, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
	}

	return response.Result, nil
}

func requestSingleAdam(ctx context.Context, id string, metric AdamMetric, npv float64) (*float64, error) {
	input := RequestInput{
		Snapshot:  snapshot,
		ValueDate: snapshot,
		Metric:    metric,
		Asset:     id,
		Price: struct {
			Value    float64 `json:"value"`
			Currency string  `json:"currency"`
			Type     string  `json:"type"`
		}{
			Value:    npv,
			Currency: "USD",
			Type:     "PRICE_TYPE_DIRTY",
		},
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the yield request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("could not create the request: %w", err)
	}
	req.Header.Set("x-internal-service", "validation")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiJkODI0NWY0Yy02YmU2LTQ0MmMtYjQ0OC1jYTQ5ZDQyMzdkMGUiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzE2MjkxODk3LCJleHAiOjE3MTYzMjc4OTYsIm5iZiI6MTcxNjI5MTg5NywiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.GTFea7TI-gDdP3p4IJcv5RJpIE6HsMIdVgG1g9zlj4ZAqujXyGqV4RcY7lpqsaFYNoGk7vewFWdXgKcUtW_BEF1bWzwshyLOZoPnJ7KVJmmOR9KA42v3q9dHRKDwtBvrL9-l8UJc29rtgRoBnsg4wnCmmCmmb5boDQQIq5Ov9NbzMHstVJwA3zhQ34FpVWOqowosGN0Zn0AgcJQEeZXyW9-KkJkUJ5JobgGRRvo_3ToREM_P8oan6ILEiY4FCasGsF6aMxwJ7uQLA85UfOPMB4GnaAuFtdNJTkGT5xZMaHiawcQ-1tEUK8yak77e1Q03TYr27xUf7h1vSOXy7gfTxw")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send the request: %w", err)
	}

	defer res.Body.Close()

	if s := res.StatusCode; s != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", s)
	}

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	var response struct {
		Value float64 `json:"value"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
	}

	return &response.Value, nil
}
