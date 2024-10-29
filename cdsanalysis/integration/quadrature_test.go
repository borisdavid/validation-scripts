package integration

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

const tol = 1.0e-14

// create a list of test functions.
var testCases = []struct {
	name     string
	f        Func
	expected float64
}{
	{
		"constant",
		func(float64) float64 { return 1.0 },
		1.0,
	},
	{
		"linear",
		func(x float64) float64 { return x },
		0.5,
	},
	{
		"affine",
		func(x float64) float64 { return 2.0*x + 3.0 },
		4.0,
	},
	{
		"quadratic",
		func(x float64) float64 { return x * x },
		1.0 / 3.0,
	},
	{
		"cubic",
		func(x float64) float64 { return x * x * x },
		0.25,
	},
	{
		"polynomial",
		func(x float64) float64 {
			return 5.0*math.Pow(x, 5) - 3.0*math.Pow(x, 4) + 2.0*math.Pow(x, 16) - math.Pow(x, 9)
		},
		5.0/6.0 - 0.6 + 2.0/17.0 - 0.1,
	},
	{
		"exponential",
		func(x float64) float64 {
			return math.Exp(-2.0 * x)
		},
		(1.0 - math.Exp(-2.0)) / 2.0,
	},
	{
		"trigonometric",
		func(x float64) float64 {
			return math.Sin(math.Pi*x) - math.Sin(4.0*math.Pi*x)
		},
		2.0 / math.Pi,
	},
}

func Test_GaussKronrod(t *testing.T) {
	t.Parallel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			If, _ := gaussKronrod(tc.f, &interval{a: 0.0, b: 1.0})
			assert.InEpsilon(t, tc.expected, If, tol)
		})
	}
}

func Test_AdaptiveGaussKronrod(t *testing.T) {
	t.Parallel()

	for _, tc := range append(testCases, []struct {
		name     string
		f        Func
		expected float64
	}{
		{
			"stiff exponential",
			func(x float64) float64 {
				return 10.0 * math.Exp(-11.0*x)
			},
			10.0 / 11.0 * (1.0 - math.Exp(-11.0)),
		},
		{
			"super-stiff exponential",
			func(x float64) float64 {
				return 1000.0 * math.Exp(-1001.0*x)
			},
			1000.0 / 1001.0 * (1.0 - math.Exp(-1001.0)),
		},
		{
			"super-stiff exponential times linear term",
			func(x float64) float64 {
				return 1000.0 * x * math.Exp(-1001.0*x)
			},
			1000.0 / 1001.0 * (1.0/1001.0 - (1.0+1.0/1001.0)*math.Exp(-1001.0)),
		},
	}...) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			If := adaptiveGaussKronrod(tc.f, 0.0, 1.0, 1.0e-6, 10)
			assert.InEpsilon(t, tc.expected, If, tol)
		})
	}
}

func Benchmark_gaussKronrod(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = gaussKronrod(func(x float64) float64 { return x }, &interval{a: 0.0, b: 1.0})
	}
}
