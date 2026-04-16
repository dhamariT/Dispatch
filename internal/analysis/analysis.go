// Package analysis is the pure statistical core: Welch's t-test,
// Cohen's d, and the regression verdict logic. It has no dependency
// on storage, HTTP, or any other Dispatch package. Anything that
// wants to compare a canary group to a control group calls Analyze.
package analysis

import "math"

const (
	minSampleSize = 5
	defaultAlpha  = 0.05
	minEffectSize = 0.5
)

type Verdict string

const (
	Regression       Verdict = "regression"
	Improvement      Verdict = "improvement"
	NoChange         Verdict = "no_change"
	InsufficientData Verdict = "insufficient_data"
)

// Direction tells the engine which way is "good" for a metric.
// HigherIsBetter: accuracy, uptime. LowerIsBetter: CPU, latency, memory.
type Direction int

const (
	HigherIsBetter Direction = iota
	LowerIsBetter
)

type Result struct {
	MetricName  string    `json:"metric_name"`
	Direction   Direction `json:"direction"`
	CanaryMean  float64   `json:"canary_mean"`
	CanarySD    float64   `json:"canary_sd"`
	CanaryN     int       `json:"canary_n"`
	ControlMean float64   `json:"control_mean"`
	ControlSD   float64   `json:"control_sd"`
	ControlN    int       `json:"control_n"`
	TStatistic  float64   `json:"t_statistic"`
	PValue      float64   `json:"p_value"`
	EffectSize  float64   `json:"effect_size"`
	Verdict     Verdict   `json:"verdict"`
}

// DefaultAlpha is exposed so callers that don't care about custom
// significance thresholds don't need to know the literal.
const DefaultAlpha = defaultAlpha

func Analyze(metricName string, dir Direction, canary, control []float64, alpha float64) Result {
	r := Result{MetricName: metricName, Direction: dir}

	r.CanaryN = len(canary)
	r.ControlN = len(control)

	if r.CanaryN < minSampleSize || r.ControlN < minSampleSize {
		r.Verdict = InsufficientData
		return r
	}

	r.CanaryMean, r.CanarySD = meanStddev(canary)
	r.ControlMean, r.ControlSD = meanStddev(control)

	if r.CanarySD == 0 && r.ControlSD == 0 {
		if r.CanaryMean == r.ControlMean {
			r.Verdict = NoChange
			return r
		}
		// Identical values within each group but different means.
		// Can't compute t-test with zero variance.
		r.Verdict = InsufficientData
		return r
	}

	r.TStatistic, r.PValue = welchTTest(r.CanaryMean, r.CanarySD, r.CanaryN, r.ControlMean, r.ControlSD, r.ControlN)
	r.EffectSize = cohenD(r.CanaryMean, r.CanarySD, r.CanaryN, r.ControlMean, r.ControlSD, r.ControlN)

	r.Verdict = decide(r.TStatistic, r.PValue, r.EffectSize, dir, alpha)
	return r
}

func meanStddev(xs []float64) (float64, float64) {
	n := float64(len(xs))
	var sum float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / n

	var ss float64
	for _, x := range xs {
		d := x - mean
		ss += d * d
	}
	// Sample standard deviation (Bessel's correction).
	sd := math.Sqrt(ss / (n - 1))
	return mean, sd
}

// welchTTest computes the t-statistic and two-tailed p-value for
// Welch's unequal variances t-test. Does not assume equal variance
// or equal sample sizes.
func welchTTest(m1, s1 float64, n1 int, m2, s2 float64, n2 int) (float64, float64) {
	nf1, nf2 := float64(n1), float64(n2)
	v1 := s1 * s1 / nf1
	v2 := s2 * s2 / nf2
	se := math.Sqrt(v1 + v2)
	if se == 0 {
		return 0, 1
	}

	t := (m1 - m2) / se

	// Welch-Satterthwaite degrees of freedom.
	num := (v1 + v2) * (v1 + v2)
	denom := (v1*v1)/(nf1-1) + (v2*v2)/(nf2-1)
	df := num / denom

	p := twoTailP(t, df)
	return t, p
}

// cohenD computes Cohen's d using pooled standard deviation.
func cohenD(m1, s1 float64, n1 int, m2, s2 float64, n2 int) float64 {
	nf1, nf2 := float64(n1), float64(n2)
	pooled := math.Sqrt(((nf1-1)*s1*s1 + (nf2-1)*s2*s2) / (nf1 + nf2 - 2))
	if pooled == 0 {
		return 0
	}
	return (m1 - m2) / pooled
}

// decide produces a verdict using metric direction to determine
// what counts as a regression. For HigherIsBetter (accuracy),
// canary < control is a regression. For LowerIsBetter (CPU, latency),
// canary > control is a regression. Requires both statistical
// significance AND meaningful effect size.
func decide(t, p, d float64, dir Direction, alpha float64) Verdict {
	if p > alpha {
		return NoChange
	}
	if math.Abs(d) < minEffectSize {
		return NoChange
	}
	regressed := (dir == HigherIsBetter && t < 0) || (dir == LowerIsBetter && t > 0)
	if regressed {
		return Regression
	}
	return Improvement
}

// twoTailP approximates the two-tailed p-value from a t-distribution
// using the regularized incomplete beta function. No external deps.
func twoTailP(t, df float64) float64 {
	x := df / (df + t*t)
	return regIncBeta(df/2, 0.5, x)
}

// regIncBeta computes the regularized incomplete beta function I_x(a,b)
// using a continued fraction expansion (Lentz's method). Accurate to
// ~1e-10 for the parameter ranges we care about.
func regIncBeta(a, b, x float64) float64 {
	if x <= 0 {
		return 0
	}
	if x >= 1 {
		return 1
	}

	// Use the symmetry relation when x > (a+1)/(a+b+2) for faster
	// convergence of the continued fraction.
	if x > (a+1)/(a+b+2) {
		return 1 - regIncBeta(b, a, 1-x)
	}

	lnBeta := lgamma(a) + lgamma(b) - lgamma(a+b)
	front := math.Exp(math.Log(x)*a+math.Log(1-x)*b-lnBeta) / a

	// Lentz's continued fraction.
	const maxIter = 200
	const epsilon = 1e-14
	const tiny = 1e-30

	f := tiny
	c := tiny
	d := 1.0

	for i := range maxIter {
		m := i / 2
		var num float64
		if i == 0 {
			num = 1.0
		} else if i%2 == 0 {
			mf := float64(m)
			num = mf * (b - mf) * x / ((a + 2*mf - 1) * (a + 2*mf))
		} else {
			mf := float64(m + 1)
			num = -(a + mf - 1) * (a + b + mf - 1) * x / ((a + 2*mf - 1) * (a + 2*mf - 2))
		}

		d = 1 + num*d
		if math.Abs(d) < tiny {
			d = tiny
		}
		c = 1 + num/c
		if math.Abs(c) < tiny {
			c = tiny
		}
		d = 1 / d
		delta := c * d
		f *= delta
		if math.Abs(delta-1) < epsilon {
			break
		}
	}

	return front * f
}

func lgamma(x float64) float64 {
	v, _ := math.Lgamma(x)
	return v
}
