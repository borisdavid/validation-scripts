package main

import (
	"encoding/csv"
	"os"

	log "github.com/sirupsen/logrus"
)

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

	bondIDs := make([]string, 0)
	for _, record := range records[1:] {
		bondIDs = append(bondIDs, record[0])
	}

	log.Info("Starting the process in Arcanist")
	output, err := requestArcanist(bondIDs)
	if err != nil {
		log.Fatal("Error while calling request", err)
	}

	log.Info("Starting the process in Eve")

	// Write output csv.
	err = outputToCsv(output)
	if err != nil {
		log.Fatalf("Error converting results to csv: %v", err)
	}
}
