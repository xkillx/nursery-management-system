package application

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	domainerrors "nursery-management-system/api/internal/platform/errors"
)

const (
	pgCodeFKViolation      = "23503"
	pgCodeUniqueViolation  = "23505"
	pgCodeCheckViolation   = "23514"
	pgCodeNotNullViolation = "23502"
	pgCodeStringTruncation = "22001"
)

// mapExecTxError converts an error returned from an ExecTx callback into a
// DomainError. Known Postgres error codes are mapped to user-facing validation
// or conflict errors so the API returns 400/409 instead of a generic 500.
// Unrecognised errors are wrapped as internal_error.
func mapExecTxError(err error) *domainerrors.DomainError {
	if err == nil {
		return nil
	}

	// Already a domain error — return as-is.
	var de *domainerrors.DomainError
	if errors.As(err, &de) {
		return de
	}

	// Attempt to extract a pgconn.PgError from the chain.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if mapped := mapPgError(pgErr); mapped != nil {
			return mapped
		}
	}

	return domainerrors.Internal(err)
}

// mapPgError converts a known Postgres error to a domain error.
// Returns nil if the error code is not recognised.
func mapPgError(pgErr *pgconn.PgError) *domainerrors.DomainError {
	switch pgErr.Code {
	case pgCodeFKViolation:
		field := extractFKField(pgErr.ConstraintName)
		return domainerrors.Validation(
			fmt.Sprintf("Referenced record not found (%s).", field),
			field,
		)

	case pgCodeUniqueViolation:
		field := extractUniqueField(pgErr.ConstraintName)
		return domainerrors.Conflict(
			strings.ReplaceAll(pgErr.ConstraintName, "_", "")+"_conflict",
			fmt.Sprintf("A record with this %s already exists.", field),
		)

	case pgCodeCheckViolation:
		field := extractCheckField(pgErr.ConstraintName)
		return domainerrors.Validation(
			fmt.Sprintf("Invalid value for %s.", field),
			field,
		)

	case pgCodeNotNullViolation:
		field := extractColumnFromMessage(pgErr.Message)
		return domainerrors.Validation(
			fmt.Sprintf("%s is required.", field),
			field,
		)

	case pgCodeStringTruncation:
		field := extractColumnFromMessage(pgErr.Message)
		return domainerrors.Validation(
			fmt.Sprintf("%s value is too long.", field),
			field,
		)
	}

	return nil
}

// extractFKField heuristically derives the user-facing field name from a
// Postgres foreign-key constraint name. Constraint names follow the pattern
// <table>_<column>_fkey.
func extractFKField(constraint string) string {
	parts := strings.Split(constraint, "_")
	// Remove trailing "fkey" and join remaining parts as field name.
	if len(parts) >= 2 && parts[len(parts)-1] == "fkey" {
		return strings.Join(parts[1:len(parts)-1], "_")
	}
	return constraint
}

// extractUniqueField heuristically derives the user-facing field name from a
// unique-constraint name.
func extractUniqueField(constraint string) string {
	parts := strings.Split(constraint, "_")
	// Remove trailing "key" and join remaining parts.
	if len(parts) >= 2 && parts[len(parts)-1] == "key" {
		return strings.Join(parts[1:len(parts)-1], "_")
	}
	return constraint
}

// extractCheckField derives a field name from a check-constraint name.
func extractCheckField(constraint string) string {
	parts := strings.Split(constraint, "_")
	if len(parts) >= 2 && parts[len(parts)-1] == "check" {
		return strings.Join(parts[1:len(parts)-1], "_")
	}
	return constraint
}

// extractColumnFromMessage parses a column name from a pgx error message like
// 'null value in column "column_name" of relation "table" violates not-null constraint'.
func extractColumnFromMessage(msg string) string {
	// Look for quoted identifiers in the message.
	start := strings.Index(msg, `"`)
	if start == -1 {
		return ""
	}
	end := strings.Index(msg[start+1:], `"`)
	if end == -1 {
		return ""
	}
	return msg[start+1 : start+1+end]
}
