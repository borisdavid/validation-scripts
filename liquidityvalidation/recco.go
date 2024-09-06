package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type ReccoRequestInput struct {
	Context   ReccoContext   `json:"context"`
	Scenarios ReccoScenarios `json:"scenarios"`
	Portfolio ReccoPortfolio `json:"portfolio"`
}

type ReccoContext struct {
	MeasureType       string  `json:"measureType"`
	ConfidenceLevel   float64 `json:"confidenceLevel"`
	LiquidityAdjusted bool    `json:"liquidityAdjusted"`
}

type ReccoScenarios struct {
	TimeHorizon int    `json:"timeHorizon"`
	Type        string `json:"type"`
}

type ReccoPortfolio struct {
	Currency     string          `json:"currency"`
	AmountScheme string          `json:"amountScheme"`
	Positions    []ReccoPosition `json:"positions"`
}

type ReccoPosition struct {
	Asset          string  `json:"asset"`
	Amount         float64 `json:"amount"`
	IdentifierType string  `json:"identifierType"`
	Key            string  `json:"key"`
}

type ReccoOutput struct {
	Results []ReccoResult `json:"results"`
}

type ReccoResult struct {
	Key    string      `json:"key"`
	Value  float64     `json:"value"`
	Status ReccoStatus `json:"status"`
}

type ReccoStatus struct {
	Code     int      `json:"code"`
	Key      string   `json:"key"`
	Messages []string `json:"messages"`
}

const (
	reccoUrl = "https://api.dev.edge-lab.ch/recco/v2/risk-measures/es/granularities/positions"
	token    = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NjkzNDkzNSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiJhNTdiODY5Yi05YzkwLTRiZWMtOGZhZC0wZTBhZjcwYjIyOTMiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5kZXYuZWRnZS1sYWIuY2gvc2VzYW1lLyIsImF1ZCI6Imh0dHBzOi8vYXBpLmRldi5lZGdlLWxhYi5jaCIsImlhdCI6MTcyNTYxMjkxOCwiZXhwIjoxNzI1NjQ4OTE3LCJuYmYiOjE3MjU2MTI5MTgsImh0dHBzOi8vZWRnZWxhYi5jaC9vcmdhbml6YXRpb24iOiIxMmFjOTIzYi0zMTNkLTRlMDktOWNmNS0wZDQzZTFmNjMwZWMifQ.zR5q74L2jaFHoT6-g0ETmWr2G-gY9mwHwATVJuz-Nsr1bWlycGL0hCtzqhyDDcSxnrr3t2JqoSvG0fVS-a2z7DpiduNyBmOT80_pbapomnr3cDXsUwzd2Un7bzExBdWzCmk2kj3mTG8ZugfcejfAxiLwGDoKvc7V4-S4FfchbppojKlCcXGT7h3sysie5EHJoqI-zcBLD__Qi0FHq77luWwOBt1M-sNvMqKMXu01RlMkgkPLW-TAkyb4Dglo5_neml4rlzAbcLgb_LEdw6fHzFVPRTOuUy7ViIre5ccgH0iE-TCLDJaCM60zxvRvdFlbLyQCILEUWNMvWl7m4hm8UA"

	measureType  = "relative"
	currency     = "local"
	amountScheme = "quantity"
)

func requestRecco(ids []string) (map[string]float64, error) {
	ctx := context.Background()

	output := make(map[string]float64, len(ids))
	maxNumber := len(ids)

	counter := 0
	partialIDs := make([]string, 0, 50)
	for i, id := range ids {
		partialIDs = append(partialIDs, id)
		counter++
		if len(partialIDs) < 50 && i != maxNumber-1 {
			continue
		}

		outputPartial, err := requestReccoPartial(ctx, partialIDs)
		if err != nil {
			return nil, fmt.Errorf("could not fetch liquidity for partial ids: %w", err)
		}

		for k, v := range outputPartial {
			output[k] = v
		}

		partialIDs = make([]string, 0, 50)

		log.Printf("Processed Recco %d/%d assets (%f%%)", counter, maxNumber, float64(counter)/float64(maxNumber)*100)
	}

	return output, nil
}

func requestReccoPartial(ctx context.Context, ids []string) (map[string]float64, error) {
	positions := make([]ReccoPosition, 0, len(ids))
	for _, id := range ids {
		positions = append(positions, ReccoPosition{
			Asset:          id,
			Amount:         1,
			IdentifierType: "id",
			Key:            id,
		})
	}

	input := ReccoRequestInput{
		Context: ReccoContext{
			MeasureType:       measureType,
			ConfidenceLevel:   confidenceLevel,
			LiquidityAdjusted: true,
		},
		Scenarios: ReccoScenarios{
			TimeHorizon: 30,
			Type:        "historicalInnovations",
		},
		Portfolio: ReccoPortfolio{
			Currency:     currency,
			AmountScheme: "quantity",
			Positions:    positions,
		},
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the liquidity request: %w", err)
	}

	// fmt.Println(string(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reccoUrl, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("could not create the request: %w", err)
	}
	req.Header.Set("x-internal-service", "validation-script")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

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
		return nil, fmt.Errorf("unexpected status code: %d", s)
	}

	var output ReccoOutput
	if err := json.Unmarshal(raw, &output); err != nil {
		return nil, fmt.Errorf("could not unmarshal the response: %w", err)
	}

	outputMap := make(map[string]float64, len(output.Results))
	for _, o := range output.Results {
		if o.Status.Code != 200 {
			continue
		}

		outputMap[o.Key] = o.Value
	}

	return outputMap, nil
}
