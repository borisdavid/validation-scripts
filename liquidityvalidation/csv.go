package main

import (
	"encoding/csv"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func outputToCsv(outputMD []liquidityOutput, outputEve map[string]eveOutput, outputArcanist, outputArcanistMD, outputRecco map[string]float64) error {
	log.Infof("Building output csv")

	csvFile, err := os.Create("output.csv")
	if err != nil {
		return fmt.Errorf("error while creating report file: %s", err)
	}
	defer csvFile.Close()

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	if err := csvwriter.Write([]string{"id", "marketCap", "horizonPoC", "horizonMD", "horizonNoVolumes", "horizonVolumes", "esHistInno30D-Arcanist", "esHistInno30D-ArcanistMD", "esHistInno30D-Recco"}); err != nil {
		return fmt.Errorf("error while writing id: %s", err)
	}

	for _, result := range outputMD {
		eveResult, ok := outputEve[result.id]
		if !ok {
			err := csvwriter.Write([]string{
				result.id,
				"",
				"",
				fmt.Sprintf("%d", result.horizon),
				"",
				"",
				"",
				"",
				"",
			})
			if err != nil {
				return fmt.Errorf("error while writing results: %s", err)
			}

			continue
		}

		strings := []string{
			result.id,
		}

		if result.marketCap != nil {
			strings = append(strings, fmt.Sprintf("%f", *result.marketCap))
		} else {
			strings = append(strings, "")
		}

		if result.pocHorizon != nil {
			strings = append(strings, fmt.Sprintf("%d", *result.pocHorizon))
		} else {
			strings = append(strings, "")
		}

		strings = append(strings,
			fmt.Sprintf("%d", result.horizon),
			fmt.Sprintf("%d", eveResult.HorizonNoTradingVolumes),
			fmt.Sprintf("%d", eveResult.HorizonTradingVolumes),
		)

		arcanistValue, ok := outputArcanist[result.id]
		if !ok {
			strings = append(strings, "")
		} else {
			strings = append(strings, fmt.Sprintf("%f", arcanistValue))
		}

		arcanistValueMD, ok := outputArcanistMD[result.id]
		if !ok {
			strings = append(strings, "")
		} else {
			strings = append(strings, fmt.Sprintf("%f", arcanistValueMD))
		}

		reccoValue, ok := outputRecco[result.id]
		if !ok {
			strings = append(strings, "")
		} else {
			strings = append(strings, fmt.Sprintf("%f", reccoValue))
		}

		err := csvwriter.Write(strings)
		if err != nil {
			return fmt.Errorf("error while writing results: %s", err)
		}
	}

	return nil
}
