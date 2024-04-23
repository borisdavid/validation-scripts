package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const edgelabBearerToken = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NzE4NDA1MSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiJkYWYzOWU2Mi03NTA0LTRkM2QtYmZlZi0yMGFkOWZiYTY5ZTciLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5lZGdlbGFiLmNoL3Nlc2FtZS8iLCJhdWQiOiJodHRwczovL2FwaS5lZGdlbGFiLmNoIiwiaWF0IjoxNzEzNzkxMTI3LCJleHAiOjE3MTM4MjcxMjcsIm5iZiI6MTcxMzc5MTEyNywiaHR0cHM6Ly9lZGdlbGFiLmNoL29yZ2FuaXphdGlvbiI6IjEyYWM5MjNiLTMxM2QtNGUwOS05Y2Y1LTBkNDNlMWY2MzBlYyJ9.f0Z2qZjg6ndbYQaI1wMjBw2Eeghe9_6364ZyeT5YnllRKeMDV2ADet_bwRsbuT_OdQKFFLc_VunPjaq_WzDfVCzBDJ5hrT9xGjODCTfbvVpiZM6oOjniOO8c7I5KtJrl1jPrs1arHcaCk62-Rj9xP155_6q-wnDYgD9mviMa1QqnMEwTTOefxD3HQnTqgOWiU565ozcwC2QBaXg1_4MLmnWmo16-YoLTAYXm4X3DYWjt-03R4_77KhYm3Wq-5ovWQe529Z6EnEsCZdFSab_KBbggEoWVrDUTmoa1g30oK6CyXNqZXdCowN4ZRTcidM0Wq2QHC8LH3h6sZl-L3jrKMw"

func filterCocoBonds(bondIDs []string) ([]string, error) {
	maxNumber := len(bondIDs)
	cocoBonds := make([]string, 0, maxNumber)
	counter := 0
	for _, id := range bondIDs {

		isCoco, err := isCocoBond(id)
		if err != nil {
			return nil, fmt.Errorf("could not check if bond is coco: %w", err)
		}

		if isCoco {
			cocoBonds = append(cocoBonds, id)
		}

		counter++
		if counter%100 == 0 {
			log.Printf("Processed %d/%d bonds (%f%%)", counter, maxNumber, float64(counter)/float64(maxNumber)*100)
		}
	}

	return cocoBonds, nil
}

func isCocoBond(id string) (bool, error) {
	url := fmt.Sprintf("https://api.edgelab.ch/cerberus/assets/id/%s?view=full", id)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("could not create the request: %w", err)
	}
	req.Header.Set("x-internal-service", "validation")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+edgelabBearerToken)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("could not send the request: %w", err)
	}

	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return false, fmt.Errorf("could not read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		log.Infof("failed with asset %s, status code %d, response %s", id, res.StatusCode, string(raw))
		return false, nil
	}

	var response struct {
		Coco bool `json:"coco"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return false, fmt.Errorf("received non-JSON response: %s, error was: %w", string(raw), err)
	}

	return response.Coco, nil
}
