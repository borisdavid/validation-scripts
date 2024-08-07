package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/edgelaboratories/eve/pkg/asset"
)

type RequestInput struct {
	Asset            string   `json:"asset"`
	TargetCurrencies []string `json:"targetCurrencies"`
	Run              RunDate  `json:"run"`
	AsOf             bool     `json:"asOf"`
}

type RunDate struct {
	Date string `json:"date"`
}

type Asset struct {
	Bond asset.Bond `json:"asset"`
}

type Request struct {
	Asset   Asset
	Payload json.RawMessage
}

const (
	requestURL = "http://adam-http.service.consul/debug/dump/request"
	snapshot   = "2024-04-11T00:30:05Z"
)

func requestAdam(ids []string) ([]Request, error) {
	ctx := context.Background()
	output := make([]Request, 0, len(ids))

	counter := 0
	maxNumber := len(ids)
	for _, id := range ids {
		input := RequestInput{
			Run: RunDate{
				Date: snapshot,
			},
			Asset:            id,
			TargetCurrencies: []string{"local"},
			AsOf:             true,
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

		var response Asset
		if err := json.Unmarshal(raw, &response); err != nil {
			return nil, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
		}

		if s := res.StatusCode; s != http.StatusOK {
			continue
		}

		if !response.Bond.IsPerpetual && len(response.Bond.DiscreteCallability) == 0 {
			log.Printf("Skipping bond %s as it is not perpetual and has no callability", id)

			continue
		}

		output = append(output, Request{
			Asset:   response,
			Payload: raw,
		})

		counter++
		if counter%100 == 0 {
			log.Printf("Processed %d/%d bonds (%f%%)", counter, maxNumber, float64(counter)/float64(maxNumber)*100)
		}
	}

	return output, nil
}
