package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/masterfabric/masterfabric_go_basic/internal/application/auth/dto"
	"github.com/masterfabric/masterfabric_go_basic/internal/domain/iam/repository"
	infraAuth "github.com/masterfabric/masterfabric_go_basic/internal/infrastructure/auth"
	domainErr "github.com/masterfabric/masterfabric_go_basic/internal/shared/errors"
)

// RefreshUseCase rotates a refresh token and issues a new pair.
type RefreshUseCase struct {
	userRepo repository.UserRepository
	jwtSvc   *infraAuth.JWTService
}

// NewRefreshUseCase constructs a RefreshUseCase.
func NewRefreshUseCase(userRepo repository.UserRepository, jwtSvc *infraAuth.JWTService) *RefreshUseCase {
	return &RefreshUseCase{userRepo: userRepo, jwtSvc: jwtSvc}
}

// Execute validates the refresh token and returns a new token pair.
func (uc *RefreshUseCase) Execute(ctx context.Context, req *dto.RefreshRequest) (*dto.AuthResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, domainErr.ErrTokenInvalid
	}

	pair, err := uc.jwtSvc.RefreshTokens(ctx, userID, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, domainErr.ErrUserNotFound
	}

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

// LogoutUseCase revokes tokens.
type LogoutUseCase struct {
	jwtSvc *infraAuth.JWTService
}

// NewLogoutUseCase constructs a LogoutUseCase.
func NewLogoutUseCase(jwtSvc *infraAuth.JWTService) *LogoutUseCase {
	return &LogoutUseCase{jwtSvc: jwtSvc}
}

// Execute revokes the user's current tokens.
func (uc *LogoutUseCase) Execute(ctx context.Context, req *dto.LogoutRequest) error {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return domainErr.ErrTokenInvalid
	}

	if err := uc.jwtSvc.RevokeTokens(ctx, userID, req.AccessToken, req.RefreshToken); err != nil {
		return fmt.Errorf("logout: revoke tokens: %w", err)
	}
	return nil
}
