package simulation

import (
	"math/rand/v2"
)

// Stream generates n samples from a normal distribution around base
// with the given standard deviation.
func Stream(rng *rand.Rand, base, stddev float64, n int) []float64 {
	out := make([]float64, n)
	for i := range n {
		out[i] = base + rng.NormFloat64()*stddev
	}
	return out
}

// ShiftedStream generates n samples from a normal distribution with
// the mean shifted by delta. Simulates a regression or improvement.
func ShiftedStream(rng *rand.Rand, base, delta, stddev float64, n int) []float64 {
	return Stream(rng, base+delta, stddev, n)
}
