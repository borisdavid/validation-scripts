package main

import (
	"encoding/csv"
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Read input csv.
	file, err := os.Open("credit_suspect_assets.csv")
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error while reading the file", err)
	}

	bondIDs := make([]string, 0)
	for _, record := range records[1:] {
		bondIDs = append(bondIDs, record[1])
	}

	// Call cerberus for coco bonds.
	cocoBonds, err := filterCocoBonds(bondIDs)
	if err != nil {
		log.Fatal("Error while filtering coco bonds", err)
	}

	// Write output csv.
	output, _ := json.MarshalIndent(cocoBonds, "", " ")
	_ = os.WriteFile("output.json", output, 0644)

	log.Info("Done")
}
