package types

import "time"

// RuntimeMetric is metric of specific deployment
type RuntimeMetric struct {
	Duration   time.Duration
	NumRequest int
	NumSuccess int
}

type RequestMetric struct {
	RequestID string
	Status    int
	Duration  time.Duration
}

func CreateRequestMetric(id string, status int, duration time.Duration) RequestMetric {
	return RequestMetric{
		RequestID: id,
		Status:    status,
		Duration:  duration,
	}
}
