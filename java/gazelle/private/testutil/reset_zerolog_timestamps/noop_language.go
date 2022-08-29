package reset_zerolog_timestamps

import (
	"github.com/bazelbuild/bazel-gazelle/language"
)

func NewLanguage() language.Language {
	return &noopLanguage{}
}

type noopLanguage struct {
	language.BaseLang
}
