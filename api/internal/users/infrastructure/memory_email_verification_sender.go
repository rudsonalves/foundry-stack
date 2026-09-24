package infrastructure

import (
	"context"
	"sync"

	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type MemoryEmailVerificationSender struct {
	mu       sync.RWMutex
	messages []domain.SendEmailVerificationCodeDTO
}

var _ domain.EmailVerificationSender = (*MemoryEmailVerificationSender)(nil)

func NewMemoryEmailVerificationSender() *MemoryEmailVerificationSender {
	return &MemoryEmailVerificationSender{
		messages: make([]domain.SendEmailVerificationCodeDTO, 0),
	}
}

func (s *MemoryEmailVerificationSender) SendVerificationCode(
	ctx context.Context,
	input domain.SendEmailVerificationCodeDTO,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, input)

	return nil
}

func (s *MemoryEmailVerificationSender) Messages() []domain.SendEmailVerificationCodeDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages := make(
		[]domain.SendEmailVerificationCodeDTO,
		len(s.messages),
	)
	copy(messages, s.messages)

	return messages
}
