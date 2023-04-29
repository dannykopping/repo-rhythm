package rhythm

import "time"

type Config struct {
	Owner           string
	Repo            string
	TimeoutDuration time.Duration
	TickInterval    time.Duration
}
