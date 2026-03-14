package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/masterfabric/masterfabric_go_basic/internal/application/settings/dto"
	settingsDomainEvent "github.com/masterfabric/masterfabric_go_basic/internal/domain/settings/event"
	settingsModel "github.com/masterfabric/masterfabric_go_basic/internal/domain/settings/model"
	settingsRepo "github.com/masterfabric/masterfabric_go_basic/internal/domain/settings/repository"
	domainErr "github.com/masterfabric/masterfabric_go_basic/internal/shared/errors"
	"github.com/masterfabric/masterfabric_go_basic/internal/shared/events"
)

// GetUserSettingsUseCase retrieves settings for a user, creating defaults if none exist.
type GetUserSettingsUseCase struct {
	repo settingsRepo.UserSettingsRepository
}

// NewGetUserSettingsUseCase constructs a GetUserSettingsUseCase.
func NewGetUserSettingsUseCase(repo settingsRepo.UserSettingsRepository) *GetUserSettingsUseCase {
	return &GetUserSettingsUseCase{repo: repo}
}

// Execute returns user settings or default values.
func (uc *GetUserSettingsUseCase) Execute(ctx context.Context, userID string) (*dto.UserSettingsResponse, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, domainErr.ErrUnauthorized
	}

	s, err := uc.repo.FindByUserID(ctx, id)
	if err != nil {
		// Return defaults if not found
		return &dto.UserSettingsResponse{
			UserID:          userID,
			NotificationsOn: true,
			Theme:           "system",
			Language:        "en",
			Timezone:        "UTC",
		}, nil
	}

	return &dto.UserSettingsResponse{
		UserID:          userID,
		NotificationsOn: s.NotificationsOn,
		Theme:           s.Theme,
		Language:        s.Language,
		Timezone:        s.Timezone,
	}, nil
}

// UpdateUserSettingsUseCase persists user preference changes.
type UpdateUserSettingsUseCase struct {
	repo     settingsRepo.UserSettingsRepository
	eventBus events.EventBus
}

// NewUpdateUserSettingsUseCase constructs an UpdateUserSettingsUseCase.
func NewUpdateUserSettingsUseCase(repo settingsRepo.UserSettingsRepository, eventBus events.EventBus) *UpdateUserSettingsUseCase {
	return &UpdateUserSettingsUseCase{repo: repo, eventBus: eventBus}
}

// Execute upserts the user settings.
func (uc *UpdateUserSettingsUseCase) Execute(ctx context.Context, req *dto.UserSettingsRequest) (*dto.UserSettingsResponse, error) {
	id, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, domainErr.ErrUnauthorized
	}

	// Load or create
	s, err := uc.repo.FindByUserID(ctx, id)
	if err != nil {
		s = &settingsModel.UserSettings{
			ID:              uuid.New(),
			UserID:          id,
			NotificationsOn: true,
			Theme:           "system",
			Language:        "en",
			Timezone:        "UTC",
		}
	}

	// Apply partial updates
	if req.NotificationsOn != nil {
		s.NotificationsOn = *req.NotificationsOn
	}
	if req.Theme != nil {
		s.Theme = *req.Theme
	}
	if req.Language != nil {
		s.Language = *req.Language
	}
	if req.Timezone != nil {
		s.Timezone = *req.Timezone
	}
	s.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Upsert(ctx, s); err != nil {
		return nil, fmt.Errorf("update settings: %w", err)
	}

	_ = uc.eventBus.Publish(ctx, events.TopicUserSettingsUpdated, events.Event{
		Type: settingsDomainEvent.EventUserSettingsUpdated,
		Payload: settingsDomainEvent.UserSettingsUpdatedPayload{
			UserID:    req.UserID,
			UpdatedAt: s.UpdatedAt,
		},
	})

	return &dto.UserSettingsResponse{
		UserID:          req.UserID,
		NotificationsOn: s.NotificationsOn,
		Theme:           s.Theme,
		Language:        s.Language,
		Timezone:        s.Timezone,
	}, nil
}

// GetAppSettingsUseCase returns public application configuration.
type GetAppSettingsUseCase struct {
	repo settingsRepo.AppSettingsRepository
}

// NewGetAppSettingsUseCase constructs a GetAppSettingsUseCase.
func NewGetAppSettingsUseCase(repo settingsRepo.AppSettingsRepository) *GetAppSettingsUseCase {
	return &GetAppSettingsUseCase{repo: repo}
}

// Execute returns all public app settings.
func (uc *GetAppSettingsUseCase) Execute(ctx context.Context) ([]*dto.AppSettingResponse, error) {
	settings, err := uc.repo.ListPublic(ctx)
	if err != nil {
		return nil, fmt.Errorf("list app settings: %w", err)
	}

	resp := make([]*dto.AppSettingResponse, 0, len(settings))
	for _, s := range settings {
		resp = append(resp, &dto.AppSettingResponse{
			Key:         s.Key,
			Value:       s.Value,
			Description: s.Description,
		})
	}
	return resp, nil
}
