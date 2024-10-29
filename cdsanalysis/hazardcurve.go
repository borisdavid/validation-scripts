package main

import (
	"fmt"
	"math"

	interp "github.com/edgelaboratories/go-libraries/interpolator"
)

// A term structure is simply a function of time.
type TermStructure interface {
	Value(yf float64) float64
	Parameters() []float64
	Derivative(yf float64) float64
}

// A parameterized term structure.
type ParametrizedTermStructure interface {
	Dimension() int
	Evaluate(params []float64) (TermStructure, error)
	Fit(points map[float64]float64) (TermStructure, error)
}

// FLAT SPREAD

const flatSpreadNbParameters = 1

// Flat term structure.
type FlatTermStructure struct {
	Val float64
}

func (ts FlatTermStructure) Value(float64) float64 {
	return ts.Val
}

func (ts FlatTermStructure) Parameters() []float64 {
	return []float64{ts.Val}
}

func (ts FlatTermStructure) Derivative(float64) float64 {
	return 0.0
}

type ParametrizedFlatTermStructure struct{}

func (pts ParametrizedFlatTermStructure) Dimension() int {
	return flatSpreadNbParameters
}

func (pts ParametrizedFlatTermStructure) Evaluate(params []float64) (TermStructure, error) {
	if len(params) != flatSpreadNbParameters {
		return &FlatTermStructure{}, fmt.Errorf("cannot create a flat term structure, number of parameters mismatch: %d provided", len(params))
	}

	return &FlatTermStructure{params[0]}, nil
}

func (pts ParametrizedFlatTermStructure) Fit(points map[float64]float64) (TermStructure, error) {
	for _, v := range points {
		return &FlatTermStructure{v}, nil
	}

	return nil, fmt.Errorf("cannot fit a flat term structure with these %d points", len(points))
}

// Linear term structure.
type LinearThresholdTermStructure struct {
	xys interp.XYs
	*interp.PiecewiseLinearThreshold
}

var _ TermStructure = &LinearThresholdTermStructure{}

func NewLinearThresholdTermStructure(xys interp.XYs) (*LinearThresholdTermStructure, error) {
	interp, err := interp.NewPiecewiseLinearThreshold(xys)
	if err != nil {
		return nil, fmt.Errorf("could not create linear threhold interpolator: %w", err)
	}

	return &LinearThresholdTermStructure{
		xys:                      xys,
		PiecewiseLinearThreshold: interp,
	}, nil
}

func (ts LinearThresholdTermStructure) Parameters() []float64 {
	parameters := make([]float64, 0, len(ts.xys))
	for _, xy := range ts.xys {
		parameters = append(parameters, xy.X, xy.Y)
	}

	return parameters
}

func (ts LinearThresholdTermStructure) Derivative(yf float64) float64 {
	return ts.Gradient(yf)
}

func survivalProbabilityDensity(ts TermStructure, yf float64) float64 {
	return (ts.Value(yf) + ts.Derivative(yf)*yf) * math.Exp(-ts.Value(yf)*yf)
}

func survivalProbability(ts TermStructure, yf float64) float64 {
	return math.Exp(-ts.Value(yf) * yf)
}

/*
// LONG-SHORT NELSON-SIEGEL STYLE

const (
	longShortNbParameters     = 2
	longshortNSTransitionTime = 12.0
)

// Flat term structure.
type LongShortNS struct {
	Shortrate float64
	Longrate  float64
}

func (ts LongShortNS) Value(t float64) float64 {
	if t == 0.0 {
		return ts.Shortrate
	}

	tn := t / longshortNSTransitionTime
	gtn := (1 - math.Exp(-tn)) / tn // =1 when tn =0, =0 when tn =1

	return ts.Shortrate*gtn + ts.Longrate*(1-gtn)
}

func (ts LongShortNS) Parameters() []float64 {
	return []float64{ts.Shortrate, ts.Longrate}
}

// Parametrized Long-short Nelson-Siegel term structure.
type ParametrizedLongShortNS struct{}

func (pts ParametrizedLongShortNS) Dimension() int {
	return longShortNbParameters
}

func (pts ParametrizedLongShortNS) Evaluate(params []float64) (TermStructure, error) {
	if len(params) != longShortNbParameters {
		return &LongShortNS{}, fmt.Errorf("cannot create a long-short NS term structure, number of parameters mismatch: %d provided", len(params))
	}

	return &LongShortNS{Shortrate: params[0], Longrate: params[1]}, nil
}

func (pts ParametrizedLongShortNS) Fit(points map[float64]float64) (TermStructure, error) {
	// Find the min t and max t
	// (most stable solution)
	if len(points) < longShortNbParameters {
		return nil, fmt.Errorf("cannot fit LS Nelson-Siegel with these %d points", len(points))
	}

	mint, minv := findMinKey(points)
	maxt, maxv := findMaxKey(points)

	// Check that the data are ok: this does not pass if
	// - There are 0 or 1 points
	// - There are multiple points with all the same time
	if maxt <= mint {
		return nil, fmt.Errorf("cannot fit LS Nelson-Siegel with these %d points", len(points))
	}

	// Compute the coefficient called gtn in the computation
	// of the parametrization
	mintn := mint / longshortNSTransitionTime
	mingt := (1 - math.Exp(-mintn)) / mintn

	maxtn := maxt / longshortNSTransitionTime
	maxgt := (1 - math.Exp(-maxtn)) / maxtn

	// Now there are two linear equations
	// minv = s*mingt + l*(1-mingt)
	// maxv = s*maxgt + l*(1-maxgt)
	s, l := solveLinearSystem(mingt, 1.0-mingt, minv, maxgt, 1.0-maxgt, maxv)

	return &LongShortNS{Shortrate: s, Longrate: l}, nil
}

func findMinKey(points map[float64]float64) (float64, float64) {
	var minv, mint float64

	for t, v := range points {
		minv = v
		mint = t

		break
	}

	for t, v := range points {
		if t < mint {
			mint = t
			minv = v
		}
	}

	return mint, minv
}

func findMaxKey(points map[float64]float64) (float64, float64) {
	var maxv, maxt float64

	for t, v := range points {
		maxv = v
		maxt = t

		break
	}

	for t, v := range points {
		if t > maxt {
			maxt = t
			maxv = v
		}
	}

	return maxt, maxv
}

// Function to solve a 2x2 linear equation system
// a1 * x + b1 * y = c1
// a2 * x + b2 * y = c2
// using Cramer's formula.
func solveLinearSystem(a1 float64, b1 float64, c1 float64, a2 float64, b2 float64, c2 float64) (float64, float64) {
	x := (c1*b2 - b1*c2) / (a1*b2 - a2*b1)
	y := (a1*c2 - a2*c1) / (a1*b2 - a2*b1)

	return x, y
}
*/
