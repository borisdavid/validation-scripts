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
	requestURL = "https://api.edgelab.ch/arcanist/v6/instruments/metric"
	snapshot   = "2024-05-14T19:30:04Z"
)

type RequestContext struct {
	Snapshot string `json:"snapshot"`
	Metric   string `json:"metric"`
}

type RequestInput struct {
	Context     RequestContext    `json:"context"`
	Instruments map[uint32]string `json:"instruments"`
}

type Result struct {
	AssetID string   `json:"assetId"`
	YTM     float64  `json:"ytm"`
	YTC     *float64 `json:"ytc"`
	YTP     *float64 `json:"ytp"`
	YTW     float64  `json:"ytw"`
}

func requestArcanist(ids []string) ([]Result, error) {
	ctx := context.Background()
	output := make([]Result, 0, len(ids))

	counter := 0
	maxNumber := len(ids)
	for _, id := range ids {
		// Fetch YTM.
		ytm, err := requestSingleArcanist(ctx, id, "YTM")
		if err != nil {
			log.Warnf("could not fetch YTM for %s: %v", id, err)

			continue
		}

		// Fetch YTW.
		ytw, err := requestSingleArcanist(ctx, id, "YTW")
		if err != nil {
			log.Warnf("could not fetch YTW for %s: %v", id, err)

			continue
		}

		// Try to fetch YTC and YTP.
		ytc, err := requestSingleArcanist(ctx, id, "YTC")
		if err != nil {
			log.Tracef("could not fetch YTC for %s: %v", id, err)
		}
		ytp, err := requestSingleArcanist(ctx, id, "YTP")
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

func requestSingleArcanist(ctx context.Context, id string, metric string) (*float64, error) {
	input := RequestInput{
		Context: RequestContext{
			Snapshot: snapshot,
			Metric:   metric,
		},
		Instruments: map[uint32]string{
			0: id,
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
	req.Header.Set("Authorization", "Bearer eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiJmMWFiODc3YS04NGQxLTRhNGQtYmRjYy02MzM3MTUzYmY1ZTQiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzE1OTQ2MzIyLCJleHAiOjE3MTU5ODIzMjAsIm5iZiI6MTcxNTk0NjMyMiwiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.QcSfQEQf46Rgx4SAIQWxKTQfAlXHbcyFAeIZWinSKaSHhwiY-9dqL2FJlCwCbdk9MRzMwBR4SrZJPR5E_Ce8jbMrGUsELN0Afm2qQzI-oSraamlYxtAVDWw85Hw1V_MGpoBYhA3cmuZfHsI2zW1oRPk687CmU-8aX_lnyNp7bZ9pIAnANpsaqdh_uWp_xwjuFWwW0XB9bXh35whReVadNk129V9MQtf0xve2tmVmHZi0Q_glYffsl3L5dmoHeUfWozROYv51K0sebn4Gr3l4Epg6qereKnbtYs9xDCbuWmpaq9WkchIInDuIA_aeKJtabHmKRHmvyCtfewjSOckixQ")

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
		Results map[uint32]struct {
			Result *float64 `json:"result"`
			Error  *struct {
				Message string `json:"message"`
			}
		} `json:"results"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
	}

	if result, ok := response.Results[0]; ok {
		if result.Error != nil {
			return nil, fmt.Errorf("received error response: %s", result.Error.Message)
		}

		return result.Result, nil
	}

	return nil, fmt.Errorf("no result found in response")
}
