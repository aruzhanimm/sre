package service

import (
	"context"
	"errors"
	"time"

	"BetKZ/internal/models"
	"BetKZ/internal/repository"

	"github.com/google/uuid"
)

type EventService struct {
	eventRepo *repository.EventRepository
	sportRepo *repository.SportRepository
}

func NewEventService(eventRepo *repository.EventRepository, sportRepo *repository.SportRepository) *EventService {
	return &EventService{
		eventRepo: eventRepo,
		sportRepo: sportRepo,
	}
}

type CreateEventRequest struct {
	SportID   int       `json:"sport_id" binding:"required"`
	HomeTeam  string    `json:"home_team" binding:"required"`
	AwayTeam  string    `json:"away_team" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
}

type UpdateEventRequest struct {
	HomeTeam  *string    `json:"home_team"`
	AwayTeam  *string    `json:"away_team"`
	StartTime *time.Time `json:"start_time"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (s *EventService) ListSports(ctx context.Context) ([]models.Sport, error) {
	return s.sportRepo.List(ctx)
}

func (s *EventService) ListEvents(ctx context.Context, filters repository.EventFilters) (*repository.EventListResult, error) {
	return s.eventRepo.List(ctx, filters)
}

func (s *EventService) GetEvent(ctx context.Context, idStr string) (*models.Event, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, errors.New("invalid event ID")
	}

	event, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, errors.New("event not found")
	}
	return event, nil
}

func (s *EventService) CreateEvent(ctx context.Context, req *CreateEventRequest) (*models.Event, error) {
	if req.HomeTeam == "" || req.AwayTeam == "" {
		return nil, errors.New("team names are required")
	}
	if req.StartTime.Before(time.Now()) {
		return nil, errors.New("start time must be in the future")
	}

	// Verify sport exists
	_, err := s.sportRepo.GetByID(ctx, req.SportID)
	if err != nil {
		return nil, errors.New("invalid sport ID")
	}

	event := &models.Event{
		SportID:   req.SportID,
		HomeTeam:  req.HomeTeam,
		AwayTeam:  req.AwayTeam,
		StartTime: req.StartTime,
		Status:    "upcoming",
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *EventService) UpdateEvent(ctx context.Context, idStr string, req *UpdateEventRequest) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return errors.New("invalid event ID")
	}

	event, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return errors.New("event not found")
	}
	if event.Status != "upcoming" {
		return errors.New("can only edit upcoming events")
	}

	if req.HomeTeam != nil {
		event.HomeTeam = *req.HomeTeam
	}
	if req.AwayTeam != nil {
		event.AwayTeam = *req.AwayTeam
	}
	if req.StartTime != nil {
		event.StartTime = *req.StartTime
	}

	return s.eventRepo.Update(ctx, event)
}

func (s *EventService) UpdateEventStatus(ctx context.Context, idStr string, req *UpdateStatusRequest) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return errors.New("invalid event ID")
	}

	validStatuses := map[string]bool{
		"upcoming": true, "live": true, "finished": true, "cancelled": true,
	}
	if !validStatuses[req.Status] {
		return errors.New("invalid status")
	}

	event, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if event == nil {
		return errors.New("event not found")
	}

	// Validate status transitions
	validTransitions := map[string][]string{
		"upcoming":  {"live", "cancelled"},
		"live":      {"finished", "cancelled"},
		"finished":  {},
		"cancelled": {},
	}

	allowed := false
	for _, s := range validTransitions[event.Status] {
		if s == req.Status {
			allowed = true
			break
		}
	}
	if !allowed {
		return errors.New("invalid status transition from " + event.Status + " to " + req.Status)
	}

	return s.eventRepo.UpdateStatus(ctx, id, req.Status)
}

func (s *EventService) DeleteEvent(ctx context.Context, idStr string) error {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return errors.New("invalid event ID")
	}
	return s.eventRepo.Delete(ctx, id)
}
