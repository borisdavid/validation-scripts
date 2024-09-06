package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_marketCapToHorizon(t *testing.T) {
	for _, marketCap := range []float64{0, 1, 100, 1000, 10000, 100000, 1000000, 10000000, 100000000, 1000000000, 10000000000, 100000000000, 1000000000000, 10000000000000, 100000000000000, 1000000000000000, 10000000000000000, 100000000000000000, 1000000000000000000, 10000000000000000000} {
		horizon := marketCapToHorizon(&marketCap)
		fmt.Println("marketCap:", marketCap, "horizon:", horizon)
	}

	assert.True(t, false)
}
