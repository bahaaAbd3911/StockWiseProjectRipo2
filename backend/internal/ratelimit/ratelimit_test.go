package ratelimit

import (
	"testing"
	"time"
)

func fakeClock(l *Limiter) *time.Time {
	t := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	l.now = func() time.Time { return t }
	return &t
}

func TestBurstThenRefill(t *testing.T) {
	l := New(Rule{Requests: 3, Per: time.Minute}) // 1 token every 20 s
	now := fakeClock(l)

	for i := 0; i < 3; i++ {
		if ok, _ := l.Allow("ip"); !ok {
			t.Fatalf("request %d of the burst was refused", i+1)
		}
	}
	ok, wait := l.Allow("ip")
	if ok || wait <= 0 || wait > 20*time.Second {
		t.Fatalf("4th request: ok=%v wait=%v; want refused with wait <= 20s", ok, wait)
	}
	if ok, _ := l.Allow("other-ip"); !ok {
		t.Fatal("keys must be independent")
	}

	*now = now.Add(20 * time.Second)
	if ok, _ := l.Allow("ip"); !ok {
		t.Fatal("one token should have come back after 20s")
	}
	if ok, _ := l.Allow("ip"); ok {
		t.Fatal("only one token should have come back")
	}
}

func TestCheckFailReset(t *testing.T) {
	l := New(Rule{Requests: 2, Per: time.Hour})
	fakeClock(l)
	l.Fail("a@x")
	if ok, _ := l.Check("a@x"); !ok {
		t.Fatal("one failure out of two is still allowed")
	}
	l.Fail("a@x")
	if ok, _ := l.Check("a@x"); ok {
		t.Fatal("two failures should block")
	}
	l.Reset("a@x")
	if ok, _ := l.Check("a@x"); !ok {
		t.Fatal("reset should unblock")
	}
}

func TestNilLimiterAllowsEverything(t *testing.T) {
	var l *Limiter = New(Rule{}) // disabled
	for i := 0; i < 100; i++ {
		if ok, _ := l.Allow("x"); !ok {
			t.Fatal("a disabled limiter must allow everything")
		}
	}
	l.Fail("x")
	l.Reset("x")
}

func TestSweepRemovesIdleBuckets(t *testing.T) {
	l := New(Rule{Requests: 1, Per: time.Second})
	now := fakeClock(l)
	l.Allow("a")
	*now = now.Add(2 * time.Minute)
	l.Allow("b") // triggers the sweep
	if _, ok := l.buckets["a"]; ok {
		t.Fatal("idle, refilled bucket should have been swept")
	}
}
