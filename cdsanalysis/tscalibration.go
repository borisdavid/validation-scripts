package main

import (
	"cdsanalysis/optimization"
	"fmt"
	"math"
)

type pricer interface {
	PriceError(creditRate TermStructure) float64
}

type objectiveFunction struct {
	parametrization          ParametrizedTermStructure
	CDSs                     []CDSAsset
	longTermLow              float64
	longTermHigh             float64
	longTermWeight           float64
	regularizationTermWeight float64
}

func (o *objectiveFunction) Value(parameters []float64) float64 {
	// Build a term structure
	ts, _ := o.parametrization.Evaluate(parameters)
	totalError := priceCDSSum(o.CDSs, ts)

	Nbonds := float64(len(o.CDSs))

	// Long term value
	longTermMaturity := 100.0
	ltv := ts.Value(longTermMaturity)

	shortTermMaturity := 0.1
	stv := ts.Value(shortTermMaturity)

	// Scaling for the penalization
	scaling := math.Sqrt(Nbonds)

	// Flat term penalty
	spread := stv - ltv
	totalError += spread * spread * scaling * o.regularizationTermWeight

	// Long term penalty
	if ltv < o.longTermLow {
		diff := o.longTermLow - ltv
		totalError += diff * diff * scaling * o.longTermWeight
	}

	if ltv > o.longTermHigh {
		diff := ltv - o.longTermHigh
		totalError += diff * diff * scaling * o.longTermWeight
	}

	return totalError
}

func (o *objectiveFunction) Gradient(parameters []float64) []float64 {
	// Use finite differences
	ndim := len(parameters)
	grad := make([]float64, ndim)

	epsilon := 0.00001 // = 0.1 bps

	parametersDelta := make([]float64, ndim)
	copy(parametersDelta, parameters)

	for i := range ndim {
		pplus := math.Min(maxRate, parameters[i]+epsilon)
		parametersDelta[i] = pplus
		vplus := o.Value(parametersDelta)

		pminus := math.Max(minRate, parameters[i]-epsilon)
		parametersDelta[i] = pminus
		vminus := o.Value(parametersDelta)

		grad[i] = (vplus - vminus) / (pplus - pminus)
		parametersDelta[i] = parameters[i]
	}

	return grad
}

// Configuration is the structure gathering all the parameters
// used for the extraction of the curves.
type Configuration struct {
	// The parameterization to be used for the curves
	Parametrization ParametrizedTermStructure
	// The lookback feature helps stabilizing the curve
	// by re-using the price from N days before for a bond
	// that has no price, with a decreased weight in the optimization
	// corresponding to weight^N , with N lower or equal to the max.
	LookbackWeight float64
	LookbackMax    int
	// Minimum time to maturity that a bond should have to be used.
	MinExpectedMaturity float64
	// The configuration for the objective function
	ObjectiveFunction ObjectiveConfiguration
	// The min difference to be categorized as suspicious
	SuspectRepricingTolerance float64
	// The min number of suspect values before a marketdata is ejected
	SuspectMinOccurence int
	// The minimum rate of suspect / total values to be ejected
	SuspectMinRate float64
	// If false, the extraction will remove suspect assets.
	KeepSuspects bool
}

func DefaultConfiguration() Configuration {
	conf := Configuration{
		Parametrization:           ParametrizedFlatTermStructure{},
		LookbackWeight:            0.8825,
		LookbackMax:               6,
		MinExpectedMaturity:       1.0 / 12.0,
		SuspectRepricingTolerance: 0.03,
		SuspectMinOccurence:       10,
		SuspectMinRate:            0.5,
		ObjectiveFunction: ObjectiveConfiguration{
			LongTermLow:          0.05,
			LongTermHigh:         0.15,
			LongTermWeight:       0.0,
			RegularizationWeight: 0.0,
		},
	}

	return conf
}

// ObjectiveConfiguration is the structure gathering all the parameters
// used for the objective function.
type ObjectiveConfiguration struct {
	// The long term penalty pushed the long term part of the curve to
	// pass in the interval defined by the low and high values. The
	// weight is the weight used for this term in the optimization and
	// helps balancing it w.r. to the other terms.
	LongTermLow    float64
	LongTermHigh   float64
	LongTermWeight float64
	// The regularization term help in having a well conditioned problem
	// when there are not enough assets or when they do not contain enough
	// information for the term structure of the curve.
	RegularizationWeight float64
}

func createObjectiveFunction(
	parametrization ParametrizedTermStructure,
	cds []CDSAsset,
	config ObjectiveConfiguration,
) objectiveFunction {
	return objectiveFunction{parametrization, cds, config.LongTermLow, config.LongTermHigh, config.LongTermWeight, config.RegularizationWeight}
}

type extractor struct {
	configuration Configuration
}

var (
	minRate = 0.0
	maxRate = 3.0
)

// This function might change the content of the slices
// "pricers" and "weights".
func (e *extractor) extractCurve(
	parametrization ParametrizedTermStructure,
	cds []CDSAsset,
) (*TermStructure, error) {
	parameterSpaceDim := parametrization.Dimension()

	if len(cds) < 1 {
		// clarify the contract on these cases
		return nil, nil
	}

	obj := createObjectiveFunction(parametrization, cds, e.configuration.ObjectiveFunction)

	settings := optimization.NewSettings(parameterSpaceDim * 2)
	solver := optimization.NewSolver(settings)
	problem := optimization.BoxedProblem{
		LowerBounds:  createArrayWithValue(parameterSpaceDim, minRate),
		UpperBounds:  createArrayWithValue(parameterSpaceDim, maxRate),
		InitialGuess: createArrayWithValue(parameterSpaceDim, 0.05), // TODO: not optimal
		Objective:    &obj,
	}

	optimalparameters, err := solver.Minimize(problem)
	if err != nil {
		return nil, fmt.Errorf("curve optimization failed: %w", err)
	}

	curve, err := parametrization.Evaluate(optimalparameters)
	if err != nil {
		return nil, fmt.Errorf("curve optimization yielded an inadmissible solution: %w", err)
	}

	return &curve, nil
}

func createArrayWithValue(dim int, value float64) []float64 {
	a := make([]float64, dim)
	for i := range dim {
		a[i] = value
	}

	return a
}
