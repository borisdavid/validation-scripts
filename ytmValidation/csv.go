package main

import (
	"encoding/csv"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func outputToCsv(output []Result) error {
	log.Infof("Building output csv")

	csvFile, err := os.Create("output.csv")
	if err != nil {
		return fmt.Errorf("error while creating report file: %s", err)
	}
	defer csvFile.Close()

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	if err := csvwriter.Write([]string{"id", "ytm", "ytw", "ytc", "ytp"}); err != nil {
		return fmt.Errorf("error while writing id: %s", err)
	}

	for _, result := range output {
		csvStr := []string{
			result.AssetID,
			fmt.Sprintf("%f", result.YTM),
			fmt.Sprintf("%f", result.YTW),
		}

		if result.YTC != nil {
			csvStr = append(csvStr, fmt.Sprintf("%f", *result.YTC))
		} else {
			csvStr = append(csvStr, "")
		}

		if result.YTP != nil {
			csvStr = append(csvStr, fmt.Sprintf("%f", *result.YTP))
		} else {
			csvStr = append(csvStr, "")
		}

		err := csvwriter.Write(csvStr)
		if err != nil {
			return fmt.Errorf("error while writing results: %s", err)
		}
	}

	return nil
}
