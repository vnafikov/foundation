package queries

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vnafikov/common/text"
	"github.com/vnafikov/common/validation"

	"github.com/vnafikov/foundation/core/contracterrors"
	"github.com/vnafikov/foundation/infrastructure/databases/pg/queries"
)

func contractError[T any](err error) error {
	if err == nil {
		return nil
	}

	if pgError, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgError.Code {
		case queries.ErrorCodeUniqueViolation:
			return &contracterrors.UniqueViolationError{
				Validator: validatorWithUniqueViolationError[T](pgError),
				Err:       pgError,
			}
		case queries.ErrorCodeCheckViolation:
			return &contracterrors.CheckViolationError{Err: pgError}
		case queries.ErrorCodeExclusionViolation:
			return &contracterrors.ExclusionViolationError{
				Validator: validatorWithExclusionViolationError[T](pgError),
				Err:       pgError,
			}
		}
	}
	return &contracterrors.InternalError{Err: err}
}

func validatorWithUniqueViolationError[T any](pgError *pgconn.PgError) validation.Validator {
	columnList := columns(pgError)
	mapper := columnMapper[T]()
	var v validation.Validator
	for _, column := range columnList {
		v.AddError(mapper(column), validation.CodeUnique, nil)
	}
	return v
}

func columns(pgError *pgconn.PgError) []string {
	cols, _ := text.ExtractBetween(pgError.Detail, '(', ')')
	columnList := strings.Split(cols, ", ")
	if columnList[0] == "" {
		return nil
	}
	return columnList
}

func columnMapper[T any]() func(column string) string {
	var entity *T
	if entity, ok := any(entity).(metaProvider); ok {
		lookupMeta := entity.LookupMeta()
		return func(column string) string {
			fieldMeta, ok := lookupMeta.ByColumn(column)
			if ok && fieldMeta.Property != "" {
				return fieldMeta.Property
			}
			return text.ToCamel(column)
		}
	}
	return text.ToCamel
}

func validatorWithExclusionViolationError[T any](pgError *pgconn.PgError) validation.Validator {
	columnList := columns(pgError)
	mapper := columnMapper[T]()
	var v validation.Validator
	for _, column := range columnList {
		v.AddError(mapper(column), validation.CodeExclusion, nil)
	}
	return v
}
