package integration

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newIntegrator(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		opts     []Option
		expected *integrator
	}{
		"default": {
			nil,
			&integrator{
				tolerance: 1e-10,
				maxDepth:  10,
			},
		},
		"custom": {
			[]Option{
				WithTolerance(1.0e-8),
				WithMaxDepth(5),
			},
			&integrator{
				tolerance: 1e-8,
				maxDepth:  5,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expected, newIntegrator(tc.opts...))
		})
	}
}

func Test_Integrator_Integrate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		f        func(float64) float64
		expected float64
	}{
		{
			"quadratic",
			func(x float64) float64 { return x * x },
			1.0 / 3.0,
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
	}

	for _, tc := range []struct {
		name string
		opts []Option
	}{
		{
			"default",
			nil,
		},
		{
			"custom",
			[]Option{
				WithTolerance(1.0e-8),
				WithMaxDepth(5),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, tt := range testCases {
				tt := tt

				t.Run(tt.name, func(t *testing.T) {
					t.Parallel()

					assert.InEpsilon(t, tt.expected, Integrate(tt.f, 0.0, 1.0), tol)
				})
			}
		})
	}
}

func Test_Integrator_InvertIntegralBounds(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		a        float64
		b        float64
		f        func(float64) float64
		expected float64
	}{
		"f(x)=x^2/in-(1,0)": {
			1.0,
			0.0,
			func(x float64) float64 { return x },
			-0.5,
		},
		"f(x)=x^5-x^3/in-(3,-2)": {
			3.0,
			-2.0,
			func(x float64) float64 { return math.Pow(x, 5) - math.Pow(x, 3) },
			-1135.0 / 12.0,
		},
		"f(x)=1/x/in-(-1,-2)": {
			-1.0,
			-2.0,
			func(x float64) float64 { return 1.0 / x },
			math.Log(2),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.InEpsilon(t, tc.expected, Integrate(tc.f, tc.a, tc.b), tol)
		})
	}
}

func Test_Integrator_EqualIntegralBounds(t *testing.T) {
	t.Parallel()

	f, a, b := func(x float64) float64 { return x }, 0.0, 0.0

	assert.InDelta(t, 0.0, Integrate(f, a, b), 1e-15)
}

func Test_Integrator_InfiniteBounds(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		a        float64
		b        float64
		f        func(float64) float64
		expected float64
	}{
		"(-inf,+inf)": {
			math.Inf(-1),
			math.Inf(1),
			func(x float64) float64 { return math.Exp(-0.5 * x * x) },
			math.Sqrt(2.0 * math.Pi),
		},
		"(-inf,0)": {
			math.Inf(-1),
			0.0,
			func(x float64) float64 { return 1.0 / (1.0 + x*x) },
			0.5 * math.Pi,
		},
		"(0,+inf)": {
			0.0,
			math.Inf(1),
			func(x float64) float64 { return 10.0 * x * math.Exp(-10.0*x) },
			0.1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.InDelta(t, tc.expected, Integrate(tc.f, tc.a, tc.b), 1e-8)
		})
	}
}
