package main

import (
	"context"

	"github.com/sirupsen/logrus"
)

func main() {
	// Call to Arcanist.
	ctx := context.Background()
	output, err := positionsCashFlows(ctx)
	if err != nil {
		logrus.Fatalf("Error calling positionsCashFlows: %v", err)
	}

	// Conversion of results in csv.
	err = positionsToCsv(output)
	if err != nil {
		logrus.Fatalf("Error converting results to csv: %v", err)
	}
}
