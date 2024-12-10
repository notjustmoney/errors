package errors

import (
	"log/slog"
	"time"
)

type Retry struct {
	Delay time.Duration
}

type Localization struct {
	Locale  string // TODO: use https://www.rfc-editor.org/rfc/bcp/bcp47.txt
	Message string
}

type Resource struct {
	Type        string
	Name        string
	Owner       string
	Description string
}

type Help struct {
	Description string
	URL         string
}

type QuotaViolation struct {
	Subject     string
	Description string
}

func (v QuotaViolation) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("subject", v.Subject),
		slog.String("description", v.Description),
	)
}

type PreconditionViolation struct {
	Type        string
	Subject     string
	Description string
}

func (v PreconditionViolation) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("type", v.Type),
		slog.String("subject", v.Subject),
		slog.String("description", v.Description),
	)
}

type FieldViolation struct {
	Field       string
	Description string
}

func (v FieldViolation) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("field", v.Field),
		slog.String("description", v.Description),
	)
}
