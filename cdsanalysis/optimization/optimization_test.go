package optimization

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Quadratic struct {
	coefficients []float64
}

func (q Quadratic) Value(x []float64) float64 {
	sum := 0.0

	dim := len(q.coefficients)

	for i := range dim {
		sum += (x[i] - q.coefficients[i]) * (x[i] - q.coefficients[i])
	}

	return sum
}

func (q Quadratic) Gradient(x []float64) []float64 {
	dim := len(q.coefficients)
	g := make([]float64, dim)

	for i := range dim {
		g[i] = 2*x[i] - 2.0*q.coefficients[i]
	}

	return g
}

func Test_OptimizerQuadraticProblem(t *testing.T) {
	// Will try to optimize in the [-1 1]x[-2 2]x[-3 3] ... box
	// the quadratic form given by the test case
	testcases := []struct {
		coefficients []float64
		optimum      []float64
	}{
		// 2D
		{[]float64{0.2, 0.2}, []float64{0.2, 0.2}},
		{[]float64{0.2, 4.0}, []float64{0.2, 2.0}},
		{[]float64{2.0, -0.2}, []float64{1.0, -0.2}},
		{[]float64{10.0, 3.0}, []float64{1.0, 2.0}},
		{[]float64{-2.0, 3.0}, []float64{-1.0, 2.0}},

		// 5D
		{[]float64{0.5, -3.0, 4.0, -1.0, 6.0}, []float64{0.5, -2.0, 3.0, -1.0, 5.0}},
	}

	for _, testcase := range testcases {
		dim := len(testcase.optimum)

		q := Quadratic{
			coefficients: testcase.coefficients,
		}

		lowerBound := make([]float64, dim)
		upperBound := make([]float64, dim)
		initial := make([]float64, dim)

		for i := range dim {
			lowerBound[i] = float64(-i - 1)
			upperBound[i] = float64(i + 1)
			initial[i] = 0.0
		}

		problem := BoxedProblem{
			lowerBound,
			upperBound,
			initial,
			q,
		}

		settings := NewSettings(20)
		solver := NewSolver(settings)

		x, err := solver.Minimize(problem)

		assert.Empty(t, err)

		for i := range dim {
			assert.InDelta(t, x[i], testcase.optimum[i], 5e-3)
		}
	}
}
