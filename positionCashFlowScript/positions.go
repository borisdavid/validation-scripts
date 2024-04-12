package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PositionsContext struct {
	Snapshot string `json:"snapshot"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Currency string `json:"currency"`
	Unit     string `json:"unit"`
}

type Position struct {
	Quantity float64 `json:"quantity"`
	Asset    string  `json:"asset"`
}

type PositionsInput struct {
	Context   PositionsContext    `json:"context"`
	Positions map[uint32]Position `json:"positions"`
}

type CashFlowOutput struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
	Type   string  `json:"type"`
}

type PositionsCashFlowsOutput struct {
	Results map[uint32]struct {
		CashFlows []CashFlowOutput `json:"cashFlows"`
	} `json:"results"`
}

type date struct {
	Year  int
	Month time.Month
	Day   int
}

func (d date) string() string {
	return fmt.Sprintf("%d-%02d-%02d", d.Year, d.Month, d.Day)
}

const (
	//positionsURL   = "http://arcanist-http.service.consul/v6/positions/cash-flows"
	positionsURL = "https://api.edgelab.ch/arcanist/v6/positions/cash-flows"

	outputCurrency = "USD"

	unit = "ABSOLUTE"
)

var (
	snapshot = date{2023, 3, 1}
	start    = date{2023, 3, 1}
	end      = date{2040, 3, 1}
)

func positionsCashFlows(ctx context.Context) (*PositionsCashFlowsOutput, error) {
	input := PositionsInput{
		Context: PositionsContext{
			Snapshot: time.Date(snapshot.Year, snapshot.Month, snapshot.Day, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Start:    time.Date(start.Year, start.Month, start.Day, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			End:      time.Date(end.Year, end.Month, end.Day, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Currency: outputCurrency,
			Unit:     unit,
		},
		Positions: sliceToPositions(data),
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the yield request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, positionsURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("could not create the request: %w", err)
	}

	req.Header.Set("x-internal-service", "validation")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiIyYjBhYjNiMi0yMWM2LTQzMDUtYTQ1Ny1lNmE2MTdlZDNkZmMiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzExNDQ2NTMwLCJleHAiOjE3MTE0ODI1MjgsIm5iZiI6MTcxMTQ0NjUzMCwiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.pEMmNCQuSNXoUlH7Jgn4fKK6VIsJRUgWpnIH6dlEQuoE_nolB6FYH_VXiIcGFvPcL957hyzzraxvkN7WXtOsnhEILQQOROHnDXIh2ToZR1xq5w_ncrG7dYui9SPcvfoxCVOrISdCplhnI8tQBhgFv10izZFlb-FPZRhntsercY4_pwi_y7EyKwZn-kObOjsA3qJhoQbAKeXEMRM727fJHMh-WM5iEyNVB-N6HYD9UJQEWWnZkVPV-bRF9oENgGPXu1I2vO2TU4n9tx1xrksMScxtT-eEiXxLHIPILcCHwnleJcTNBPyZa3_51t2PfbeXw2C4Xmff_GqpuhVLdW-YPA")

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

	var response PositionsCashFlowsOutput
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
	}

	if s := res.StatusCode; s != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d, response was: %s", s, string(raw))
	}

	return &response, nil
}

func sliceToPositions(data []inputData) map[uint32]Position {
	positions := make(map[uint32]Position)

	for i, d := range data {
		positions[uint32(i)] = Position{
			Quantity: d.quantity,
			Asset:    d.id,
		}
	}

	return positions
}
