package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type ArcanistRequestInput struct {
	Context      ArcanistContext          `json:"context"`
	QuantityUnit string                   `json:"quantityUnit"`
	Positions    map[int]ArcanistPosition `json:"positions"`
}

type ArcanistContext struct {
	Snapshot              string                      `json:"snapshot"`
	Metric                string                      `json:"metric"`
	MetricUnit            string                      `json:"metricUnit"`
	MetricCurrency        string                      `json:"metricCurrency"`
	ConfidenceLevel       float64                     `json:"confidenceLevel"`
	Scenarios             map[string]ArcanistScenario `json:"scenarios"`
	TimeHorizon           TimeHorizon                 `json:"timeHorizon"`
	FetchLiquidityHorizon bool                        `json:"fetchLiquidityHorizon"`
}

type ArcanistScenario struct {
	ID        int     `json:"id"`
	Weight    float64 `json:"weight"`
	Amplitude float64 `json:"amplitude"`
}

type TimeHorizon struct {
	ScenarioHorizon Value `json:"scenarioHorizon"`
}

type Value struct {
	Value float64 `json:"value"`
}

type ArcanistPosition struct {
	Asset     string  `json:"asset"`
	Quantity  float64 `json:"quantity"`
	Currency  string  `json:"currency"`
	Liquidity float64 `json:"liquidity"`
}

type ArcanistOutput struct {
	Results map[int]ArcanistResult `json:"results"`
}

type ArcanistResult struct {
	Result *float64 `json:"result"`
}

const (
	arcanistRequestURL = "http://arcanist-http.service.consul/v6/positions/quantile-risk-measure"

	snapshotArcanist = "2024-09-12T00:30:05Z"
	metric           = "ES"
	metricUnit       = "RELATIVE"
	metricCurrency   = "local"
	confidenceLevel  = 0.9
	quantityUnit     = "ABSOLUTE"
)

func requestArcanist(ids []liquidityOutput) (map[string]float64, map[string]float64, error) {
	ctx := context.Background()

	output := make(map[string]float64, len(ids))
	outputMDLiquidity := make(map[string]float64, len(ids))
	maxNumber := len(ids)

	counter := 0
	partialIDs := make([]liquidityOutput, 0, 50)
	for i, id := range ids {
		partialIDs = append(partialIDs, id)
		counter++
		if len(partialIDs) < 50 && i != maxNumber-1 {
			continue
		}

		outputPartial, err := requestArcanistPartial(ctx, partialIDs, true)
		if err != nil {
			return nil, nil, fmt.Errorf("could not fetch liquidity for partial ids: %w", err)
		}

		for k, v := range outputPartial {
			output[k] = v
		}

		outputPartialMD, err := requestArcanistPartial(ctx, partialIDs, false)
		if err != nil {
			return nil, nil, fmt.Errorf("could not fetch liquidity for partial ids: %w", err)
		}

		for k, v := range outputPartialMD {
			outputMDLiquidity[k] = v
		}

		partialIDs = make([]liquidityOutput, 0, 50)

		log.Printf("Processed Arcanist %d/%d assets (%f%%)", counter, maxNumber, float64(counter)/float64(maxNumber)*100)
	}

	return output, outputMDLiquidity, nil
}

func requestArcanistPartial(ctx context.Context, ids []liquidityOutput, withQELiquidity bool) (map[string]float64, error) {
	scenarios := make(map[string]ArcanistScenario, 500)
	for i := 1; i <= 500; i++ {
		id := i + 6499
		scenarios[fmt.Sprintf("%d", id)] = ArcanistScenario{
			ID:        id,
			Weight:    1.0,
			Amplitude: 1.0,
		}
	}

	positions := make(map[int]ArcanistPosition, len(ids))
	for i, id := range ids {
		positions[i] = ArcanistPosition{
			Asset:     id.id,
			Quantity:  1.0,
			Currency:  "USD",
			Liquidity: float64(id.horizon),
		}
	}

	input := ArcanistRequestInput{
		Context: ArcanistContext{
			Snapshot:        snapshotArcanist,
			Metric:          metric,
			MetricUnit:      metricUnit,
			MetricCurrency:  metricCurrency,
			ConfidenceLevel: confidenceLevel,
			Scenarios:       scenarios,
			TimeHorizon: TimeHorizon{
				ScenarioHorizon: Value{
					Value: 30,
				},
			},
			FetchLiquidityHorizon: withQELiquidity,
		},
		QuantityUnit: quantityUnit,
		Positions:    positions,
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the liquidity request: %w", err)
	}

	_ = os.WriteFile("aaa/aaa.json", body, 0644)

	// fmt.Println(string(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, arcanistRequestURL, bytes.NewReader(body))
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
		return nil, fmt.Errorf("unexpected status code %d: %s", s, string(raw))
	}

	var output ArcanistOutput
	if err := json.Unmarshal(raw, &output); err != nil {
		return nil, fmt.Errorf("could not unmarshal the response: %w", err)
	}

	outputMap := make(map[string]float64, len(output.Results))
	for i, o := range output.Results {
		if o.Result == nil {
			continue
		}
		outputMap[positions[i].Asset] = *o.Result
	}

	return outputMap, nil
}
