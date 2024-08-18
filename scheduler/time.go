package scheduler

import "time"

func futureTime(timestamp *time.Time, delay *string) time.Time {
	if timestamp != nil {
		return *timestamp
	}

	parsedDelay, err := time.ParseDuration(*delay)
	if err != nil {
		return time.Now()
	}

	return time.Now().Add(parsedDelay)
}
