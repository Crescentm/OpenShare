package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"

	"openshare/backend/internal/repository"
)

type ReceiptCodeService struct {
	repository *repository.ReceiptCodeRepository
	length     int
	codeGen    func(int) (string, error)
}

func NewReceiptCodeService(repository *repository.ReceiptCodeRepository, length int) *ReceiptCodeService {
	return &ReceiptCodeService{
		repository: repository,
		length:     length,
		codeGen:    generateReceiptCode,
	}
}

func (s *ReceiptCodeService) ResolveForSession(ctx context.Context, existing string) (string, error) {
	if normalized, err := normalizeReceiptCode(existing); err == nil && normalized != "" {
		return normalized, nil
	}

	for i := 0; i < maxGeneratedReceiptAttempts; i++ {
		candidate, err := s.codeGen(s.length)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrReceiptCodeGenerate, err)
		}

		exists, err := s.repository.Exists(ctx, candidate)
		if err != nil {
			return "", fmt.Errorf("check receipt code existence: %w", err)
		}
		if !exists {
			return candidate, nil
		}
	}

	return "", ErrReceiptCodeGenerate
}

func normalizeReceiptCode(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if len(raw) < 6 || len(raw) > 64 {
		return "", ErrInvalidUploadInput
	}
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return "", ErrInvalidUploadInput
		}
	}

	return raw, nil
}

func generateReceiptCode(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	builder := strings.Builder{}
	builder.Grow(length)
	for _, b := range raw {
		builder.WriteByte(alphabet[int(b)%len(alphabet)])
	}

	return builder.String(), nil
}
