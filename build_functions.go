package errors

func New(message string) error {
	return newBuilder().Error(message)
}

func Wrap(err error) error {
	return newBuilder().Wrap(err)
}

func Wrapf(err error, format string, args ...any) error {
	return newBuilder().Wrapf(err, format, args...)
}

func Errorf(format string, args ...any) error {
	return newBuilder().Errorf(format, args...)
}

func Join(errs ...error) error {
	return newBuilder().Join(errs...)
}

func Reason(reason string) ErrorBuilder {
	return newBuilder().Reason(reason)
}

func Domain(domain string) ErrorBuilder {
	return newBuilder().Domain(domain)
}

func WithMetadata(key, value string) ErrorBuilder {
	return newBuilder().WithMetadata(key, value)
}

func WithQuotaViolation(subject string, description string) ErrorBuilder {
	return newBuilder().WithQuotaViolation(subject, description)
}

func WithPreconditionViolation(subject string, description string) ErrorBuilder {
	return newBuilder().WithPreconditionViolation(subject, description)
}

func WithFieldViolation(field string, description string) ErrorBuilder {
	return newBuilder().WithFieldViolation(field, description)
}

func WithLocalization(localization Localization) ErrorBuilder {
	return newBuilder().WithLocalization(localization)
}

func UserID(userID string) ErrorBuilder {
	return newBuilder().UserID(userID)
}

func TenantID(tenantID string) ErrorBuilder {
	return newBuilder().TenantID(tenantID)
}

func Trace(trace string) ErrorBuilder {
	return newBuilder().Trace(trace)
}

func Span(span string) ErrorBuilder {
	return newBuilder().Span(span)
}

func WithTag(tag string) ErrorBuilder {
	return newBuilder().WithTag(tag)
}
