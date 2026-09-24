package domain

import "errors"

var (
	ErrClientNotAllowed    = errors.New("client not allowed")
	ErrInvalidRefreshToken = errors.New("refresh token inválido")
)
