package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const hippoURL = "http://hippo.service.consul/credit/proxies"

//const hippoURL = "https://api.edgelab.ch/hippo/credit/proxies"

type proxyPair struct {
	Issuer string `json:"issuer"`
	Proxy  string `json:"proxy"`
}

func readManualProxies(ctx context.Context) ([]proxyPair, error) {
	data, err := makeRequestManualHippo(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get manual proxies: %w", err)
	}

	var proxies []proxyPair
	if err := json.Unmarshal(data, &proxies); err != nil {
		return nil, fmt.Errorf("could not unmarshal manual proxies: %w", err)
	}

	return proxies, nil
}

func readBlockedProxies(ctx context.Context) (map[string]struct{}, error) {
	data, err := makeRequestBlockedHippo(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get manual proxies: %w", err)
	}

	var blocked []string
	if err := json.Unmarshal(data, &blocked); err != nil {
		return nil, fmt.Errorf("could not unmarshal manual proxies: %w", err)
	}

	uniqueBlocked := make(map[string]struct{})
	for _, p := range blocked {
		uniqueBlocked[p] = struct{}{}
	}

	return uniqueBlocked, nil
}

func makeRequestManualHippo(ctx context.Context) (json.RawMessage, error) {
	data, err := makeRequestHippo(ctx, hippoURL+"/manual/bulk")
	if err != nil {
		return nil, fmt.Errorf("could not make request to hippo: %w", err)
	}

	return data, nil
}

func makeRequestBlockedHippo(ctx context.Context) (json.RawMessage, error) {
	data, err := makeRequestHippo(ctx, hippoURL+"/block/bulk")
	if err != nil {
		return nil, fmt.Errorf("could not make request to hippo: %w", err)
	}

	return data, nil
}

func makeRequestHippo(ctx context.Context, url string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create the request: %w", err)
	}
	req.Header.Set("x-internal-service", "validation")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiI3ZDUyYWM3OS0zMDdhLTRlZGQtOTUyNy1kNzU4MDI4ZTI2MDIiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzEzODc0MTEwLCJleHAiOjE3MTM5MTAxMTAsIm5iZiI6MTcxMzg3NDExMCwiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.mvxZdBxdOxp5-orl-bXgiId7LcoiyncDAJ9DifF3AFXfn_uMSTCj1KdZxe7Iwer84XqfWchJ8f9UhxeoboqIb2SoSCufXXwm9YaR6sfOK4lHUCZXPILpQrhokEeBFUjbNCNXcN1PRdG_L7bVrLAz94jYXzkj0-a4ycYT8Dx7EOg1_B7t83ziR40VlPiwtdoH55BQZt6gWvQfGj6MPNO9vzjWmgSHniC1Oy2jLfdhEZkiGiM4RcH3dhYZnD59Oq55qAQp5MM8ANcFmKCGghJZCQ1ePs2xdfGbUEOg6uQXS8K7V_CKwj7qg_iZifV17jQT8h2HHqUuDNtE2i6BF6bB9A")

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
