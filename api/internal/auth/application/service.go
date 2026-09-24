package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	authdomain "github.com/rudsonalves/foundry-stack/api/internal/auth/domain"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	usersApp "github.com/rudsonalves/foundry-stack/api/internal/users/application"
	userDomain "github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type RefreshTokens interface {
	Issue(
		ctx context.Context,
		userID uuid.UUID,
		clientID string,
	) (RefreshTokenResult, error)

	Authenticate(
		ctx context.Context,
		token string,
	) (RefreshIdentity, error)

	Revoke(
		ctx context.Context,
		token string,
	) error
}

type RefreshResult struct {
	AccessToken string
	ExpiresAt   time.Time
}

type CredentialAuthenticator interface {
	Authenticate(
		ctx context.Context,
		email string,
		password string,
	) (userDomain.User, error)
}

type TokenIssuer interface {
	Issue(
		userID uuid.UUID,
		clientID string,
	) (string, time.Time, error)
}

type AuthService struct {
	users         CredentialAuthenticator
	tokens        TokenIssuer
	refreshTokens RefreshTokens
}

func NewAuthService(
	users CredentialAuthenticator,
	tokens TokenIssuer,
	refreshTokens RefreshTokens,
) *AuthService {
	return &AuthService{
		users:         users,
		tokens:        tokens,
		refreshTokens: refreshTokens,
	}
}

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
	clientID string,
) (LoginResult, error) {
	user, err := s.users.Authenticate(
		ctx,
		email,
		password,
	)
	if errors.Is(err, usersApp.ErrInvalidCredentials) {
		return LoginResult{}, sharederrors.NewUnauthorized(
			"invalid email or password",
			err,
		)
	}
	if err != nil {
		return LoginResult{}, err
	}

	accessToken, expiresAt, err := s.tokens.Issue(
		user.ID,
		clientID,
	)
	if errors.Is(err, authdomain.ErrClientNotAllowed) {
		return LoginResult{}, sharederrors.NewUnauthorized(
			"client not allowed",
			err,
		)
	}
	if err != nil {
		return LoginResult{}, fmt.Errorf(
			"issue access token: %w",
			err,
		)
	}

	refreshToken, err := s.refreshTokens.Issue(
		ctx,
		user.ID,
		clientID,
	)
	if err != nil {
		return LoginResult{}, fmt.Errorf(
			"issue refresh token: %w",
			err,
		)
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *AuthService) Refresh(
	ctx context.Context,
	refreshToken string,
) (RefreshResult, error) {
	identity, err := s.refreshTokens.Authenticate(
		ctx,
		refreshToken,
	)
	if errors.Is(err, authdomain.ErrInvalidRefreshToken) {
		return RefreshResult{}, sharederrors.NewUnauthorized(
			"invalid refresh token",
			err,
		)
	}
	if err != nil {
		return RefreshResult{}, fmt.Errorf(
			"authenticate refresh token: %w",
			err,
		)
	}

	accessToken, expiresAt, err := s.tokens.Issue(
		identity.UserID,
		identity.ClientID,
	)
	if err != nil {
		return RefreshResult{}, fmt.Errorf(
			"issue access token: %w",
			err,
		)
	}

	return RefreshResult{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *AuthService) Logout(
	ctx context.Context,
	refreshToken string,
) error {
	if err := s.refreshTokens.Revoke(
		ctx,
		refreshToken,
	); err != nil {
		return fmt.Errorf(
			"revoke refresh token: %w",
			err,
		)
	}

	return nil
}
