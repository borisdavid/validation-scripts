package main

import (
	"encoding/csv"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func outputToCsv(output []SensitivityOutput) error {
	log.Infof("Building output csv")

	csvFile, err := os.Create("output.csv")
	if err != nil {
		return fmt.Errorf("error while creating report file: %s", err)
	}
	defer csvFile.Close()

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	if err := csvwriter.Write([]string{"id", "currency", "perpetual", "callable", "rho", "rho-to-call", "rho-to-maturity"}); err != nil {
		return fmt.Errorf("error while writing id: %s", err)
	}

	for _, result := range output {
		rhoToCall := ""
		if result.RhoToCall != nil {
			rhoToCall = fmt.Sprintf("%f", *result.RhoToCall)
		}

		err := csvwriter.Write([]string{
			result.Asset,
			result.Currency,
			fmt.Sprintf("%t", result.Perpetual),
			fmt.Sprintf("%t", result.Callable),
			fmt.Sprintf("%f", result.Rho),
			rhoToCall,
			fmt.Sprintf("%f", result.RhoToMaturity),
		})
		if err != nil {
			return fmt.Errorf("error while writing results: %s", err)
		}
	}

	return nil
}
