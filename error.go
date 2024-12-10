package errors

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type Error struct {
	err     error
	message *string

	// error information
	reason   *string
	domain   *string
	metadata map[string]string

	// failure
	quotaViolations        []QuotaViolation
	preconditionViolations []PreconditionViolation
	fieldViolations        []FieldViolation

	// user
	userID   *string
	tenantID *string

	// tracing
	trace     *string
	span      *string
	requestID *string
	tags      []string
	time      time.Time

	// guidance
	help          Help
	resource      Resource
	localizations []Localization
	retry         Retry

	// debug
	stackTrace stackTrace
}

// Error returns the error message.
func (e *Error) Error() string {
	var sb strings.Builder
	if e.message != nil {
		sb.WriteString(*e.message)
		if e.err != nil {
			sb.WriteString(": ")
		}
	}
	if e.err != nil {
		sb.WriteString(e.err.Error())
	}

	return sb.String()
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Is(err error) bool {
	if errors.Is(e.err, err) {
		return true
	}
	return e == err
}

func (e *Error) StackTrace() string {
	var (
		blocks   []string
		topFrame stackTraceFrame
	)
	recursive(e, func(ee *Error) {
		if len(ee.stackTrace) > 0 {
			var message string
			if ee.message != nil {
				message = *ee.message
			} else {
				message = coalesceOrEmpty(
					lo.TernaryF(
						ee.err != nil,
						func() string { return ee.err.Error() },
						func() string { return "" }),
					"Error",
				)
			}
			block := fmt.Sprintf("%s\n%s", message, ee.stackTrace.StringUntilFrame(topFrame))
			blocks = append([]string{block}, blocks...)
			topFrame = (ee.stackTrace)[0]
		}
	})

	if len(blocks) == 0 {
		return ""
	}

	return "Error: " + strings.Join(blocks, "\nThrown: ")
}

// Sources returns the source fragments of the error.
func (e *Error) Sources() string {
	var blocks [][]string

	recursive(e, func(e *Error) {
		if e.stackTrace == nil || len(e.stackTrace) == 0 {
			return
		}
		header, body := e.stackTrace.Source()

		if e.message != nil {
			header = fmt.Sprintf("%s\n%s", *e.message, header)
		}

		if header != "" && len(body) > 0 {
			blocks = append(
				[][]string{append([]string{header}, body...)},
				blocks...,
			)
		}
	})

	if len(blocks) == 0 {
		return ""
	}

	return "Error: " + strings.Join(
		lo.Map(blocks, func(items []string, _ int) string {
			return strings.Join(items, "\n")
		}),
		"\n\nThrown: ",
	)
}

func (e *Error) Message() *string {
	return recursiveAttr(e, func(e *Error) *string {
		return e.message
	})
}

func (e *Error) Reason() *string {
	return recursiveAttr(e, func(e *Error) *string {
		return e.reason
	})
}

func (e *Error) Domain() *string {
	return recursiveAttr(e, func(e *Error) *string {
		return e.domain
	})
}

func (e *Error) Metadata() map[string]string {
	return recursiveAttr(e, func(e *Error) map[string]string {
		return e.metadata
	})
}

func (e *Error) QuotaViolations() []QuotaViolation {
	return recursiveAttr(e, func(e *Error) []QuotaViolation {
		return e.quotaViolations
	})
}

func (e *Error) PreconditionViolations() []PreconditionViolation {
	return recursiveAttr(e, func(e *Error) []PreconditionViolation {
		return e.preconditionViolations
	})
}

func (e *Error) FieldViolations() []FieldViolation {
	return recursiveAttr(e, func(e *Error) []FieldViolation {
		return e.fieldViolations
	})
}

func (e *Error) Trace() *string {
	trace := recursiveAttr(e, func(e *Error) *string {
		return e.trace
	})
	return lo.If(trace != nil, trace).ElseF(func() *string {
		traceID := uuid.NewString() // TODO: use a sortable unique identifier(ref: https://github.com/oklog/ulid)
		e.trace = &traceID
		return &traceID
	})
}

func (e *Error) Span() *string {
	return e.span
}

func (e *Error) RequestID() *string {
	return e.requestID
}

func (e *Error) Tags() []string {
	var tags []string

	recursive(e, func(e *Error) {
		tags = append(tags, e.tags...)
	})

	return lo.Uniq(tags)
}

func (e *Error) Time() time.Time {
	t := recursiveAttr(e, func(e *Error) time.Time {
		return e.time
	})

	return lo.TernaryF(
		t.IsZero(),
		func() time.Time {
			now := time.Now()
			e.time = now
			return now
		},
		func() time.Time {
			return t
		},
	)
}

func (e *Error) Help() Help {
	return recursiveAttr(e, func(e *Error) Help {
		return e.help
	})
}

func (e *Error) Resource() Resource {
	return recursiveAttr(e, func(e *Error) Resource {
		return e.resource
	})
}

func (e *Error) Localizations() []Localization {
	return recursiveAttr(e, func(e *Error) []Localization {
		return e.localizations
	})
}

func (e *Error) Retry() Retry {
	return recursiveAttr(e, func(e *Error) Retry {
		return e.retry
	})
}

func (e *Error) LogValue() slog.Value {
	if e == nil {
		return slog.GroupValue()
	}

	var attrs []slog.Attr
	if message := e.Message(); message != nil {
		attrs = append(attrs, slog.String("message", *e.message))
	}

	if reason := e.Reason(); reason != nil {
		attrs = append(attrs, slog.String("reason", *reason))
	}

	if domain := e.Domain(); domain != nil {
		attrs = append(attrs, slog.String("domain", *domain))
	}

	if len(e.metadata) > 0 {
		attrs = append(attrs,
			slog.Group(
				"metadata",
				lo.ToAnySlice(
					lo.MapToSlice(e.metadata, func(k string, v string) slog.Attr {
						return slog.String(k, v)
					}),
				)...,
			),
		)
	}

	if quotaViolations := e.QuotaViolations(); len(quotaViolations) > 0 {
		attrs = append(attrs,
			slog.Any(
				"quotaViolations",
				quotaViolations,
			),
		)
	}

	if preconditionViolations := e.PreconditionViolations(); len(preconditionViolations) > 0 {
		attrs = append(attrs,
			slog.Any(
				"preconditionViolations",
				preconditionViolations,
			),
		)
	}

	if fieldViolations := e.FieldViolations(); len(fieldViolations) > 0 {
		attrs = append(attrs,
			slog.Any(
				"fieldViolations",
				fieldViolations,
			))
	}

	if userID := e.userID; userID != nil {
		attrs = append(attrs, slog.String("userId", *userID))
	}

	if tenantID := e.tenantID; tenantID != nil {
		attrs = append(attrs, slog.String("tenantId", *tenantID))
	}

	if trace := e.Trace(); trace != nil {
		attrs = append(attrs, slog.String("trace", *trace))
	}

	if span := e.Span(); span != nil {
		attrs = append(attrs, slog.String("span", *span))
	}

	if requestID := e.RequestID(); requestID != nil {
		attrs = append(attrs, slog.String("requestId", *requestID))
	}

	if tags := e.Tags(); len(tags) > 0 {
		attrs = append(attrs, slog.Any("tags", tags))
	}

	if time := e.Time(); !time.IsZero() {
		attrs = append(attrs, slog.Time("time", time))
	}

	if help := e.Help(); lo.IsNotEmpty(help) {
		attrs = append(attrs, slog.Group(
			"help",
			slog.String("description", help.Description),
			slog.String("url", help.URL),
		))
	}

	if resource := e.Resource(); lo.IsNotEmpty(resource) {
		attrs = append(attrs, slog.Group(
			"resource",
			slog.String("type", resource.Type),
			slog.String("name", resource.Name),
			slog.String("owner", resource.Owner),
			slog.String("description", resource.Description),
		))
	}

	if localizations := e.Localizations(); len(localizations) > 0 {
		attrs = append(attrs, slog.Group(
			"localizations",
			lo.ToAnySlice(localizations)...,
		))
	}

	if retry := e.Retry(); lo.IsNotEmpty(retry) {
		attrs = append(attrs, slog.Group(
			"retry",
			slog.String("delay", retry.Delay.String()),
		))
	}

	if st := e.StackTrace(); st != "" {
		attrs = append(attrs, slog.String("stackTrace", st))
	}

	return slog.GroupValue(attrs...)
}

func (e *Error) Format(s fmt.State, verb rune) {
	if verb == 'v' && s.Flag('+') {
		fmt.Fprint(s, e.formatVerbose())
	} else {
		fmt.Fprint(s, e.formatSummary())
	}
}

func (e *Error) formatVerbose() string {
	var sb strings.Builder
	sb.WriteString("Error: ")
	sb.WriteString(e.Error())
	sb.WriteString("\n")

	if reason := e.Reason(); reason != nil {
		sb.WriteString("Reason: ")
		sb.WriteString(*reason)
		sb.WriteString("\n")
	}

	if domain := e.Domain(); domain != nil {
		sb.WriteString("Domain: ")
		sb.WriteString(*domain)
		sb.WriteString("\n")
	}

	if metadata := e.Metadata(); len(metadata) > 0 {
		sb.WriteString("Metadata:\n")
		for k, v := range metadata {
			printTab(&sb)
			sb.WriteString(k)
			sb.WriteString(": ")
			sb.WriteString(v)
			sb.WriteString("\n")
		}
	}

	if quotaViolations := e.QuotaViolations(); len(quotaViolations) > 0 {
		sb.WriteString("QuotaViolations:\n")
		for _, violation := range quotaViolations {
			printTab(&sb)
			sb.WriteString("QuotaViolation:\n")
			printTab(&sb)
			printTab(&sb)
			sb.WriteString("Subject: ")
			sb.WriteString(violation.Subject)
			sb.WriteString("\n")
			printTab(&sb)
			printTab(&sb)
			sb.WriteString("Description: ")
			sb.WriteString(violation.Description)
			sb.WriteString("\n")
		}
	}

	if preconditionViolations := e.PreconditionViolations(); len(preconditionViolations) > 0 {
		sb.WriteString("PreconditionViolations:\n")
		for _, violation := range preconditionViolations {
			printTab(&sb)
			sb.WriteString("PreconditionViolation:\n")
			printTab(&sb)
			printTab(&sb)
			sb.WriteString("Type: ")
			sb.WriteString(violation.Type)
			sb.WriteString("\n")
			printTab(&sb)
			printTab(&sb)
			sb.WriteString("Subject: ")
			sb.WriteString(violation.Subject)
			sb.WriteString("\n")
			printTab(&sb)
			printTab(&sb)
			sb.WriteString("Description: ")
			sb.WriteString(violation.Description)
			sb.WriteString("\n")
		}
	}

	if fieldViolations := e.FieldViolations(); len(fieldViolations) > 0 {
		sb.WriteString("FieldViolations:\n")
		for _, violation := range fieldViolations {
			printTab(&sb)
			sb.WriteString("FieldViolation:\n")
			printTab(&sb)
			printTab(&sb)
			sb.WriteString("Field: ")
			sb.WriteString(violation.Field)
			sb.WriteString("\n")
			printTab(&sb)
			printTab(&sb)
			sb.WriteString("Description: ")
			sb.WriteString(violation.Description)
			sb.WriteString("\n")
		}
	}

	if userID := e.userID; userID != nil {
		sb.WriteString("UserId: ")
		sb.WriteString(*userID)
		sb.WriteString("\n")
	}

	if tenantID := e.tenantID; tenantID != nil {
		sb.WriteString("TenantId: ")
		sb.WriteString(*tenantID)
		sb.WriteString("\n")
	}

	if trace := e.Trace(); trace != nil {
		sb.WriteString("Trace: ")
		sb.WriteString(*trace)
		sb.WriteString("\n")
	}

	if span := e.Span(); span != nil {
		sb.WriteString("Span: ")
		sb.WriteString(*span)
		sb.WriteString("\n")
	}

	if requestID := e.RequestID(); requestID != nil {
		sb.WriteString("RequestId: ")
		sb.WriteString(*requestID)
		sb.WriteString("\n")
	}

	if tags := e.Tags(); len(tags) > 0 {
		sb.WriteString("Tags: ")
		sb.WriteString("[")
		sb.WriteString(strings.Join(tags, ", "))
		sb.WriteString("]\n")
	}

	if time := e.Time(); !time.IsZero() {
		sb.WriteString("Time: ")
		sb.WriteString(time.String())
		sb.WriteString("\n")
	}

	if help := e.Help(); lo.IsNotEmpty(help) {
		sb.WriteString("Help:\n")
		printTab(&sb)
		sb.WriteString("Description: ")
		sb.WriteString(help.Description)
		printTab(&sb)
		sb.WriteString("	URL: ")
		sb.WriteString(help.URL)
		sb.WriteString("\n")
	}

	if resource := e.Resource(); lo.IsNotEmpty(resource) {
		sb.WriteString("Resource:\n")
		printTab(&sb)
		sb.WriteString("Type: ")
		sb.WriteString(resource.Type)
		printTab(&sb)
		sb.WriteString("Name: ")
		sb.WriteString(resource.Name)
		if resource.Owner != "" {
			printTab(&sb)
			sb.WriteString("Owner: ")
			sb.WriteString(resource.Owner)
		}
		if resource.Description != "" {
			printTab(&sb)
			sb.WriteString("Description: ")
			sb.WriteString(resource.Description)
		}
		sb.WriteString("\n")
	}

	if localizations := e.Localizations(); len(localizations) > 0 {
		sb.WriteString("Localizations:\n")
		for _, l := range localizations {
			printTab(&sb)
			sb.WriteString("Locale: ")
			sb.WriteString(l.Locale)
			printTab(&sb)
			sb.WriteString("Message: ")
			sb.WriteString(l.Message)
			sb.WriteString("\n")
		}
	}

	if retry := e.Retry(); lo.IsNotEmpty(retry) {
		sb.WriteString("Retry:\n")
		printTab(&sb)
		sb.WriteString("Delay: ")
		sb.WriteString(retry.Delay.String())
		sb.WriteString("\n")
	}

	if st := e.StackTrace(); st != "" {
		sb.WriteString(st)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (e *Error) formatSummary() string {
	return e.Error()
}
