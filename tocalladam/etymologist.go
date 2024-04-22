package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const etymologistURL = "http://etymologist.service.consul/description/assets"

func etymologist(output []SensitivityOutput) ([]SensitivityOutput, error) {
	ctx := context.Background()
	for i, result := range output {
		url := fmt.Sprintf("%s/%s?snapshot=%s&as-of=true", etymologistURL, result.Asset, snapshot)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create request: %w", err)
		}

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not send the request: %w", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("etymologist returned status code %d", res.StatusCode)
		}

		var etymologistOutput struct {
			Representation struct {
				Perpetual     bool `json:"perpetual"`
				DiscreteCalls []struct {
					CallDate string `json:"callDate"`
				} `json:"discreteCalls"`
			} `json:"representation"`
		}
		if err := json.NewDecoder(res.Body).Decode(&etymologistOutput); err != nil {
			return nil, fmt.Errorf("could not decode etymologist response: %w", err)
		}

		// Update the output with the etymologist results.
		output[i].Perpetual = etymologistOutput.Representation.Perpetual
		output[i].Callable = len(etymologistOutput.Representation.DiscreteCalls) > 0
	}

	return output, nil
}
