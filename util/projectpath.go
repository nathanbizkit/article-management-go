package util

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// Root is the project root directory path
	Root = filepath.Join(filepath.Dir(b), "../")
)
