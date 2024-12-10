package errors_test

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/notjustmoney/errors"
)

func TestErrorIs(t *testing.T) {
	is := assert.New(t)

	err := errors.Errorf("Error: %w", fs.ErrExist)
	is.True(errors.Is(err, fs.ErrExist))

	err = errors.Wrap(fs.ErrExist)
	is.True(errors.Is(err, fs.ErrExist))

	err = errors.Wrapf(fs.ErrExist, "Error: %w", assert.AnError)
	is.True(errors.Is(err, fs.ErrExist))

	err = errors.Join(fs.ErrExist, assert.AnError)
	is.True(errors.Is(err, fs.ErrExist))
	err = errors.Join(assert.AnError, fs.ErrExist)
	is.True(errors.Is(err, fs.ErrExist))
}

func TestWrap(t *testing.T) {
	is := assert.New(t)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       nil,
		ReplaceAttr: nil,
	}))

	err := a()
	is.Error(err)
	logger.Error(
		"test wrap",
		slog.Any("err", err),
	)

	fmt.Printf("%+v\n", err)
}

func a() error {
	return errors.
		Trace("trace").
		Wrap(b())
}

func b() error {
	return errors.Wrap(c())
}

func c() error {
	return errors.
		WithLocalization(errors.Localization{
			Locale:  "ko",
			Message: "refresh token이 유효하지 않습니다.",
		}).
		Wrap(d())
}

func d() error {
	return errors.
		Reason("ERROR_REASON_D").
		Domain("identity").
		WithFieldViolation("refreshToken", "refresh-token-string").
		WithTag("identity").
		Trace("trace").
		Wrap(e())

}

func e() error {
	return errors.Wrap(f())
}

func f() error {
	return errors.
		Reason("ERROR_REASON_INVALID_REFRESH_TOKEN").
		Domain("identity").
		WithMetadata("refreshToken", "refresh-token-string").
		WithFieldViolation("refreshToken", "refresh-token-string").
		WithTag("identity").
		Errorf("Invalid refresh token")
}
