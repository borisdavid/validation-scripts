package integration

import "math"

// integrator is a structure providing routines for numerical integration.
type integrator struct {
	tolerance float64
	maxDepth  int
}

// Option allows to customize the integrator settings.
type Option func(*integrator)

// WithTolerance defines the integration error tolerance.
// The input parameter is supposed to be strictly positive.
func WithTolerance(tolerance float64) Option {
	return func(i *integrator) {
		i.tolerance = tolerance
	}
}

// WithMaxDepth defines the maximum number of interval sudivisions.
// The input parameter is supposed to be strictly positive.
func WithMaxDepth(maxDepth int) Option {
	return func(i *integrator) {
		i.maxDepth = maxDepth
	}
}

func newDefaultIntegrator() *integrator {
	return &integrator{
		tolerance: 1e-10,
		maxDepth:  10,
	}
}

var defaultIntegrator = newDefaultIntegrator()

// newIntegrator builds an integrator with given settings.
func newIntegrator(opts ...Option) *integrator {
	if len(opts) == 0 {
		return defaultIntegrator
	}

	integrator := newDefaultIntegrator()
	for _, opt := range opts {
		opt(integrator)
	}

	return integrator
}

// Integrate computes the integral of f(x) for x in (a,b)
// The parameters of the Integrator must always be set and correct before calling Integrate
// If a > b, the extremes of the integration interval are exchanged
// If the integral bounds are infinite, suitable changes of variable are used.
func Integrate(f Func, a, b float64, opts ...Option) float64 {
	if a > b {
		return -Integrate(f, b, a)
	}

	// Degenerate integration interval
	if math.Abs(b-a) < 1e-15 {
		return 0.0
	}

	i := newIntegrator(opts...)

	switch {
	case math.IsInf(a, -1) && math.IsInf(b, 1): // (-Inf,Inf)
		// u(t) = t/(1-t^2)
		g := func(x float64) float64 {
			v := (1.0 - x*x)
			return f(x/v) * (1.0 + x*x) / (v * v)
		}

		return i.integrate(g, -1.0, 1.0)

	case math.IsInf(a, -1): // (-Inf,b)
		// u(t) = b - (1-t)/t
		g := func(x float64) float64 {
			return f(b-(1.0-x)/x) / (x * x)
		}

		return i.integrate(g, 0.0, 1.0)

	case math.IsInf(b, 1): // (a,Inf)
		// u(t) = a + t/(1-t)
		g := func(x float64) float64 {
			v := 1.0 - x
			return f(a+x/v) / (v * v)
		}

		return i.integrate(g, 0.0, 1.0)

	default:
		return i.integrate(f, a, b)
	}
}

func (i integrator) integrate(f Func, a, b float64) float64 {
	return adaptiveGaussKronrod(f, a, b, i.tolerance, i.maxDepth)
}
