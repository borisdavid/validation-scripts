package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

const outputLocation = "./positions"

func positionsToCsv(output *PositionsCashFlowsOutput) error {
	logrus.Infof("Building %s file...", outputLocation)

	for id, result := range output.Results {
		data := data[id]

		fileName := fmt.Sprintf("%s/%s.csv", outputLocation, data.id)
		csvFile, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("error while creating report file: %s", err)
		}
		defer csvFile.Close()

		csvwriter := csv.NewWriter(csvFile)
		defer csvwriter.Flush()

		// ID.
		err = csvwriter.Write([]string{"id", data.id})
		if err != nil {
			return fmt.Errorf("error while writing id: %s", err)
		}

		// Currency.
		err = csvwriter.Write([]string{"input currency", data.currency})
		if err != nil {
			return fmt.Errorf("error while writing currency: %s", err)
		}

		// Quantity.
		err = csvwriter.Write([]string{"quantity", fmt.Sprintf("%f", data.quantity)})
		if err != nil {
			return fmt.Errorf("error while writing quantity: %s", err)
		}

		err = csvwriter.Write([]string{})
		if err != nil {
			return fmt.Errorf("error while writing new line: %s", err)
		}

		// Input Dates.
		err = csvwriter.Write([]string{"snapshot", snapshot.string()})
		if err != nil {
			return fmt.Errorf("error while writing snapshot: %s", err)
		}

		err = csvwriter.Write([]string{"start", start.string()})
		if err != nil {
			return fmt.Errorf("error while writing start: %s", err)
		}

		err = csvwriter.Write([]string{"end", end.string()})
		if err != nil {
			return fmt.Errorf("error while writing end: %s", err)
		}

		// Output Currency.
		err = csvwriter.Write([]string{"output currency", outputCurrency})
		if err != nil {
			return fmt.Errorf("error while writing output currency: %s", err)
		}

		// Unit.
		err = csvwriter.Write([]string{"unit", unit})
		if err != nil {
			return fmt.Errorf("error while writing unit: %s", err)
		}

		err = csvwriter.Write([]string{})
		if err != nil {
			return fmt.Errorf("error while writing new line: %s", err)
		}

		err = positionToCsv(result.CashFlows, csvwriter)
		if err != nil {
			return err
		}
	}

	return nil
}

func positionToCsv(cashFlows []CashFlowOutput, csvWriter *csv.Writer) error {
	// Write header
	err := csvWriter.Write([]string{"Date", "Amount", "Type"})
	if err != nil {
		return err
	}

	for _, cf := range cashFlows {
		err := csvWriter.Write([]string{cf.Date, fmt.Sprintf("%f", cf.Amount), cf.Type})
		if err != nil {
			return err
		}
	}

	return nil
}
