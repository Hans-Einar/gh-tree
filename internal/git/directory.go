package git

import (
	"strings"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
)

func nativeComponent(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return diagnostic(api.Invalid, "InvalidNativeComponent", "A native relative operation requires one exact path component.")
	}
	return nil
}
