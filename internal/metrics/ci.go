package metrics

import (
	"math"
	"sort"
)

// z95 is the two-sided 95% normal quantile.
const z95 = 1.959963984540054

// bootstrapIterations is the resample count of the bootstrap interval.
const bootstrapIterations = 1000

// bootstrapSeed keeps the bootstrap interval reproducible. A run that
// reported a different interval for the same trials on every call would not
// be a measurement.
const bootstrapSeed = 20261022

// wilsonInterval computes the Wilson score interval for a binomial
// proportion.
//
// It is the default because it stays inside [0, 1] and stays sensible at
// small sample sizes, where the normal approximation does not.
func wilsonInterval(successes, n float64) Interval {
	if n <= 0 {
		return Interval{Method: "wilson", Requested: "wilson", Basis: "pooled", Level: 0.95}
	}
	p := successes / n
	z := z95
	denom := 1 + z*z/n
	center := (p + z*z/(2*n)) / denom
	margin := z * math.Sqrt(p*(1-p)/n+z*z/(4*n*n)) / denom
	return Interval{
		Lower:     math.Max(0, center-margin),
		Upper:     math.Min(1, center+margin),
		Method:    "wilson",
		Requested: "wilson",
		Basis:     "pooled",
		Level:     0.95,
	}
}

// NormalInterval computes the normal-approximation interval over a sample
// of per-unit values.
//
// It is exported because a weighted total over several cells is itself an
// average of per-cell values, and the report needs the same interval the
// metric layer would compute.
func NormalInterval(values []float64) Interval { return normalInterval(values) }

// normalInterval computes the normal-approximation interval over a sample
// of per-unit values.
func normalInterval(values []float64) Interval {
	if len(values) == 0 {
		return Interval{Method: "normal", Requested: "normal", Basis: "task_values", Level: 0.95}
	}
	mean := meanOf(values)
	if len(values) == 1 {
		return Interval{Lower: mean, Upper: mean, Method: "normal", Requested: "normal", Basis: "task_values", Level: 0.95}
	}
	variance := varianceOf(values, mean)
	se := math.Sqrt(variance / float64(len(values)))
	return Interval{
		Lower:     mean - z95*se,
		Upper:     mean + z95*se,
		Method:    "normal",
		Requested: "normal",
		Basis:     "task_values",
		Level:     0.95,
	}
}

// bootstrapInterval computes the percentile bootstrap interval over a sample
// of per-unit values.
func bootstrapInterval(values []float64) Interval {
	if len(values) == 0 {
		return Interval{Method: "bootstrap", Requested: "bootstrap", Basis: "task_values", Level: 0.95}
	}
	if len(values) == 1 {
		return Interval{Lower: values[0], Upper: values[0], Method: "bootstrap", Requested: "bootstrap", Basis: "task_values", Level: 0.95}
	}
	rng := newLCG(bootstrapSeed)
	means := make([]float64, 0, bootstrapIterations)
	for i := 0; i < bootstrapIterations; i++ {
		sum := 0.0
		for j := 0; j < len(values); j++ {
			sum += values[rng.intn(len(values))]
		}
		means = append(means, sum/float64(len(values)))
	}
	sort.Float64s(means)
	lower := percentile(means, 0.025)
	upper := percentile(means, 0.975)
	return Interval{Lower: lower, Upper: upper, Method: "bootstrap", Requested: "bootstrap", Basis: "task_values", Level: 0.95}
}

// ciProportion computes the interval of a metric whose unit of observation
// is a Bernoulli outcome, such as the share of trials that passed.
//
// The bootstrap resamples the underlying outcomes, so it is meaningful for
// this shape; the normal approximation uses the pooled variance.
func ciProportion(method string, successes, n float64) *Interval {
	if n <= 0 {
		return nil
	}
	switch method {
	case "wilson":
		iv := wilsonInterval(successes, n)
		return &iv
	case "normal":
		p := successes / n
		se := math.Sqrt(p * (1 - p) / n)
		return &Interval{Lower: p - z95*se, Upper: p + z95*se, Method: "normal", Requested: "normal", Basis: "pooled", Level: 0.95}
	case "bootstrap":
		outcomes := make([]float64, 0, int(n))
		for i := 0; i < int(successes); i++ {
			outcomes = append(outcomes, 1)
		}
		for i := 0; i < int(n-successes); i++ {
			outcomes = append(outcomes, 0)
		}
		iv := bootstrapInterval(outcomes)
		iv.Basis = "pooled"
		return &iv
	}
	iv := wilsonInterval(successes, n)
	return &iv
}

// ciValues computes the interval of a metric whose unit of observation is a
// rate, such as the per-task pass@k.
//
// The Wilson interval is defined only for a Bernoulli proportion, so for
// this shape the implementation falls back to the normal approximation and
// records the basis it used. The basis is shown in the report so a reader
// knows which interval they are looking at.
func ciValues(method string, values []float64) *Interval {
	if len(values) == 0 {
		return nil
	}
	switch method {
	case "bootstrap":
		iv := bootstrapInterval(values)
		return &iv
	case "normal":
		iv := normalInterval(values)
		return &iv
	default:
		iv := normalInterval(values)
		iv.Requested = method
		iv.Basis = "task_values"
		return &iv
	}
}

func meanOf(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func varianceOf(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		d := v - mean
		sum += d * d
	}
	return sum / float64(len(values)-1)
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := p * float64(len(sorted)-1)
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo == hi {
		return sorted[lo]
	}
	frac := idx - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

// lcg is a small deterministic pseudo-random generator.
//
// math/rand's global source is not guaranteed stable across Go releases, and
// a bootstrap interval that changes between runs of the same binary is
// worse than no interval. The generator is seeded explicitly for that
// reason.
type lcg struct {
	state uint64
}

func newLCG(seed int64) *lcg {
	return &lcg{state: uint64(seed)*6364136223846793005 + 1442695040888963407}
}

func (l *lcg) next() uint64 {
	l.state = l.state*6364136223846793005 + 1442695040888963407
	return l.state >> 11
}

func (l *lcg) intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(l.next() % uint64(n))
}
