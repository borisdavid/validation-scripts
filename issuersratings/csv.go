package main

import (
	"encoding/csv"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func outputToCsv(outputFolder string, issuersRatings chan *IssuerCreditRating) error {
	log.Infof("Building output csvs")

	csvFile, err := os.Create(outputFolder + "issuers_rating.csv")
	if err != nil {
		return fmt.Errorf("error while creating report file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	if err := csvwriter.Write([]string{"id", "shortTerm", "longTerm"}); err != nil {
		return fmt.Errorf("error while writing id: %s", err)
	}

	for rating := range issuersRatings {
		err := csvwriter.Write([]string{
			rating.Issuer,
			rating.ShortTerm.Rating,
			rating.LongTerm.Rating,
		})
		if err != nil {
			return fmt.Errorf("error while writing results: %s", err)
		}
	}

	return nil
}
