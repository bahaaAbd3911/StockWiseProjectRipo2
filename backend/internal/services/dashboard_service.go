package services

import (
	"context"
	"time"

	"ecommerce-backend/internal/models"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/validators"
)

// DashboardService turns "the last N days" into exact time windows in the
// store's time zone (a day starts at local midnight, not UTC midnight) and
// asks the repository for the numbers.
type DashboardService struct {
	repo     *repository.DashboardRepository
	loc      *time.Location
	lowStock int
	now      func() time.Time // replaced in tests
}

func NewDashboardService(repo *repository.DashboardRepository, loc *time.Location, lowStockThreshold int) *DashboardService {
	if loc == nil {
		loc = time.UTC
	}
	return &DashboardService{repo: repo, loc: loc, lowStock: lowStockThreshold, now: time.Now}
}

const (
	DefaultDashboardDays = 30
	MaxDashboardDays     = 366
)

func (s *DashboardService) Get(ctx context.Context, days int) (models.Dashboard, error) {
	if days < 1 || days > MaxDashboardDays {
		return models.Dashboard{}, validators.FieldError("days", "must be between 1 and 366")
	}
	return s.repo.Build(ctx, DashboardWindow(s.now(), s.loc, days, s.lowStock))
}

// DashboardWindow computes the windows: today, the last `days` local days
// (today included), and the same length just before. AddDate keeps local
// midnight correct across daylight-saving changes.
func DashboardWindow(now time.Time, loc *time.Location, days, lowStock int) repository.DashboardRange {
	local := now.In(loc)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	tomorrow := today.AddDate(0, 0, 1)
	from := today.AddDate(0, 0, -(days - 1))
	return repository.DashboardRange{
		TodayFrom: today, TodayTo: tomorrow,
		From: from, To: tomorrow,
		PrevFrom: from.AddDate(0, 0, -days), PrevTo: from,
		Timezone:          loc.String(),
		LowStockThreshold: lowStock,
	}
}
