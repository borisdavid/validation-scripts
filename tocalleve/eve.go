package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/edgelaboratories/eve/pkg/asset"
	"github.com/edgelaboratories/go-libraries/date"
	log "github.com/sirupsen/logrus"
)

type SensitivitiesOutput struct {
	Asset                  string
	SensitivitiesToCall    Sensitivities
	SensitivitiesTruncated Sensitivities
}

type Sensitivities struct {
	CS01      float64
	DV01      float64
	Rho       float64
	Convexity float64
}

type PriceResult struct {
	Value float64 `json:"value"`
}

type SenstitivitiesBody struct {
	CS01      map[string]PriceResult `json:"CS01"`
	DV01      map[string]PriceResult `json:"DV01"`
	Rho       map[string]PriceResult `json:"rho"`
	Convexity map[string]PriceResult `json:"convexity"`
}

const eveURL = "http://eve-live.service.consul/debug/value"

func priceEve(requests []Request) ([]SensitivitiesOutput, error) {
	ctx := context.Background()
	output := make([]SensitivitiesOutput, 0, len(requests))
	counter := 0
	maxNumber := len(requests)
	for _, request := range requests {
		result := SensitivitiesOutput{
			Asset: request.Asset.Bond.ID,
		}

		baseResponse, err := makeRequestEve(ctx, request.Payload)
		if err != nil {
			file, _ := json.MarshalIndent(request.Payload, "", " ")
			_ = os.WriteFile(fmt.Sprintf("%s.json", request.Asset.Bond.ID), file, 0644)

			continue
		}

		var sensitivities struct {
			Sensitivities SenstitivitiesBody `json:"sensitivitiesToCall"`
		}
		if err := json.Unmarshal(baseResponse, &sensitivities); err != nil {
			log.Infof("Error while unmarshalling response from Eve for sensitivities to call: %v", err)

			continue
		}

		sensitivitiesToCall, err := toSensitivities(sensitivities.Sensitivities)
		if err != nil {
			log.Infof("Error while converting sensitivities to call: %v", err)

			continue
		}

		result.SensitivitiesToCall = sensitivitiesToCall

		// Truncate the bond and recompute the sensitivities.
		bondToCall := newBondToCall(request.Asset.Bond)
		if bondToCall == nil {
			log.Infof("Error while creating bond to call")

			continue
		}

		var objmap map[string]any
		if err := json.Unmarshal(request.Payload, &objmap); err != nil {
			return nil, fmt.Errorf("could not unmarshal the request payload: %w", err)
		}
		objmap["asset"] = bondToCall

		updatedBody, err := json.Marshal(objmap)
		if err != nil {
			return nil, fmt.Errorf("could not marshal the updated request: %w", err)
		}

		truncatedResponse, err := makeRequestEve(ctx, updatedBody)
		if err != nil {
			file, _ := json.MarshalIndent(objmap, "", " ")
			_ = os.WriteFile(fmt.Sprintf("%s-truncated.json", request.Asset.Bond.ID), file, 0644)

			continue
		}

		var truncatedSensitivities struct {
			Sensitivities SenstitivitiesBody `json:"sensitivities"`
		}
		if err := json.Unmarshal(truncatedResponse, &truncatedSensitivities); err != nil {
			log.Infof("Error while unmarshalling response from Eve for truncated sensitivities: %v", err)

			continue
		}

		sensitivitiesTruncated, err := toSensitivities(truncatedSensitivities.Sensitivities)
		if err != nil {
			log.Infof("Error while converting truncated sensitivities: %v", err)

			continue
		}

		result.SensitivitiesTruncated = sensitivitiesTruncated
		output = append(output, result)

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

func toSensitivities(body SenstitivitiesBody) (Sensitivities, error) {
	sensitivities := Sensitivities{}

	if len(body.CS01) != 1 || len(body.DV01) != 1 || len(body.Rho) != 1 || len(body.Convexity) != 1 {
		return sensitivities, fmt.Errorf("expected exactly one value for each sensitivity, got CS01: %d, DV01: %d, Rho: %d, Convexity: %d", len(body.CS01), len(body.DV01), len(body.Rho), len(body.Convexity))
	}

	for _, value := range body.CS01 {
		sensitivities.CS01 = value.Value
	}
	for _, value := range body.DV01 {
		sensitivities.DV01 = value.Value
	}
	for _, value := range body.Rho {
		sensitivities.Rho = value.Value
	}
	for _, value := range body.Convexity {
		sensitivities.Convexity = value.Value
	}

	return sensitivities, nil
}

func newBondToCall(bond asset.Bond) *asset.Bond {
	if len(bond.DiscreteCallability) == 0 {
		return nil
	}

	earliestCall := earliestCall(bond.DiscreteCallability)

	// Remove late coupons, puts, sinks.
	remainingCoupons := make([]asset.Coupon, 0, len(bond.Coupons))
	for _, coupon := range bond.Coupons {
		if !coupon.CouponPaymentDate().After(earliestCall.Date) {
			remainingCoupons = append(remainingCoupons, coupon)
		}
	}

	remainingPuts := make([]asset.DiscretePutability, 0, len(bond.DiscretePutability))
	for _, put := range bond.DiscretePutability {
		if put.Date.Before(earliestCall.Date) {
			remainingPuts = append(remainingPuts, put)
		}
	}

	remainingSinks := make([]asset.Sinkability, 0, len(bond.Sinkability))
	for _, sink := range bond.Sinkability {
		if sink.Date.Before(earliestCall.Date) {
			remainingSinks = append(remainingSinks, sink)
		}
	}

	return &asset.Bond{
		ID:                  bond.ID,
		Currency:            bond.Currency,
		Issuer:              bond.Issuer,
		Maturity:            earliestCall.Date,
		Settlement:          earliestCall.Date,
		Recovery:            bond.Recovery,
		Coupons:             remainingCoupons,
		CouponFrequency:     bond.CouponFrequency,
		DiscreteCallability: []asset.DiscreteCallability{},
		DiscretePutability:  remainingPuts,
		Sinkability:         remainingSinks,
		Notional: &asset.Notional{
			Amount:   earliestCall.Rate,
			Currency: bond.Currency,
		},
		DaycountConvention: bond.DaycountConvention,
		IsCoco:             bond.IsCoco,
	}
}

func earliestCall(calls []asset.DiscreteCallability) *asset.DiscreteCallability {
	var earliestCall asset.DiscreteCallability

	earliestCallDate := date.Max()
	for _, c := range calls {
		if c.Date.Before(earliestCallDate) {
			earliestCallDate = c.Date
			earliestCall = c
		}
	}

	return &earliestCall
}
