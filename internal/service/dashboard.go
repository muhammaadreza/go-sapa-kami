package service

import (
	"context"
	"fmt"
	"time"

	"go-sapa-kami/internal/apperror"
	"go-sapa-kami/internal/dto"
	"go-sapa-kami/internal/model"

	"golang.org/x/sync/errgroup"
)

type DashboardService interface {
	GetDashboard(ctx context.Context) (*dto.Dashboard, error)
}

type dashboardService struct {
	users    UserService
	messages AspirationService
}

func NewDashboardService(users UserService, messages AspirationService) DashboardService {
	return &dashboardService{users, messages}
}

func (s *dashboardService) GetDashboard(ctx context.Context) (*dto.Dashboard, error) {
	now := time.Now()
	currStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	currEnd := currStart.AddDate(0, 1, -1)

	prevStart := currStart.AddDate(0, -1, 0)
	prevEnd := currStart.AddDate(0, 0, -1)

	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	yearEnd := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, now.Location())

	days := 7
	sevenDaysAgo := now.AddDate(0, 0, -(days - 1))

	currStartStr, currEndStr := currStart.Format("2006-01-02"), currEnd.Format("2006-01-02")
	prevStartStr, prevEndStr := prevStart.Format("2006-01-02"), prevEnd.Format("2006-01-02")
	yearStartStr, yearEndStr := yearStart.Format("2006-01-02"), yearEnd.Format("2006-01-02")

	typeInnovation := model.TypeInnovation
	typeCriticism := model.TypeCriticism

	var (
		currTotalMsg, currTotalInn, currTotalCrit, currTotalUsers int64
		prevTotalMsg, prevTotalInn, prevTotalCrit, prevTotalUsers int64

		allDailyMap        map[string]int64
		innovationDailyMap map[string]int64
		criticismDailyMap  map[string]int64
		userDailyMap       map[string]int64
		recentAspirations  []dto.AspirationResponse
	)

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		filter := dto.AspirationFilterRequest{StartDate: &currStartStr, EndDate: &currEndStr}
		count, err := s.messages.Count(groupCtx, &filter)
		currTotalMsg = count
		return err
	})

	group.Go(func() error {
		filter := dto.AspirationFilterRequest{Type: &typeInnovation, StartDate: &currStartStr, EndDate: &currEndStr}
		count, err := s.messages.Count(groupCtx, &filter)
		currTotalInn = count
		return err
	})

	group.Go(func() error {
		filter := dto.AspirationFilterRequest{Type: &typeCriticism, StartDate: &currStartStr, EndDate: &currEndStr}
		count, err := s.messages.Count(groupCtx, &filter)
		currTotalCrit = count
		return err
	})

	group.Go(func() error {
		filter := dto.UserFilterRequest{StartDate: &currStartStr, EndDate: &currEndStr}
		count, err := s.users.CountActive(groupCtx, &filter)
		currTotalUsers = count
		return err
	})

	group.Go(func() error {
		filter := dto.AspirationFilterRequest{StartDate: &prevStartStr, EndDate: &prevEndStr}
		count, err := s.messages.Count(groupCtx, &filter)
		prevTotalMsg = count
		return err
	})

	group.Go(func() error {
		filter := dto.AspirationFilterRequest{Type: &typeInnovation, StartDate: &prevStartStr, EndDate: &prevEndStr}
		count, err := s.messages.Count(groupCtx, &filter)
		prevTotalInn = count
		return err
	})

	group.Go(func() error {
		filter := dto.AspirationFilterRequest{Type: &typeCriticism, StartDate: &prevStartStr, EndDate: &prevEndStr}
		count, err := s.messages.Count(groupCtx, &filter)
		prevTotalCrit = count
		return err
	})

	group.Go(func() error {
		filter := dto.UserFilterRequest{StartDate: &prevStartStr, EndDate: &prevEndStr}
		count, err := s.users.CountActive(groupCtx, &filter)
		prevTotalUsers = count
		return err
	})

	group.Go(func() error {
		filter := dto.AspirationFilterRequest{StartDate: &yearStartStr, EndDate: &yearEndStr}

		dataMap, err := s.messages.CountDaily(groupCtx, &filter)
		allDailyMap = dataMap
		return err
	})

	group.Go(func() error {
		filter := dto.AspirationFilterRequest{Type: &typeInnovation, StartDate: &yearStartStr, EndDate: &yearEndStr}
		dataMap, err := s.messages.CountDaily(groupCtx, &filter)
		innovationDailyMap = dataMap
		return err
	})

	group.Go(func() error {
		filter := dto.AspirationFilterRequest{Type: &typeCriticism, StartDate: &yearStartStr, EndDate: &yearEndStr}
		dataMap, err := s.messages.CountDaily(groupCtx, &filter)
		criticismDailyMap = dataMap
		return err
	})

	group.Go(func() error {
		filter := dto.UserFilterRequest{StartDate: &yearStartStr, EndDate: &yearEndStr}
		dataMap, err := s.users.CountDaily(groupCtx, &filter)
		userDailyMap = dataMap
		return err
	})

	group.Go(func() error {
		filter := dto.AspirationFilterRequest{Limit: 5, Page: 1}
		messages, _, err := s.messages.GetAll(groupCtx, 0, model.RoleAdmin, &filter)
		recentAspirations = messages
		return err
	})

	err := group.Wait()
	if err != nil {
		return nil, apperror.ErrInternal("Failed to retrieve dashboard data")
	}

	allSparkline := generateSparkline(allDailyMap, sevenDaysAgo, days)
	innovationSparkline := generateSparkline(innovationDailyMap, sevenDaysAgo, days)
	criticismSparkline := generateSparkline(criticismDailyMap, sevenDaysAgo, days)
	userSparkline := generateSparkline(userDailyMap, sevenDaysAgo, days)

	var weeklyCharts []dto.Chart
	for i := 0; i < days; i++ {
		currentDate := sevenDaysAgo.AddDate(0, 0, i).Format("2006-01-02")
		weeklyCharts = append(weeklyCharts, dto.Chart{
			Date:        currentDate,
			Innovations: innovationSparkline[i],
			Criticisms:  criticismSparkline[i],
		})
	}

	monthlyCharts := generateMonthlyChart(innovationDailyMap, criticismDailyMap, now.Year())
	response := &dto.Dashboard{
		KPIs: dto.KPIs{
			Aspirations: dto.KPI{
				Total:      currTotalMsg,
				Percentage: calculatePercentage(currTotalMsg, prevTotalMsg),
				Sparkline:  allSparkline,
			},
			Innovations: dto.KPI{
				Total:      currTotalInn,
				Percentage: calculatePercentage(currTotalInn, prevTotalInn),
				Sparkline:  innovationSparkline,
			},
			Critisms: dto.KPI{
				Total:      currTotalCrit,
				Percentage: calculatePercentage(currTotalCrit, prevTotalCrit),
				Sparkline:  criticismSparkline,
			},
			ActiveUsers: dto.KPI{
				Total:      currTotalUsers,
				Percentage: calculatePercentage(currTotalUsers, prevTotalUsers),
				Sparkline:  userSparkline,
			},
		},
		Charts: dto.Charts{
			Weekly:  weeklyCharts,
			Monthly: monthlyCharts,
		},
		Aspirations: recentAspirations,
	}

	return response, nil
}

func calculatePercentage(current, previous int64) int64 {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return int64((float64(current-previous) / float64(previous)) * 100)
}

// Helper function to generate sparkline data for the dashboard
func generateSparkline(count map[string]int64, startDate time.Time, day int) []int64 {
	sparkline := make([]int64, day)

	for i := 0; i < day; i++ {
		current := startDate.AddDate(0, 0, i).Format("2006-01-02")
		if val, ok := count[current]; ok {
			sparkline[i] = val
		} else {
			sparkline[i] = 0
		}
	}

	return sparkline
}

func generateMonthlyChart(innMap, critMap map[string]int64, year int) []dto.Chart {
	charts := make([]dto.Chart, 12)
	for i := 0; i < 12; i++ {
		monthPrefix := fmt.Sprintf("%04d-%02d", year, i+1)
		var innTotal, critTotal int64

		for date, count := range innMap {
			if len(date) >= 7 && date[:7] == monthPrefix {
				innTotal += count
			}
		}
		for date, count := range critMap {
			if len(date) >= 7 && date[:7] == monthPrefix {
				critTotal += count
			}
		}

		charts[i] = dto.Chart{
			Date:        monthPrefix,
			Innovations: innTotal,
			Criticisms:  critTotal,
		}
	}
	return charts
}
