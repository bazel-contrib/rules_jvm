/*
 * Copyright 2023 Aspect Build Systems, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package treesitter

import (
	"context"
	"fmt"
	"log"
	"path"

	"github.com/bazel-contrib/rules_jvm/gazelle/common/treesitter/grammars/java"
	"github.com/bazel-contrib/rules_jvm/gazelle/common/treesitter/grammars/json"
	"github.com/bazel-contrib/rules_jvm/gazelle/common/treesitter/grammars/kotlin"
	sitter "github.com/smacker/go-tree-sitter"
)

type LanguageGrammar string

const (
	Kotlin      LanguageGrammar = "kotlin"
	JSON                        = "json"
	Java                        = "java"
)

type ASTQueryResult interface {
	Captures() map[string]string
}

type AST interface {
	Query(query TreeQuery) <-chan ASTQueryResult
	QueryErrors() []error

	// Wrapper utils
	// TODO: delete
	QueryStrings(query TreeQuery, returnVar string) []string
	RootNode() *sitter.Node
}
type treeAst struct {
	lang       LanguageGrammar
	filePath   string
	sourceCode []byte

	sitterTree *sitter.Tree
}

var _ AST = (*treeAst)(nil)

func (tree *treeAst) String() string {
	return fmt.Sprintf("treeAst{\n lang: %q,\n filePath: %q,\n AST:\n  %v\n}", tree.lang, tree.filePath, tree.sitterTree.RootNode().String())
}

func toSitterLanguage(lang LanguageGrammar) *sitter.Language {
	switch lang {
	case Java:
		return java.GetLanguage()
	case JSON:
		return json.GetLanguage()
	case Kotlin:
		return kotlin.GetLanguage()
	}

	log.Panicf("Unknown LanguageGrammar %q", lang)
	return nil
}

func PathToLanguage(p string) LanguageGrammar {
	return extensionToLanguage(path.Ext(p))
}

// Based on https://github.com/github-linguist/linguist/blob/master/lib/linguist/languages.yml
var EXT_LANGUAGES = map[string]LanguageGrammar{
	"kt":  Kotlin,
	"ktm": Kotlin,
	"kts": Kotlin,
	"java": Java,
	"jav":  Java,
	"jsh":  Java,
	"json": JSON,
}

// In theory, this is a mirror of
// https://github.com/github-linguist/linguist/blob/master/lib/linguist/languages.yml
func extensionToLanguage(ext string) LanguageGrammar {
	var lang, found = EXT_LANGUAGES[ext[1:]]

	// TODO: allow override or fallback language for files
	if !found {
		log.Panicf("Unknown source file extension %q", ext)
	}

	return lang
}

func ParseSourceCode(lang LanguageGrammar, filePath string, sourceCode []byte) (AST, error) {
	ctx := context.Background()

	parser := sitter.NewParser()
	parser.SetLanguage(toSitterLanguage(lang))

	tree, err := parser.ParseCtx(ctx, nil, sourceCode)
	if err != nil {
		return nil, err
	}

	return &treeAst{lang: lang, filePath: filePath, sourceCode: sourceCode, sitterTree: tree}, nil
}
