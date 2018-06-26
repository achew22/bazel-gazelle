/* Copyright 2018 The Bazel Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package typescript

import (
	"errors"
	"fmt"
	"log"
	"path"

	"github.com/bazelbuild/bazel-gazelle/internal/config"
	"github.com/bazelbuild/bazel-gazelle/internal/label"
	"github.com/bazelbuild/bazel-gazelle/internal/repos"
	"github.com/bazelbuild/bazel-gazelle/internal/resolve"
	"github.com/bazelbuild/bazel-gazelle/internal/rule"
)

func (_ *typescriptLang) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	rel := f.Rel(c.RepoRoot)
	srcs := r.AttrStrings("srcs")
	imports := make([]resolve.ImportSpec, len(srcs))
	for i, src := range srcs {
		imports[i] = resolve.ImportSpec{Lang: "typescript", Imp: path.Join(rel, src)}
	}
	return imports
}

func (_ *typescriptLang) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

func (_ *typescriptLang) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repos.RemoteCache, r *rule.Rule, from label.Label) {
	importsRaw := r.PrivateAttr(config.GazelleImportsKey)
	if importsRaw == nil {
		// may not be set in tests.
		return
	}
	imports := importsRaw.([]string)
	r.DelAttr("deps")
	deps := make([]string, 0, len(imports))
	for _, imp := range imports {
		l, err := resolveTypeScript(ix, r, imp, from)
		if err == skipImportError {
			continue
		} else if err != nil {
			log.Print(err)
		} else {
			l = l.Rel(from.Repo, from.Pkg)
			deps = append(deps, l.String())
		}
	}
	if len(deps) > 0 {
		r.SetAttr("deps", deps)
	}
}

var (
	skipImportError = errors.New("std import")
	notFoundError   = errors.New("not found")
)

func resolveTypeScript(ix *resolve.RuleIndex, r *rule.Rule, imp string, from label.Label) (label.Label, error) {
	if l, err := resolveWithIndex(ix, imp, from); err == nil || err == skipImportError {
		return l, err
	} else if err != notFoundError {
		return label.NoLabel, err
	}

	rel := path.Dir(imp)
	if rel == "." {
		rel = ""
	}
	name := RuleName(rel)
	print("Name: %s. Rel: %s\n", name, rel)
	return label.New("", rel, name), nil
}

func resolveWithIndex(ix *resolve.RuleIndex, imp string, from label.Label) (label.Label, error) {
	matches := ix.FindRulesByImport(resolve.ImportSpec{Lang: "typescript", Imp: imp}, "typescript")
	if len(matches) == 0 {
		return label.NoLabel, notFoundError
	}
	if len(matches) > 1 {
		return label.NoLabel, fmt.Errorf("multiple rules (%s and %s) may be imported with %q from %s", matches[0].Label, matches[1].Label, imp, from)
	}
	if from.Equal(matches[0].Label) {
		return label.NoLabel, skipImportError
	}
	return matches[0].Label, nil
}
