package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

const eveURL = "http://eve-live.service.consul/debug/value"

type eveOutput struct {
	ID                      string
	HorizonNoTradingVolumes int
	HorizonTradingVolumes   *int
}

func requestEve(requests []Request) (map[string]eveOutput, error) {
	ctx := context.Background()
	output := make(map[string]eveOutput, len(requests))
	counter := 0
	maxNumber := len(requests)
	for _, request := range requests {
		baseResponse, err := makeRequestEve(ctx, request.Payload)
		/* 	file, _ := json.MarshalIndent(request.Payload, "", " ")
		_ = os.WriteFile(fmt.Sprintf("%s.json", request.ID), file, 0644)
		*/
		if err != nil {
			file, _ := json.MarshalIndent(request.Payload, "", " ")
			_ = os.WriteFile(fmt.Sprintf("%s.json", request.ID), file, 0644)

			continue
		}

		var result struct {
			Liquidity struct {
				Horizon struct {
					Value int `json:"value"`
				} `json:"horizon"`
			} `json:"liquidity"`
		}
		if err := json.Unmarshal(baseResponse, &result); err != nil {
			log.Infof("Error while unmarshalling response: %v", err)

			continue
		}

		// Update payload by removing trading volumes.
		payload := request.Payload
		var payloadMap map[string]any
		if err := json.Unmarshal(payload, &payloadMap); err != nil {
			log.Infof("Error while unmarshalling payload: %v", err)
		}

		if _, ok := payloadMap["tradingVolumes"]; !ok {
			output[request.ID] = eveOutput{
				ID:                      request.ID,
				HorizonNoTradingVolumes: result.Liquidity.Horizon.Value,
			}

			counter++
			if counter%100 == 0 {
				log.Printf("Processed %d/%d bonds (%f%%)", counter, maxNumber, float64(counter)/float64(maxNumber)*100)
			}

			continue
		}

		delete(payloadMap, "tradingVolumes")
		payload, err = json.Marshal(payloadMap)
		if err != nil {
			log.Infof("Error while marshalling payload: %v", err)
		}

		updatedResponse, err := makeRequestEve(ctx, payload)
		if err != nil {
			file, _ := json.MarshalIndent(payload, "", " ")
			_ = os.WriteFile(fmt.Sprintf("%s-2.json", request.ID), file, 0644)

			continue
		}

		var result2 struct {
			Liquidity struct {
				Horizon struct {
					Value int `json:"value"`
				} `json:"horizon"`
			} `json:"liquidity"`
		}
		if err := json.Unmarshal(updatedResponse, &result2); err != nil {
			log.Infof("Error while unmarshalling response 2: %v", err)

			continue
		}

		output[request.ID] = eveOutput{
			ID:                      request.ID,
			HorizonTradingVolumes:   &result.Liquidity.Horizon.Value,
			HorizonNoTradingVolumes: result2.Liquidity.Horizon.Value,
		}

		counter++
		if counter%100 == 0 {
			log.Printf("Processed %d/%d bonds (%f%%)", counter, maxNumber, float64(counter)/float64(maxNumber)*100)
		}
	}

	return output, nil
}

func makeRequestEve(ctx context.Context, body json.RawMessage) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, eveURL, bytes.NewReader(body))
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

	if s := res.StatusCode; s != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", s)
	}

	return json.RawMessage(raw), nil
}
