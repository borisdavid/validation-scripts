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

	assetIDs := make([]string, 0)
	for _, record := range records[1:] {
		assetIDs = append(assetIDs, record[0])
	}

	log.Info("Starting the process by fetching liquidity in MD")
	outputMD, err := requestMarketdata(assetIDs)
	if err != nil {
		log.Fatal("Error while calling request", err)
	}

	log.Info("fetch requests in Adam")
	outputAdam, err := requestAdam(assetIDs)
	if err != nil {
		log.Fatal("Error while calling request", err)
	}

	log.Info("fetch responses in Eve")
	outputEve, err := requestEve(outputAdam)
	if err != nil {
		log.Fatal("Error while calling request", err)
	}

	log.Info("Save results in csv")
	// Write output csv.
	err = outputToCsv(outputMD, outputEve)
	if err != nil {
		log.Fatalf("Error converting results to csv: %v", err)
	}
}
