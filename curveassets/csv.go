package main

import (
	"encoding/csv"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func outputToCsv(outputFolder string, outputChan chan IssuerCount, nbIssuers int) error {
	log.Infof("Building output csvs")

	csvFile, err := os.Create(outputFolder + "output.csv")
	if err != nil {
		return fmt.Errorf("error while creating report file: %s", err)
	}
	defer csvFile.Close()

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	if err := csvwriter.Write([]string{"id", "asset used"}); err != nil {
		return fmt.Errorf("error while writing 1st line: %s", err)
	}

	csvFileComplete, err := os.Create(outputFolder + "output-complete.csv")
	if err != nil {
		return fmt.Errorf("error while creating report file: %s", err)
	}
	defer csvFileComplete.Close()

	csvwriterComplete := csv.NewWriter(csvFileComplete)
	defer csvwriterComplete.Flush()

	if err := csvwriterComplete.Write([]string{"id", "asset used", "assets"}); err != nil {
		return fmt.Errorf("error while writing 1st line: %s", err)
	}

	count := 0
	for issuerCount := range outputChan {
		strings := []string{
			issuerCount.ID,
			fmt.Sprintf("%d", issuerCount.Count),
		}

		err := csvwriter.Write(strings)
		if err != nil {
			return fmt.Errorf("error while writing results: %s", err)
		}

		strings = append(strings, fmt.Sprintf("%v", issuerCount.Assets))
		err = csvwriterComplete.Write(strings)
		if err != nil {
			return fmt.Errorf("error while writing complete results: %s", err)
		}

		count++
		if count%100 == 0 {
			log.Infof("Completed %d/%d issuers (%.2f%%)", count, nbIssuers, float64(count)/float64(nbIssuers)*100)
		}
	}

	return nil
}
