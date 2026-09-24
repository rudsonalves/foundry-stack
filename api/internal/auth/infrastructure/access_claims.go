package infrastructure

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessClaims struct {
	ClientID string `json:"client_id"`
	jwt.RegisteredClaims
}

func newAccessClaims(
	userID uuid.UUID,
	clientID string,
	issuer string,
	audience string,
	issuedAt time.Time,
	expiresAt time.Time,
) AccessClaims {
	return AccessClaims{
		ClientID: clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
}
