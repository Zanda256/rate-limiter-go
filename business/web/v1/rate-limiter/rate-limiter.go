package rate_limiter

// Write the supported types of rate limiting and their implementations in this file
// Start with Leaky bucket

type Algo int

const (
	LeakyBucket = iota
	TokenBucket
	FixedWindow
	SlidingLog
	SlidingWindow
)

var supportedAlgos = map[string]int{
	"LeakyBucket":   LeakyBucket,
	"TokenBucket":   TokenBucket,
	"FixedWindow":   FixedWindow,
	"SlidingLog":    SlidingLog,
	"SlidingWindow": SlidingWindow,
}
