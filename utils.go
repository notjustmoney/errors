package errors

import (
	"errors"
	"strings"

	"github.com/samber/lo"
)

func recursive(err *Error, tap func(*Error)) {
	if err == nil {
		return
	}

	tap(err)

	if err.err == nil {
		return
	}

	var child *Error
	if errors.As(err.err, &child) {
		recursive(child, tap)
	}
}

func recursiveAttr[T any](err *Error, attr func(*Error) T) T {
	if err == nil {
		var zero T
		return zero
	}

	if err.err == nil {
		return attr(err)
	}

	var child *Error
	if errors.As(err.err, &child) {
		return recursiveAttr[T](child, attr)
	}

	return attr(err)
}

func deepCopyPtr[T any](p *T) *T {
	if p == nil {
		return nil
	}

	return lo.ToPtr(lo.FromPtr(p))
}

func coalesceOrEmpty[T comparable](v ...T) T {
	result, _ := lo.Coalesce(v...)
	return result
}

func printTab(sb *strings.Builder) {
	sb.WriteString("	")
}
