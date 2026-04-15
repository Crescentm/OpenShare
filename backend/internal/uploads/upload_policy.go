package uploads

import (
	"context"
	"errors"

	"openshare/backend/internal/receipts"
	"openshare/backend/internal/settings"
)

func (s *PublicUploadService) effectivePolicy(ctx context.Context) settings.SystemPolicy {
	if s.systemSetting == nil {
		return settings.SystemPolicy{
			Upload: settings.UploadPolicy{
				MaxUploadTotalBytes: s.config.MaxUploadTotalBytes,
			},
		}
	}

	policy, err := s.systemSetting.GetPolicy(ctx)
	if err != nil || policy == nil {
		return settings.SystemPolicy{
			Upload: settings.UploadPolicy{
				MaxUploadTotalBytes: s.config.MaxUploadTotalBytes,
			},
		}
	}

	return *policy
}

func (s *PublicUploadService) resolveReceiptCode(ctx context.Context, receiptCode string) (string, error) {
	code, err := s.receiptCodes.ResolveForSession(ctx, receiptCode)
	if errors.Is(err, receipts.ErrReceiptCodeGenerate) {
		return "", ErrReceiptCodeGenerate
	}
	return code, err
}
