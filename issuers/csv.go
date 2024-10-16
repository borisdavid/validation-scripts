package main

import (
	"encoding/csv"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

func outputToCsv(outputFolder string, descriptions []IssuerDescription) error {
	log.Infof("Building output csvs")

	csvFile, err := os.Create(outputFolder + "issuers_descriptions.csv")
	if err != nil {
		return fmt.Errorf("error while creating report file: %s", err)
	}

	csvwriter := csv.NewWriter(csvFile)
	defer csvwriter.Flush()

	if err := csvwriter.Write([]string{"id", "name", "shortName", "country", "stateProvince", "sector name", "countryOfRisk ID", "industry name", "countryOfIncorporation ID", "market value", "market cap", "primary exchange countries ids"}); err != nil {
		return fmt.Errorf("error while writing id: %s", err)
	}

	for _, description := range descriptions {
		strings := []string{
			description.ID,
			description.Name,
			description.ShortName,
			description.Country.Name,
			description.StateProvince,
			description.Sector.Name,
			description.CountryOfRisk.ID,
			description.Industry.Name,
			description.CountryOfIncorporation.ID,
		}

		if description.MarketValue != nil {
			strings = append(strings, fmt.Sprintf("%f", *description.MarketValue))
		} else {
			strings = append(strings, "")
		}

		if description.MarketCap != nil {
			strings = append(strings, *description.MarketCap)
		} else {
			strings = append(strings, "")
		}

		primaryExchangeCountries := make([]string, 0, len(description.PrimaryExchangeCountries))
		for _, country := range description.PrimaryExchangeCountries {
			primaryExchangeCountries = append(primaryExchangeCountries, country.ID)
		}

		strings = append(strings, fmt.Sprintf("%v", primaryExchangeCountries))

		err := csvwriter.Write(strings)
		if err != nil {
			return fmt.Errorf("error while writing results: %s", err)
		}
	}

	return nil
}
