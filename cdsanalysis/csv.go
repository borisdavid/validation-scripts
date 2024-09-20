package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"slices"

	log "github.com/sirupsen/logrus"
)

func outputToCsv(outputFolder string, creditCurves map[string]CreditCurve) error {
	log.Infof("Building output csvs")

	for issuerID, result := range creditCurves {
		csvFile, err := os.Create(outputFolder + issuerID + "-scalpel.csv")
		if err != nil {
			return fmt.Errorf("error while creating report file: %s", err)
		}
		defer csvFile.Close()

		csvwriter := csv.NewWriter(csvFile)
		defer csvwriter.Flush()

		if err := csvwriter.Write(append([]string{"id"}, tenors...)); err != nil {
			return fmt.Errorf("error while writing id: %s", err)
		}

		// dateList := make([]string, 0, len(result[tenors[0]]))
		for _, date := range creditCurveDates(result) {
			strings := []string{
				date,
			}

			for _, tenor := range tenors {
				strings = append(strings, fmt.Sprintf("%f", result[tenor][date]))
			}

			err := csvwriter.Write(strings)
			if err != nil {
				return fmt.Errorf("error while writing results: %s", err)
			}
		}
	}

	return nil
}

func creditCurveDates(curve CreditCurve) []string {
	dates := make([]string, 0, len(curve[tenors[0]]))
	for date := range curve[tenors[0]] {
		dates = append(dates, date)
	}

	// sort the dates
	slices.SortFunc(dates, func(d1, d2 string) int {
		if d1 < d2 {
			return -1
		}

		if d1 > d2 {
			return 1
		}

		return 0
	})

	return dates
}
