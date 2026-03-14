package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/masterfabric/masterfabric_go_basic/internal/application/auth/dto"
	"github.com/masterfabric/masterfabric_go_basic/internal/domain/iam/event"
	"github.com/masterfabric/masterfabric_go_basic/internal/domain/iam/repository"
	infraAuth "github.com/masterfabric/masterfabric_go_basic/internal/infrastructure/auth"
	domainErr "github.com/masterfabric/masterfabric_go_basic/internal/shared/errors"
	"github.com/masterfabric/masterfabric_go_basic/internal/shared/events"
)

// LoginUseCase authenticates a user by email/password.
type LoginUseCase struct {
	userRepo repository.UserRepository
	jwtSvc   *infraAuth.JWTService
	eventBus events.EventBus
}

// NewLoginUseCase constructs a LoginUseCase.
func NewLoginUseCase(
	userRepo repository.UserRepository,
	jwtSvc *infraAuth.JWTService,
	eventBus events.EventBus,
) *LoginUseCase {
	return &LoginUseCase{userRepo: userRepo, jwtSvc: jwtSvc, eventBus: eventBus}
}

// Execute validates credentials and issues tokens.
func (uc *LoginUseCase) Execute(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, domainErr.ErrInvalidCredentials
	}

	if err := infraAuth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, domainErr.ErrInvalidCredentials
	}

	pair, err := uc.jwtSvc.GenerateTokenPair(ctx, user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("login: generate tokens: %w", err)
	}

	_ = uc.eventBus.Publish(ctx, events.TopicUserLoggedIn, events.Event{
		Type: event.EventUserLoggedIn,
		Payload: event.UserLoggedInPayload{
			UserID:   user.ID.String(),
			Email:    user.Email,
			LoggedAt: time.Now().UTC(),
		},
	})

	return &dto.AuthResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		User: dto.UserData{
			ID:          user.ID.String(),
			Email:       user.Email,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
			Role:        string(user.Role),
		},
	}, nil
}
