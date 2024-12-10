package errors

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type ErrorBuilder Error

func newBuilder() ErrorBuilder {
	return ErrorBuilder{
		err:     nil,
		message: nil,

		reason:   nil,
		domain:   nil,
		metadata: nil,

		quotaViolations:        nil,
		preconditionViolations: nil,
		fieldViolations:        nil,

		userID:   nil,
		tenantID: nil,

		trace:     nil,
		span:      nil,
		requestID: nil,
		tags:      nil,
		time:      time.Now(),

		help:          Help{},
		resource:      Resource{},
		localizations: nil,
		retry:         Retry{},

		stackTrace: nil,
	}
}

func (e ErrorBuilder) Wrap(err error) error {
	e2 := e.wrap(err)
	if e2 == nil {
		return nil
	}
	return (*Error)(e2)
}

func (e ErrorBuilder) Wrapf(err error, format string, args ...any) error {
	e2 := e.wrap(err)
	if e2 == nil {
		return nil
	}
	e2.message = lo.ToPtr(fmt.Errorf(format, args...).Error())
	return (*Error)(e2)
}

func (e ErrorBuilder) Error(message string) error {
	e2 := e.deepCopy()
	e2.message = &message
	e2.stackTrace = newStacktrace()
	return (*Error)(&e2)
}

func (e ErrorBuilder) Errorf(format string, args ...any) error {
	e2 := e.deepCopy()
	e2.err = fmt.Errorf(format, args...)
	e2.stackTrace = newStacktrace()
	return (*Error)(&e2)
}

func (e ErrorBuilder) Join(errs ...error) error {
	return e.Wrap(errors.Join(errs...))
}

func (e ErrorBuilder) wrap(err error) *ErrorBuilder {
	if err == nil {
		return nil
	}
	e2 := e.deepCopy()
	e2.err = err
	if e2.span == nil {
		e2.span = lo.ToPtr(uuid.NewString()) // TODO: use a unique identifier
	}
	e2.stackTrace = newStacktrace()

	return &e2
}

func (e ErrorBuilder) Reason(reason string) ErrorBuilder {
	e.reason = &reason
	return e
}

func (e ErrorBuilder) Domain(domain string) ErrorBuilder {
	e.domain = &domain
	return e
}

func (e ErrorBuilder) WithMetadata(key, value string) ErrorBuilder {
	if e.metadata == nil {
		e.metadata = map[string]string{}
	}
	e.metadata[key] = value
	return e
}

func (e ErrorBuilder) WithQuotaViolation(subject string, description string) ErrorBuilder {
	e.quotaViolations = append(e.quotaViolations, QuotaViolation{
		Subject:     subject,
		Description: description,
	})
	return e
}

func (e ErrorBuilder) WithPreconditionViolation(subject string, description string) ErrorBuilder {
	e.preconditionViolations = append(e.preconditionViolations, PreconditionViolation{
		Subject:     subject,
		Description: description,
	})
	return e
}

func (e ErrorBuilder) WithFieldViolation(field string, description string) ErrorBuilder {
	e.fieldViolations = append(e.fieldViolations, FieldViolation{
		Field:       field,
		Description: description,
	})
	return e
}

func (e ErrorBuilder) UserID(userID string) ErrorBuilder {
	e.userID = &userID
	return e
}

func (e ErrorBuilder) TenantID(tenantID string) ErrorBuilder {
	e.tenantID = &tenantID
	return e
}

func (e ErrorBuilder) Trace(trace string) ErrorBuilder {
	e.trace = &trace
	return e
}

func (e ErrorBuilder) Span(span string) ErrorBuilder {
	e.span = &span
	return e
}

func (e ErrorBuilder) RequestID(requestID string) ErrorBuilder {
	e.requestID = &requestID
	return e
}

func (e ErrorBuilder) WithTag(tag string) ErrorBuilder {
	e.tags = append(e.tags, tag)
	return e
}

func (e ErrorBuilder) Help(help Help) ErrorBuilder {
	e.help = help
	return e
}

func (e ErrorBuilder) Resource(resource Resource) ErrorBuilder {
	e.resource = resource
	return e
}

func (e ErrorBuilder) WithLocalization(localization Localization) ErrorBuilder {
	e.localizations = append(e.localizations, localization)
	return e
}

func (e ErrorBuilder) Retry(retry Retry) ErrorBuilder {
	e.retry = retry
	return e
}

func (e ErrorBuilder) deepCopy() ErrorBuilder {
	return ErrorBuilder{
		err:      e.err,
		message:  deepCopyPtr(e.message),
		reason:   deepCopyPtr(e.reason),
		domain:   deepCopyPtr(e.domain),
		metadata: lo.Assign(map[string]string{}, e.metadata),

		quotaViolations:        lo.Slice(e.quotaViolations, 0, len(e.quotaViolations)),
		preconditionViolations: lo.Slice(e.preconditionViolations, 0, len(e.preconditionViolations)),
		fieldViolations:        lo.Slice(e.fieldViolations, 0, len(e.fieldViolations)),

		userID:   deepCopyPtr(e.userID),
		tenantID: deepCopyPtr(e.tenantID),

		trace: deepCopyPtr(e.trace),
		span:  deepCopyPtr(e.span),
		tags:  lo.Slice(e.tags, 0, len(e.tags)),

		help:          e.help,
		resource:      e.resource,
		localizations: lo.Slice(e.localizations, 0, len(e.localizations)),
		retry:         e.retry,

		stackTrace: nil,
	}
}
