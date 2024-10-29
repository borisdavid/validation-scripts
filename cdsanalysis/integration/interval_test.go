package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_interval(t *testing.T) {
	t.Parallel()

	const tol = 1e-15

	testIntervals := []struct {
		i        interval
		a        float64
		b        float64
		length   float64
		midpoint float64
	}{
		{
			interval{a: 0.0, b: 1.0},
			0.0,
			1.0,
			1.0,
			0.5,
		},
		{
			interval{a: -1.0, b: 2.0},
			-1.0,
			2.0,
			3.0,
			0.5,
		},
		{
			interval{a: -51.0, b: 32.0},
			-51.0,
			32.0,
			83.0,
			-9.5,
		},
	}
	for _, tc := range testIntervals {
		assert.InDelta(t, tc.a, tc.i.a, tol)
		assert.InDelta(t, tc.b, tc.i.b, tol)
		assert.InEpsilon(t, tc.length, tc.i.length(), tol)
		assert.InEpsilon(t, tc.midpoint, tc.i.midPoint(), tol)
	}
}

func Test_interval_halve(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		input    *interval
		expected []*interval
	}{
		{
			&interval{a: 0.0, b: 1.0},
			[]*interval{
				{a: 0.0, b: 0.5},
				{a: 0.5, b: 1.0},
			},
		},
		{
			&interval{a: -1.0, b: 1.0},
			[]*interval{
				{a: -1.0, b: 0.0},
				{a: 0.0, b: 1.0},
			},
		},
	} {
		assert.Equal(t, tc.expected, tc.input.halve())
	}
}
