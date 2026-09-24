package infrastructure

import (
	"time"

	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type SystemClock struct{}

var _ domain.Clock = SystemClock{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}
