package main

import (
	"fmt"
	"math"

	"github.com/edgelaboratories/eve/pkg/marketdata"
	interp "github.com/edgelaboratories/go-libraries/interpolator"
)

// InterestRateCurveRepresentation is the model representation of an InterestRateCurve.
type InterestRateCurveRepresentation struct {
	// Data represents the term structure of interest spot rates
	Data marketdata.TermStructure
	// Model is the underlying model used to interpolate on interest spot rates
	Model interestRateCurveModel
}

// Build calibrates the underlying model.
func (ir *InterestRateCurveRepresentation) Build() error {
	return ir.Model.calibrate(ir.Data)
}

// DiscountFactor returns the discount factor for a given year fraction.
func (ir *InterestRateCurveRepresentation) DiscountFactor(yf float64) float64 {
	return math.Exp(-ir.Spot(yf) * yf)
}

// Spot returns the spot rate for a given year fraction.
func (ir *InterestRateCurveRepresentation) Spot(yf float64) float64 {
	return ir.Model.value(yf)
}

// interpolator is an interface for numerical interpolations.
type interpolator interface {
	Value(x float64) float64
	Gradient(x float64) float64
}

type interestRateCurveModel interface {
	calibrate(data marketdata.TermStructure) error
	value(yf float64) float64
}

// PiecewiseLinearCurveModel is a piece-wise linear model for curves.
type PiecewiseLinearCurveModel struct {
	interpolator interpolator
}

func (m *PiecewiseLinearCurveModel) calibrate(data marketdata.TermStructure) error {
	dataPoints, err := convertTermStructureToRawData(data)
	if err != nil {
		return fmt.Errorf("could not convert the term structure data into raw data: %w", err)
	}
	interpolator, err := interp.NewPiecewiseLinear(dataPoints)
	if err != nil {
		return fmt.Errorf("could not create the piecewise linear interpolator: %w", err)
	}
	m.interpolator = interpolator
	return nil
}

func (m PiecewiseLinearCurveModel) value(yf float64) float64 {
	return m.interpolator.Value(yf)
}
