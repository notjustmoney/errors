package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
)

const (
	// StackTraceMaxDepth is the maximum number of frames to retrieve for a stack trace.
	StackTraceMaxDepth = 50
)

var (
	// packageName is the name of the package.
	packageName = reflect.TypeOf(Error{}).PkgPath()
)

type stackTrace []stackTraceFrame

func newStacktrace() stackTrace {
	var frames []stackTraceFrame

	// We loop until we have StackTraceMaxDepth frames or we run out of frames.
	// Frames from this package are skipped.
	for i := 0; len(frames) < StackTraceMaxDepth; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		file = removeGoPath(file)

		f := runtime.FuncForPC(pc)
		if f == nil {
			break
		}
		function := shortenFuncName(f)

		packageNameExamples := packageName + "/examples/"

		isGoPkg := len(runtime.GOROOT()) > 0 && strings.Contains(file, runtime.GOROOT()) // skip frames in GOROOT if it's set
		isThisPkg := strings.Contains(file, packageName)                                 // skip frames in this package
		isExamplePkg := strings.Contains(file, packageNameExamples)                      // do not skip frames in this package examples
		isTestPkg := strings.Contains(file, "_test.go")                                  // do not skip frames in tests

		if !isGoPkg && (!isThisPkg || isExamplePkg || isTestPkg) {
			frames = append(frames, stackTraceFrame{
				pc:       pc,
				file:     file,
				function: function,
				line:     line,
			})
		}
	}

	return frames
}

func (st stackTrace) Source() (string, []string) {
	if len(st) == 0 {
		return "", []string{}
	}

	firstFrame := st[0]

	header := firstFrame.String()
	body := getSourceFromFrame(firstFrame)

	return header, body
}

func (st stackTrace) Error() string {
	return st.String()
}

func (st stackTrace) StringUntilFrame(deepestFrame stackTraceFrame) string {
	var s string
	for _, frame := range st {
		if frame.file == "" {
			continue
		}

		frameStr := frame.String()
		if frameStr == "" {
			break
		}
		if frame.Equals(deepestFrame) {
			break
		}
		if s != "" && !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		s += "  --- at " + frameStr
	}
	return s
}

func (st stackTrace) String() string {
	return st.StringUntilFrame(stackTraceFrame{})
}

type stackTraceFrame struct {
	pc       uintptr
	file     string
	function string
	line     int
}

func (f *stackTraceFrame) String() string {
	s := fmt.Sprintf("%v:%v", f.file, f.line)
	if f.function != "" {
		s = fmt.Sprintf("%v:%v %v()", f.file, f.line, f.function)
	}

	return s
}

func (f *stackTraceFrame) Equals(other stackTraceFrame) bool {
	return f.file == other.file && f.function == other.function && f.line == other.line
}

func shortenFuncName(f *runtime.Func) string {
	// f.Name() is like one of these:
	// - "github.com/palantir/shield/package.FuncName"
	// - "github.com/palantir/shield/package.Receiver.MethodName"
	// - "github.com/palantir/shield/package.(*PtrReceiver).MethodName"
	longName := f.Name()

	withoutPath := longName[strings.LastIndex(longName, "/")+1:]
	withoutPackage := withoutPath[strings.Index(withoutPath, ".")+1:]

	shortName := withoutPackage
	shortName = strings.Replace(shortName, "(", "", 1)
	shortName = strings.Replace(shortName, "*", "", 1)
	shortName = strings.Replace(shortName, ")", "", 1)

	return shortName
}

/*
removeGoPath makes a path relative to one of the src directories in the $GOPATH
environment variable. If $GOPATH is empty or the input path is not contained
within any of the src directories in $GOPATH, the original path is returned. If
the input path is contained within multiple of the src directories in $GOPATH,
it is made relative to the longest one of them.
*/
func removeGoPath(path string) string {
	dirs := filepath.SplitList(os.Getenv("GOPATH"))
	// Sort in decreasing order by length so the longest matching prefix is removed
	sort.Stable(longestFirst(dirs))
	for _, dir := range dirs {
		srcDir := filepath.Join(dir, "src")
		rel, err := filepath.Rel(srcDir, path)
		// filepath.Rel can traverse parent directories, don't want those
		if err == nil && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return rel
		}
	}
	return path
}

type longestFirst []string

func (ss longestFirst) Len() int           { return len(ss) }
func (ss longestFirst) Less(i, j int) bool { return len(ss[i]) > len(ss[j]) }
func (ss longestFirst) Swap(i, j int)      { ss[i], ss[j] = ss[j], ss[i] }
