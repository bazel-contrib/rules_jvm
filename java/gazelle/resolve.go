package gazelle

import (
	"fmt"
	"sort"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/buildtools/build"
	lru "github.com/hashicorp/golang-lru"
)

const languageName = "java"

// Resolver satisfies the resolve.Resolver interface. It's the
// language-specific resolver extension.
//
// See resolve.Resolver for more information.
type Resolver struct {
	lang          *javaLang
	internalCache *lru.Cache
}

func NewResolver(lang *javaLang) *Resolver {
	internalCache, err := lru.New(10000)
	if err != nil {
		lang.logger.Fatal().Err(err).Msg("error creating cache")
	}

	return &Resolver{
		lang:          lang,
		internalCache: internalCache,
	}
}

func (Resolver) Name() string {
	return languageName
}

func (jr Resolver) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	log := jr.lang.logger.With().Str("step", "Imports").Str("rel", f.Pkg).Str("rule", r.Name()).Logger()

	if !isJavaLibrary(r.Kind()) {
		return nil
	}

	var out []resolve.ImportSpec
	if pkgs := r.PrivateAttr(packagesKey); pkgs != nil {
		for _, pkg := range pkgs.([]string) {
			out = append(out, resolve.ImportSpec{Lang: languageName, Imp: pkg})
		}
	}

	log.Debug().Str("out", fmt.Sprintf("%#v", out)).Msg("return")
	return out
}

func (Resolver) Embeds(r *rule.Rule, from label.Label) []label.Label {
	embedStrings := r.AttrStrings("embed")
	if isJavaProtoLibrary(r.Kind()) {
		embedStrings = append(embedStrings, r.AttrString("proto"))
	}
	embedLabels := make([]label.Label, 0, len(embedStrings))
	for _, s := range embedStrings {
		l, err := label.Parse(s)
		if err != nil {
			continue
		}
		l = l.Abs(from.Repo, from.Pkg)
		embedLabels = append(embedLabels, l)
	}
	return embedLabels
}

func (jr Resolver) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, imports interface{}, from label.Label) {
	javaImports := imports.([]string)
	if len(javaImports) == 0 {
		return
	}

	deps := make(map[string][]string)

	addDep := func(dep, imp string) {
		if _, found := deps[dep]; !found {
			deps[dep] = []string{}
		}
		deps[dep] = append(deps[dep], imp)
	}

	for _, implicitDep := range r.AttrStrings("deps") {
		addDep(implicitDep, "__implicit__")
	}

	for _, imp := range javaImports {
		dep, err := jr.convertImport(c, imp, ix, c.RepoName, from)
		if err != nil {
			jr.lang.logger.Error().Str("import", dep.String()).Err(err).Msg("error converting import")
			panic(fmt.Sprintf("error converting import: %s", err))
		}
		if dep == label.NoLabel {
			continue
		}

		addDep(dep.String(), imp)
	}

	if len(deps) > 0 {
		var labels []label.Label
		for key := range deps {
			l, err := label.Parse(key)
			if err != nil {
				jr.lang.logger.Fatal().Str("key", key).Err(err).Msg("cannot parse label")
			}

			if l.Relative && l.Name == from.Name {
				continue
			}

			labels = append(labels, l)
		}

		sort.Slice(labels, func(i, j int) bool {
			if labels[i].Relative {
				if labels[j].Relative {
					return labels[i].String() < labels[j].String()
				}

				return true
			}

			if labels[j].Relative {
				return false
			}

			return labels[i].String() < labels[j].String()
		})

		var exprs []build.Expr
		for _, l := range labels {
			exprs = append(exprs, &build.StringExpr{Value: l.String()})
		}
		r.SetAttr("deps", exprs)
	}
}

func simplifyLabel(repoName string, l label.Label, from label.Label) label.Label {
	if l.Repo == repoName || l.Repo == "" {
		if l.Pkg == from.Pkg {
			l.Relative = true
		} else {
			l.Repo = ""
		}
	}
	return l
}

func (jr *Resolver) convertImport(c *config.Config, imp string, ix *resolve.RuleIndex, repo string, from label.Label) (out label.Label, err error) {
	parsedImport := java.NewImport(imp)
	importSpec := resolve.ImportSpec{Lang: languageName, Imp: parsedImport.Pkg}
	if ol, found := resolve.FindRuleWithOverride(c, importSpec, languageName); found {
		return simplifyLabel(c.RepoName, ol, from), nil
	}

	matches := ix.FindRulesByImportWithConfig(c, importSpec, languageName)
	if len(matches) == 1 {
		return simplifyLabel(c.RepoName, matches[0].Label, from), nil
	}

	if len(matches) > 1 {
		labels := make([]string, 0, len(matches))
		for _, match := range matches {
			labels = append(labels, match.Label.String())
		}
		sort.Strings(labels)

		jr.lang.logger.Error().
			Str("pkg", parsedImport.Pkg).
			Strs("targets", labels).
			Msg("convertImport found MULTIPLE results in rule index")
	}

	if v, ok := jr.internalCache.Get(parsedImport.Pkg); ok {
		return v.(label.Label), nil
	}

	jr.lang.logger.Debug().Str("parsedImport", parsedImport.Pkg).Stringer("from", from).Msg("not found yet")

	defer func() {
		if err == nil {
			jr.internalCache.Add(parsedImport.Pkg, out)
		}
	}()

	if java.IsStdlib(imp) {
		return label.NoLabel, nil
	}

	if l, err := jr.lang.mavenResolver.Resolve(parsedImport.Pkg); err == nil {
		return simplifyLabel(c.RepoName, l, from), nil
	}

	jr.lang.logger.Warn().
		Str("package", parsedImport.Pkg).
		Str("from rule", from.String()).
		Msg("Unable to find package for import in any dependency")

	return label.NoLabel, nil
}

func isJavaLibrary(kind string) bool {
	return kind == "java_library" || isJavaProtoLibrary(kind)
}

func isJavaProtoLibrary(kind string) bool {
	return kind == "java_proto_library" || kind == "java_grpc_library"
}
