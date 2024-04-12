package main

import (
	"encoding/csv"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func outputToCsv(output []SensitivitiesOutput) error {
	log.Infof("Building output csv")

	csvFile, err := os.Create("output.csv")
	if err != nil {
		return fmt.Errorf("error while creating report file: %s", err)
	}
	defer csvFile.Close()

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	if err := csvwriter.Write([]string{"id", "cs01", "dv01", "rho", "convexity", "cs01-bis", "dv01-bis", "rho-bis", "convexity-bis"}); err != nil {
		return fmt.Errorf("error while writing id: %s", err)
	}

	for _, result := range output {
		err := csvwriter.Write([]string{
			result.Asset,
			fmt.Sprintf("%f", result.SensitivitiesToCall.CS01),
			fmt.Sprintf("%f", result.SensitivitiesToCall.DV01),
			fmt.Sprintf("%f", result.SensitivitiesToCall.Rho),
			fmt.Sprintf("%f", result.SensitivitiesToCall.Convexity),
			fmt.Sprintf("%f", result.SensitivitiesTruncated.CS01),
			fmt.Sprintf("%f", result.SensitivitiesTruncated.DV01),
			fmt.Sprintf("%f", result.SensitivitiesTruncated.Rho),
			fmt.Sprintf("%f", result.SensitivitiesTruncated.Convexity),
		})
		if err != nil {
			return fmt.Errorf("error while writing results: %s", err)
		}
	}

	return nil
}
