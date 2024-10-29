package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/edgelaboratories/eve/pkg/marketdata"
	"github.com/edgelaboratories/go-libraries/date"
)

const (
	overnighKey = "ON"
	weekKey     = "W"
	monthKey    = "M"
	yearKey     = "Y"

	daysPerWeek   = 7.0
	monthsPerYear = 12.0
	daysPerYear   = 365.0
)

type Tenor marketdata.Tenor

func (t Tenor) ToYearFraction() (float64, error) {
	return marketdata.Tenor(t).ToYearFraction()
}

func (t Tenor) ShiftDateByTenor(d date.Date) (date.Date, error) {
	switch {
	case t.isON():
		return d.AddDate(0, 0, 1), nil
	case t.isWeek():
		weeks, err := t.weekTenorToYearFraction()
		if err != nil {
			return date.Date{}, fmt.Errorf("could not convert the week tenor to year fraction: %w", err)
		}

		return d.AddDate(0, 0, int(weeks*daysPerWeek)), nil
	case t.isMonth():
		months, err := t.monthTenorToYearFraction()
		if err != nil {
			return date.Date{}, fmt.Errorf("could not convert the month tenor to year fraction: %w", err)
		}

		return d.AddDate(0, int(months), 0), nil
	case t.isYear():
		years, err := t.yearTenorToYearFraction()
		if err != nil {
			return date.Date{}, fmt.Errorf("could not convert the year tenor to year fraction: %w", err)
		}

		return d.AddDate(int(years), 0, 0), nil
	default:
		return date.Date{}, fmt.Errorf("could not parse the type of tenor %s", t)
	}
}

// isON tests whether the tenor is the ON tenor.
func (t Tenor) isON() bool {
	return strings.HasPrefix(string(t), overnighKey)
}

// isWeek tests whether the tenor is a week tenor.
func (t Tenor) isWeek() bool {
	return strings.HasPrefix(string(t), weekKey)
}

// isMonth tests whether the tenor is a month tenor.
func (t Tenor) isMonth() bool {
	return strings.HasPrefix(string(t), monthKey)
}

// isYear tests whether the tenor is a year tenor.
func (t Tenor) isYear() bool {
	return strings.HasPrefix(string(t), yearKey)
}

// weekTenorToYearFraction converts a week tenor to a year fraction.
func (t Tenor) weekTenorToYearFraction() (float64, error) {
	nbDays := strings.TrimPrefix(string(t), weekKey)
	d, err := strconv.ParseUint(nbDays, 10, 64)
	if err != nil {
		return 0.0, fmt.Errorf("failed to parse the number of weeks for tenor %s: %w", t, err)
	}
	return float64(d) * daysPerWeek / daysPerYear, nil
}

// monthTenorToYearFraction converts a month tenor to a year fraction.
func (t Tenor) monthTenorToYearFraction() (float64, error) {
	nbMonths := strings.TrimPrefix(string(t), monthKey)
	m, err := strconv.ParseUint(nbMonths, 10, 64)
	if err != nil {
		return 0.0, fmt.Errorf("failed to parse the number of months for tenor %s: %w", t, err)
	}
	return float64(m), nil
}

// yearTenorToYearFraction converts a year tenor to a year fraction.
func (t Tenor) yearTenorToYearFraction() (float64, error) {
	nbYears := strings.TrimPrefix(string(t), yearKey)
	y, err := strconv.ParseUint(nbYears, 10, 64)
	if err != nil {
		return 0.0, fmt.Errorf("failed to parse the number of years for tenor %s: %w", t, err)
	}
	return float64(y), nil
}
