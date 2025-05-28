package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/edgelaboratories/go-libraries/go-client/httpreq"
	log "github.com/sirupsen/logrus"
)

const (
	cerberusHost = "https://api.edgelab.ch/cerberus"
)

// IssuerList  lists all issuers IDs by querying cerberus.
func IssuerList() ([]string, error) {
	size := 1000
	issuers := make([]string, 0, size)
	url := fmt.Sprintf("%s/v2/issuers?size=%d", cerberusHost, size)

	for {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create http request for url %s: %w", url, err)
		}
		req.Header.Set("X-Internal-Service", "Validation")
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

		var output struct {
			Data     []string `json:"data"`
			NextPage *string  `json:"next"`
		}
		if err := json.Unmarshal(raw, &output); err != nil {
			return nil, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
		}

		issuers = append(issuers, output.Data...)
		if output.NextPage == nil {
			break
		}

		url = fmt.Sprintf("%s%s", cerberusHost, *output.NextPage)

		log.Infof("Fetched %d issuers", len(issuers))
	}

	return issuers, nil
}

type IssuerCreditRating struct {
	Issuer    string `json:"issuer"`
	ShortTerm struct {
		Rating string `json:"rating"`
	} `json:"shortTerm"`
	LongTerm struct {
		Rating string `json:"rating"`
	} `json:"longTerm"`
}

func FetchIssuerCreditRating(issuer string) (*IssuerCreditRating, error) {
	ctx := context.Background()
	client := &http.Client{}

	url := fmt.Sprintf("%s/credit-ratings/issuers/%s", cerberusHost, issuer)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create http request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	response, _, _, err := httpreq.Request[IssuerCreditRating](ctx, client, req, httpreq.WithXInternalService("hippo"))
	if err != nil {
		return &IssuerCreditRating{}, err
	}

	response.Issuer = issuer

	return &response, nil
}
