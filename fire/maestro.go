package main

import (
	"context"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const maestroHost = "https://api.dev.edge-lab.ch/maestro"

func retriggerPricings(assets []string) error {
	for _, asset := range assets {
		log.Infof("Retriggering pricing for asset %s", asset)
		err := retriggerPricing(asset)
		if err != nil {
			log.Errorf("Error while retriggering pricing for asset %s: %v", asset, err)
		}
	}

	return nil
}

func retriggerPricing(asset string) error {
	// Call the pricing service.
	url := fmt.Sprintf("%s/pricing/%s", maestroHost, asset)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, url, nil)
	if err != nil {
		return fmt.Errorf("could not create http request for url %s: %w", url, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send the request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return fmt.Errorf("received non-202 response: %d", res.StatusCode)
	}

	return nil
}
