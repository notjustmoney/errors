package errors

import (
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func a() stackTrace {
	return b()
}

func b() stackTrace {
	return c()
}

func c() stackTrace {
	return d()
}

func d() stackTrace {
	return e()
}

func e() stackTrace {
	return f()
}

func f() stackTrace {
	return newStacktrace()
}

func TestStackTrace(t *testing.T) {
	is := assert.New(t)

	st := a()

	is.NotNil(st)
	bi, ok := debug.ReadBuildInfo()
	is.True(ok)

	if st == nil {
		return
	}

	for _, frame := range st {
		is.Truef(strings.Contains(frame.file, bi.Path), "frame file %s should contain %s", frame.file, bi.Path)
		is.Len(st, 7, "expected 7 frames")
		if len(st) != 7 {
			return
		}
		is.Equal("f", st[0].function)
		is.Equal("e", st[1].function)
		is.Equal("d", st[2].function)
		is.Equal("c", st[3].function)
		is.Equal("b", st[4].function)
		is.Equal("a", st[5].function)
		is.Equal("TestStackTrace", st[6].function)
	}
}
