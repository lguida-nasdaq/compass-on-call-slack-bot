package app

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/metriodev/pompiers/internal/adapters/api"
	"github.com/metriodev/pompiers/internal/config"
	"github.com/metriodev/pompiers/internal/domain"
	"golang.org/x/sync/errgroup"
)

type App struct {
	Config        config.Config
	CompassClient *api.CompassClient
	JiraClient    *api.JiraClient
}

func NewApp(cfg config.Config, cc api.CompassClient, jc api.JiraClient) *App {
	return &App{
		Config:        cfg,
		CompassClient: &cc,
		JiraClient:    &jc,
	}
}

type AppError struct {
	Err      error
	HttpCode int
}

func (a AppError) Error() string {
	return a.Err.Error()
}

func (a *App) GetCurrentOnCallSchedule() (domain.CurrentOnCallSchedule, error) {
	// Fetch all schedules
	schedules, err := a.CompassClient.GetSchedules()
	if err != nil {
		return domain.CurrentOnCallSchedule{}, err
	}

	var mu sync.Mutex
	g := new(errgroup.Group)

	currentSchedules := make([]domain.Schedule, 0)

	// currentSchedule := domain.CurrentOnCallSchedule{
	// 	Schedules: make([]domain.Schedule, len(schedules)),
	// }

	// Fetch OnCallParticipants for each schedule in parallel
	for _, schedule := range schedules {
		g.Go(func() error {
			onCallResponse, err := a.CompassClient.GetOnCallSchedules(schedule.ID)
			if err != nil {
				slog.Error(
					"Error fetching on-call schedule",
					"scheduleID", schedule.ID,
					"scheduleName", schedule.Name,
					"error", err,
				)
				return fmt.Errorf("error fetching on-call schedule for %s: %w", schedule.Name, err)
			}

			// Filter participants with type "user"
			var users []string
			for _, participant := range onCallResponse.OnCallParticipants {
				if participant.Type == "user" {
					userInfo, err := a.JiraClient.GetUserInfo(participant.ID)
					if err != nil {
						slog.Error("Error fetching user info", "error", err)
						return fmt.Errorf("error fetching user info for %s: %w", participant.ID, err)
					}
					users = append(users, userInfo.DisplayName)
				}
			}

			if len(users) == 0 {
				users = []string{"No one is on call"}
			}

			// Append to the response safely
			mu.Lock()
			currentSchedules = append(
				currentSchedules,
				domain.Schedule{
					Name:        schedule.Name,
					OnCallUsers: strings.Join(users, ", "),
				},
			)
			mu.Unlock()

			return nil
		})
	}

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return domain.CurrentOnCallSchedule{}, err
	}

	return domain.CurrentOnCallSchedule{Schedules: currentSchedules}, nil
}
