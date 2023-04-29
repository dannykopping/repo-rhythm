package metrics

import (
	"math"

	"facette.io/natsort"
	"github.com/prometheus/client_golang/prometheus"
)

// Distribution is a prometheus.Collector that collects observations and buckets them.
// It differs from a histogram in two key ways:
// 1. It does not collect observations over time, only a snapshot of the current state.
// 2. Its buckets are strings, not float64s, so that they are easier to diagram.
type Distribution interface {
	prometheus.Collector

	// Observe adds a single observation to the distribution in the appropriate bucket.
	Observe(float64)
	Reset()
}

type DistributionOpts prometheus.Opts

const infBucket = "+Inf"

// NewDistribution creates
func NewDistribution(opts DistributionOpts, buckets map[string]float64) Distribution {
	dist := &distribution{
		buckets: make(map[string]*bucket, len(buckets)),
	}

	for id, max := range buckets {
		dist.buckets[id] = &bucket{
			gauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name:        opts.Name,
				Help:        opts.Help,
				ConstLabels: opts.ConstLabels,
			}, []string{"bucket"}),
			bucket: id,
			max:    max,
		}
	}

	// add infinity bucket to catch all observations above the highest bucket
	dist.buckets[infBucket] = &bucket{
		gauge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:        opts.Name,
			Help:        opts.Help,
			ConstLabels: opts.ConstLabels,
		}, []string{"bucket"}),
		bucket: infBucket,
		max:    math.Inf(1),
	}

	return dist
}

type distribution struct {
	buckets map[string]*bucket
}

type bucket struct {
	gauge  *prometheus.GaugeVec
	bucket string
	max    float64
}

func (d *distribution) Reset() {
	for _, d := range d.buckets {
		d.gauge.WithLabelValues(d.bucket).Set(0)
	}
}

func (d *distribution) Describe(descs chan<- *prometheus.Desc) {
	for _, d := range d.buckets {
		d.gauge.Describe(descs)
	}
}

func (d *distribution) Collect(metrics chan<- prometheus.Metric) {
	for _, d := range d.buckets {
		d.gauge.Collect(metrics)
	}
}

func (d *distribution) Observe(v float64) {
	var keys []string
	for k := range d.buckets {
		keys = append(keys, k)
	}
	natsort.Sort(keys)

	var prev float64 = 0
	for _, bucket := range keys {
		max := d.buckets[bucket].max
		// find a bucket whose upper bound is less than or equal to v
		if v > prev && v <= max {
			d.buckets[bucket].gauge.WithLabelValues(bucket).Inc()
		}
		prev = max
	}
}
