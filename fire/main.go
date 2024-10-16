package main

import (
	"encoding/csv"
	"os"

	log "github.com/sirupsen/logrus"
)

const token = "eyJraWQiOiJzZXNhbWUtZWRnZWxhYi1hcGktMTY1NjkzNDkzNSIsInR5cCI6IkpXVCIsImFsZyI6IlJTMjU2In0.eyJqdGkiOiIyNTFkOTc2ZS0zYjgwLTQ3NzAtOTdiNy1mMDM3NDBlMmI1OGMiLCJzdWIiOiJhdXRoMHw2MTcxMGFkZTJjYWVlODAwNzFiOGNkNTQiLCJpc3MiOiJodHRwczovL2ludGVybmFsLWFwaS5kZXYuZWRnZS1sYWIuY2gvc2VzYW1lLyIsImF1ZCI6Imh0dHBzOi8vYXBpLmRldi5lZGdlLWxhYi5jaCIsImlhdCI6MTcyNzY4NjM0OCwiZXhwIjoxNzI3NzIyMzQ3LCJuYmYiOjE3Mjc2ODYzNDgsImh0dHBzOi8vZWRnZWxhYi5jaC9vcmdhbml6YXRpb24iOiIxMmFjOTIzYi0zMTNkLTRlMDktOWNmNS0wZDQzZTFmNjMwZWMifQ.B6exsseXtKZp-pBHxydwe5i1lK_4xsjnC5Q3PFZsP-S2J4KUYT8ALQ_3KLlNj9jjU4ZqS7gLoFTzlE0OCHtWJ7kLLNVNXQWqX0cs-L8pi7oRWrgThqyZr6sMR6J8AtMmN-E8vRb6XK8MOv5GaTsdao_ZUo7o1iAu_2rLn8cI--KpfkWGiQniguTTjBriURw-kZsIxQtL1CvAIyK6WACw50uTr138ULnKV_835G4KMlCtabU-JMAUqzuwkuftK7DP5-YRzrRiP62U-xgB1m7aXRuwzAxqwIIj1Dd7jEgtLwqwMUPSFrgulQDaYwbXtAuHtvcqBUnD0Kec9HcbF753rQ"

func main() {
	file, err := os.Open("input.csv")
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	assetIDs := make([]string, 0)
	for _, record := range records[1:] {
		assetIDs = append(assetIDs, record[0])
	}

	log.Infof("%d issuers loaded", len(assetIDs))

	// Retrigger pricings.
	err = retriggerPricings(assetIDs)
	if err != nil {
		log.Fatal("Error while retriggering pricings", err)
	}

	log.Info("Pricings retriggered")
}
