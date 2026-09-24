package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
	authdomain "github.com/rudsonalves/foundry-stack/api/internal/auth/domain"
)

type RefreshTokenResult struct {
	Token     string
	ExpiresAt time.Time
}

type RefreshIdentity struct {
	UserID   uuid.UUID
	ClientID string
}

type RefreshTokenService struct {
	repository authdomain.RefreshTokenRepository
	lifetime   time.Duration
	now        func() time.Time
}

func NewRefreshTokenService(
	repository authdomain.RefreshTokenRepository,
	lifetime time.Duration,
) *RefreshTokenService {
	return &RefreshTokenService{
		repository: repository,
		lifetime:   lifetime,
		now:        time.Now,
	}
}

func (s *RefreshTokenService) Issue(
	ctx context.Context,
	userID uuid.UUID,
	clientID string,
) (RefreshTokenResult, error) {
	token, tokenHash, err := generateRefreshToken()
	if err != nil {
		return RefreshTokenResult{}, err
	}

	expiresAt := s.now().UTC().Add(s.lifetime)
	err = s.repository.Create(
		ctx,
		authdomain.CreateRefreshTokenDTO{
			TokenHash: tokenHash,
			UserID:    userID,
			ClientID:  clientID,
			ExpiresAt: expiresAt,
		},
	)
	if err != nil {
		return RefreshTokenResult{}, fmt.Errorf(
			"persist refresh token: %w",
			err,
		)
	}

	return RefreshTokenResult{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *RefreshTokenService) Authenticate(
	ctx context.Context,
	token string,
) (RefreshIdentity, error) {
	if token == "" {
		return RefreshIdentity{},
			authdomain.ErrInvalidRefreshToken
	}

	stored, err := s.repository.FindActive(
		ctx,
		refreshTokenHash(token),
		s.now().UTC(),
	)
	if err != nil {
		return RefreshIdentity{}, err
	}

	return RefreshIdentity{
		UserID:   stored.UserID,
		ClientID: stored.ClientID,
	}, nil
}

func (s *RefreshTokenService) Revoke(
	ctx context.Context,
	token string,
) error {
	if token == "" {
		return nil
	}

	return s.repository.Revoke(
		ctx,
		refreshTokenHash(token),
		s.now().UTC(),
	)
}

func generateRefreshToken() (string, []byte, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", nil, fmt.Errorf(
			"generate refresh token: %w",
			err,
		)
	}

	token := base64.RawURLEncoding.EncodeToString(value)
	return token, refreshTokenHash(token), nil
}

func refreshTokenHash(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}
