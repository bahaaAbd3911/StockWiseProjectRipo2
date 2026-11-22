// Package ratelimit is a small in-memory token-bucket rate limiter.
//
// Every key (an IP address, an email, ...) has a bucket holding up to
// Rule.Requests tokens. A request takes one token; tokens flow back in at a
// steady pace so that the bucket is full again after Rule.Per. When the
// bucket is empty the request is refused, together with how long to wait.
//
// Example: Rule{Requests: 10, Per: time.Minute} allows a burst of 10 requests,
// then one more every 6 seconds.
//
// The state lives in this process's memory. That is fine for one server; if
// the API ever runs on several servers behind a load balancer, each one counts
// separately (limits become N times looser) - move the buckets to Redis then.
package ratelimit

import (
	"math"
	"sync"
	"time"
)

// Rule is "Requests per Per", e.g. 5 per 15 minutes.
type Rule struct {
	Requests int
	Per      time.Duration
}

type bucket struct {
	tokens float64
	last   time.Time
}

// Limiter tracks one bucket per key. A nil *Limiter allows everything, which
// is how rate limiting is switched off.
type Limiter struct {
	mu        sync.Mutex
	capacity  float64
	perSecond float64
	buckets   map[string]*bucket
	lastSweep time.Time
	now       func() time.Time
}

// New creates a limiter for a rule. A rule with no requests returns nil
// (= unlimited).
func New(r Rule) *Limiter {
	if r.Requests <= 0 || r.Per <= 0 {
		return nil
	}
	return &Limiter{
		capacity:  float64(r.Requests),
		perSecond: float64(r.Requests) / r.Per.Seconds(),
		buckets:   map[string]*bucket{},
		now:       time.Now,
	}
}

// Allow takes one token for key. If none is left it returns false and how long
// until the next token.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	return l.take(key, true)
}

// Check reports whether a request would be allowed, without taking a token.
// Used for "failed login" limits: check before, Fail only when it failed.
func (l *Limiter) Check(key string) (bool, time.Duration) {
	return l.take(key, false)
}

// Fail takes one token without checking (records a failed attempt).
func (l *Limiter) Fail(key string) {
	l.take(key, true)
}

// Reset forgets a key (e.g. after a successful login).
func (l *Limiter) Reset(key string) {
	if l == nil {
		return
	}
	l.mu.Lock()
	delete(l.buckets, key)
	l.mu.Unlock()
}

func (l *Limiter) take(key string, consume bool) (bool, time.Duration) {
	if l == nil {
		return true, 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.sweep(now)

	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: l.capacity, last: now}
		l.buckets[key] = b
	}
	// Refill for the time that passed since the last visit.
	b.tokens = math.Min(l.capacity, b.tokens+now.Sub(b.last).Seconds()*l.perSecond)
	b.last = now

	if b.tokens < 1 {
		wait := time.Duration((1 - b.tokens) / l.perSecond * float64(time.Second))
		return false, wait
	}
	if consume {
		b.tokens--
	}
	return true, 0
}

// sweep drops buckets that have refilled completely, so memory stays small.
// It runs at most once a minute, inside a normal call (no background goroutine).
func (l *Limiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < time.Minute {
		return
	}
	l.lastSweep = now
	for k, b := range l.buckets {
		if b.tokens+now.Sub(b.last).Seconds()*l.perSecond >= l.capacity {
			delete(l.buckets, k)
		}
	}
}
