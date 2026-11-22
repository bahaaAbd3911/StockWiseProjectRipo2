package services

import (
	"testing"
	"time"
)

func TestDashboardWindowUsesLocalDays(t *testing.T) {
	amman, err := time.LoadLocation("Asia/Amman")
	if err != nil {
		t.Fatal(err)
	}
	// 22:30 UTC on 20 Nov = 01:30 on 21 Nov in Amman (UTC+3): "today" is the 21st.
	now := time.Date(2026, 11, 20, 22, 30, 0, 0, time.UTC)
	w := DashboardWindow(now, amman, 7, 5)

	want := func(name string, got time.Time, y int, m time.Month, d int) {
		t.Helper()
		if exp := time.Date(y, m, d, 0, 0, 0, 0, amman); !got.Equal(exp) {
			t.Errorf("%s: got %s, want %s", name, got.In(amman), exp)
		}
	}
	want("today from", w.TodayFrom, 2026, 11, 21)
	want("today to", w.TodayTo, 2026, 11, 22)
	want("period from (7 days incl. today)", w.From, 2026, 11, 15)
	want("period to", w.To, 2026, 11, 22)
	want("previous from", w.PrevFrom, 2026, 11, 8)
	want("previous to", w.PrevTo, 2026, 11, 15)
	if w.Timezone != "Asia/Amman" || w.LowStockThreshold != 5 {
		t.Fatalf("unexpected window: %+v", w)
	}
}
