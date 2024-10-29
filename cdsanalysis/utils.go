package main

import (
	"fmt"
	"slices"
	"time"

	"github.com/edgelaboratories/eve/pkg/marketdata"
	"github.com/edgelaboratories/go-libraries/date"
	interp "github.com/edgelaboratories/go-libraries/interpolator"
)

func convertTermStructureToRawData(input marketdata.TermStructure) (interp.XYs, error) {
	output := make(interp.XYs, 0, len(input))
	for k, v := range input {
		yf, err := k.ToYearFraction()
		if err != nil {
			return nil, fmt.Errorf("could not convert the tenor to year fraction: %w", err)
		}

		output = append(output, interp.XY{
			X: yf,
			Y: v,
		})
	}

	slices.SortFunc(output, func(a, b interp.XY) int {
		if a.X < b.X {
			return -1
		}

		return 1
	})

	return output, nil
}

func nextIMMDate(d date.Date) date.Date {
	const immDay = 20

	// Find the next IMM date
	allowedIMMMonths := map[time.Month]struct{}{
		time.March:     {},
		time.June:      {},
		time.September: {},
		time.December:  {},
	}

	day := d.Day()
	month := d.Month()
	year := d.Year()

	if _, ok := allowedIMMMonths[month]; ok {
		if day < immDay {
			return date.New(year, month, immDay)
		}

		return date.New(year, month+3, immDay)
	}

	returnedDate := date.New(year, month, immDay)
	for {
		returnedDate = returnedDate.AddDate(0, 1, 0)
		if _, ok := allowedIMMMonths[returnedDate.Month()]; ok {
			return returnedDate
		}
	}
}

func lastIMMDate(d date.Date) date.Date {
	const immDay = 20

	// Find the next IMM date
	allowedIMMMonths := map[time.Month]struct{}{
		time.March:     {},
		time.June:      {},
		time.September: {},
		time.December:  {},
	}

	day := d.Day()
	month := d.Month()
	year := d.Year()

	if _, ok := allowedIMMMonths[month]; ok {
		if day > immDay {
			return date.New(year, month, immDay)
		}

		return date.New(year, month-3, immDay)
	}

	returnedDate := date.New(year, month, immDay)
	for {
		returnedDate = returnedDate.AddDate(0, -1, 0)
		if _, ok := allowedIMMMonths[returnedDate.Month()]; ok {
			return returnedDate
		}
	}
}
