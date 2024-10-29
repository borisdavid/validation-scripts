package integration

import "math"

// gaussKronrod computes the integral of f in (leftBound,rightBound) and the corresponding error estimate.
func gaussKronrod(f Func, i *interval) (float64, float64) {
	length, mid := i.length(), i.midPoint()
	scale := 0.5 * length

	var gaussValue float64
	var kronrodValueA float64

	incrementA := func(v, wg, wgk float64) {
		fValue := f(scale*v + mid)
		gaussValue += wg * fValue
		kronrodValueA += wgk * fValue
	}

	incrementA(referenceGaussNodes1, referenceGaussWeights1, referenceGaussKronrodWeights1)
	incrementA(referenceGaussNodes2, referenceGaussWeights2, referenceGaussKronrodWeights2)
	incrementA(referenceGaussNodes3, referenceGaussWeights3, referenceGaussKronrodWeights3)
	incrementA(referenceGaussNodes4, referenceGaussWeights4, referenceGaussKronrodWeights4)
	incrementA(referenceGaussNodes5, referenceGaussWeights5, referenceGaussKronrodWeights5)
	incrementA(referenceGaussNodes6, referenceGaussWeights6, referenceGaussKronrodWeights6)
	incrementA(referenceGaussNodes7, referenceGaussWeights7, referenceGaussKronrodWeights7)

	var kronrodValueB float64
	incrementB := func(v, w float64) {
		kronrodValueB += w * f(scale*v+mid)
	}

	incrementB(referenceKronrodNodes1, referenceKronrodWeights1)
	incrementB(referenceKronrodNodes2, referenceKronrodWeights2)
	incrementB(referenceKronrodNodes3, referenceKronrodWeights3)
	incrementB(referenceKronrodNodes4, referenceKronrodWeights4)
	incrementB(referenceKronrodNodes5, referenceKronrodWeights5)
	incrementB(referenceKronrodNodes6, referenceKronrodWeights6)
	incrementB(referenceKronrodNodes7, referenceKronrodWeights7)
	incrementB(referenceKronrodNodes8, referenceKronrodWeights8)

	var (
		gaussIntegral   = scale * gaussValue
		kronrodIntegral = scale * (kronrodValueA + kronrodValueB)
	)

	return kronrodIntegral, math.Abs(gaussIntegral - kronrodIntegral)
}

// adaptiveGaussKronrod computes the integral of f in (leftBound,rightBound) up to a given tolerance.
func adaptiveGaussKronrod(f Func, leftBound, rightBound, tolerance float64, maxDepth int) float64 {
	originalInterval := &interval{a: leftBound, b: rightBound}

	intervalSize := originalInterval.length()

	criterion := convergenceCriterion{
		minIntervalSize:    1.5 * intervalSize * math.Pow(2, -float64(maxDepth)),
		targetErrorDensity: tolerance / intervalSize,
	}

	// Compute the value of the integral adaptively
	integral := 0.0

	originalIntegral, subError := gaussKronrod(f, originalInterval)
	if criterion.isMet(intervalSize, subError) {
		return originalIntegral
	}

	leftSubIntervals := originalInterval.halve()
	for len(leftSubIntervals) > 0 {
		subInterval := leftSubIntervals[0]
		leftSubIntervals = leftSubIntervals[1:]

		subIntegral, subError := gaussKronrod(f, subInterval)
		if subIntervalSize := subInterval.length(); criterion.isMet(subIntervalSize, subError) {
			integral += subIntegral
			continue
		}

		leftSubIntervals = append(leftSubIntervals, subInterval.halve()...)
	}

	return integral
}

type convergenceCriterion struct {
	minIntervalSize    float64
	targetErrorDensity float64
}

func (c convergenceCriterion) isMet(intervalSize, integrationErr float64) bool {
	return intervalSize < c.minIntervalSize || integrationErr < c.targetErrorDensity*intervalSize
}
