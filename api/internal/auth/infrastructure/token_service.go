package infrastructure

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	authdomain "github.com/rudsonalves/foundry-stack/api/internal/auth/domain"
)

type TokenServiceConfig struct {
	Secret    []byte
	Issuer    string
	Audience  string
	ClientIDs []string
	Lifetime  time.Duration
}

type TokenService struct {
	secret    []byte
	issuer    string
	audience  string
	clientIDs []string
	lifetime  time.Duration
	now       func() time.Time
}

func NewTokenService(
	config TokenServiceConfig,
) (*TokenService, error) {
	if len(config.Secret) < 32 {
		return nil, errors.New("JWT secret must be at least 32 bytes long")
	}
	if config.Issuer == "" || config.Audience == "" {
		return nil, errors.New("JWT issuer and audience are required")
	}
	if config.Lifetime <= 0 {
		return nil, errors.New("JWT lifetime must be greater than zero")
	}

	return &TokenService{
		secret:    append([]byte(nil), config.Secret...),
		issuer:    config.Issuer,
		audience:  config.Audience,
		clientIDs: append([]string(nil), config.ClientIDs...),
		lifetime:  config.Lifetime,
		now:       time.Now,
	}, nil
}

func (s *TokenService) acceptsClient(clientID string) bool {
	return slices.Contains(s.clientIDs, clientID)
}

func (s *TokenService) Issue(
	userID uuid.UUID,
	clientID string,
) (string, time.Time, error) {
	if !s.acceptsClient(clientID) {
		return "", time.Time{}, authdomain.ErrClientNotAllowed
	}

	now := s.now().UTC()
	expiresAt := now.Add(s.lifetime)

	claims := newAccessClaims(
		userID,
		clientID,
		s.issuer,
		s.audience,
		now,
		expiresAt,
	)

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf(
			"sign token: %w",
			err,
		)
	}

	return signed, expiresAt, nil
}

var ErrInvalidToken = errors.New("invalid token")

func (s *TokenService) Parse(
	tokenString string,
) (uuid.UUID, error) {
	claims := new(AccessClaims)

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return s.secret, nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(s.now),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}
	if !s.acceptsClient(claims.ClientID) {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}
