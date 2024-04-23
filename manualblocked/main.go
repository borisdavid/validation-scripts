package main

import (
	"context"
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	// Call Hippo for manuals.
	manualProxies, err := readManualProxies(ctx)
	if err != nil {
		log.Fatal("Error while reading manual proxies", err)
	}

	// Call Hippo for blocked.
	blockedProxies, err := readBlockedProxies(ctx)
	if err != nil {
		log.Fatal("Error while reading blocked proxies", err)
	}

	// Find the intersection.
	manualAndBlockedProxies, affectedIssuers := findIntersection(manualProxies, blockedProxies)

	// Save result in json file.
	output, _ := json.MarshalIndent(manualAndBlockedProxies, "", " ")
	_ = os.WriteFile("output.json", output, 0644)

	outputIssuers, _ := json.MarshalIndent(affectedIssuers, "", " ")
	_ = os.WriteFile("output-issuers.json", outputIssuers, 0644)

	log.Info("Done")
}

func findIntersection(manualProxies []proxyPair, blockedProxies map[string]struct{}) ([]string, []string) {
	intersection := make([]string, 0)
	affectedIssuers := make([]string, 0)
	for _, p := range manualProxies {
		if _, ok := blockedProxies[p.Proxy]; ok {
			intersection = append(intersection, p.Proxy)
			affectedIssuers = append(affectedIssuers, p.Issuer)
		}
	}

	return intersection, affectedIssuers
}
