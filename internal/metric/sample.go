package metric

import "time"

type Group string

const (
	Canary  Group = "canary"
	Control Group = "control"
)

type Sample struct {
	ExperimentID string
	DeviceID     string
	MetricName   string
	Value        float64
	Timestamp    time.Time
	Group        Group
}

type Summary struct {
	Mean    float64   `json:"mean"`
	Stddev  float64   `json:"stddev"`
	N       int       `json:"n"`
	Samples []float64 `json:"samples"`
}
