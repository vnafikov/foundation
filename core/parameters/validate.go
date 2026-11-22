package parameters

import (
	"context"
	"log/slog"

	"github.com/vnafikov/foundation/core/contracterrors"
	"github.com/vnafikov/foundation/logger"
)

func NormalizeAndValidate(ctx context.Context, params NormalizerValidator) error {
	params.Normalize()
	return Validate(ctx, params)
}

func Validate(ctx context.Context, params Validator) error {
	if params.IsValid() {
		return nil
	}

	if logger.ShouldLogDebug() {
		logger.DebugError(ctx, contracterrors.ErrInvalid, "parameters.Validate")
	}
	return contracterrors.ErrInvalid
}

func NormalizeAndValidateWithErrors(ctx context.Context, params NormalizerValidatorWithErrors) error {
	params.Normalize()
	return ValidateWithErrors(ctx, params)
}

func ValidateWithErrors(ctx context.Context, params ValidatorWithErrors) error {
	v := params.Validate()
	if !v.HasErrors() {
		return nil
	}

	err := &contracterrors.InvalidError{Validator: v}
	if logger.ShouldLogDebug() {
		logger.DebugError(ctx, err, "parameters.ValidateWithErrors", slog.Any("errors", v.Errors))
	}
	return err
}
