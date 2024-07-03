package main

import (
	"encoding/csv"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func outputToCsv(outputMD []liquidityOutput, outputEve map[string]eveOutput) error {
	log.Infof("Building output csv")

	csvFile, err := os.Create("output.csv")
	if err != nil {
		return fmt.Errorf("error while creating report file: %s", err)
	}
	defer csvFile.Close()

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	if err := csvwriter.Write([]string{"id", "horizonMD", "horizonNoVolumes", "horizonVolumes"}); err != nil {
		return fmt.Errorf("error while writing id: %s", err)
	}

	for _, result := range outputMD {
		eveResult := outputEve[result.id]

		err := csvwriter.Write([]string{
			result.id,
			fmt.Sprintf("%d", result.horizon),
			fmt.Sprintf("%d", eveResult.HorizonNoTradingVolumes),
			fmt.Sprintf("%d", eveResult.HorizonTradingVolumes),
		})
		if err != nil {
			return fmt.Errorf("error while writing results: %s", err)
		}
	}

	return nil
}
