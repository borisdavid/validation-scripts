package main

import (
	"encoding/csv"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Read input csv.
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

	bondIDs := make([]string, 0)
	for _, record := range records[1:] {
		bondIDs = append(bondIDs, record[0])
	}

	// Call Adam's sensitivity for each possible value and each parameter of the sensitivity endpoint.
	output, err := sensitivityAdam(bondIDs)
	if err != nil {
		log.Fatal("Error while calling sensitivity", err)
	}

	// Call etymologist.
	output, err = etymologist(output)
	if err != nil {
		log.Fatal("Error while calling etymologist", err)
	}

	// Write output csv.
	err = outputToCsv(output)
	if err != nil {
		log.Fatalf("Error converting results to csv: %v", err)
	}

	log.Info("Done")
}
