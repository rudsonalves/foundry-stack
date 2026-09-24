package sharederrors

const (
	ErrCodeNotFound                 = "NOT_FOUND"
	ErrCodeBadRequest               = "BAD_REQUEST"
	ErrCodeUnauthorized             = "UNAUTHORIZED"
	ErrCodeForbidden                = "FORBIDDEN"
	ErrCodeConflict                 = "CONFLICT"
	ErrCodeEmailAlreadyRegistered   = "EMAIL_ALREADY_REGISTERED"
	ErrCodeEmailDeliveryUnavailable = "EMAIL_DELIVERY_UNAVAILABLE"
	ErrCodeRateLimitExceeded        = "RATE_LIMIT_EXCEEDED"
	ErrCodeInternalError            = "INTERNAL_ERROR"
	ErrCodeInvalidEmailVerification = "INVALID_EMAIL_VERIFICATION"
	ErrCodeInvalidPasswordReset     = "INVALID_PASSWORD_RESET"
)
