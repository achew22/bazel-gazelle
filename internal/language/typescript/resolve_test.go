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
	"path/filepath"
	"strings"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/internal/config"
	"github.com/bazelbuild/bazel-gazelle/internal/label"
	"github.com/bazelbuild/bazel-gazelle/internal/repos"
	"github.com/bazelbuild/bazel-gazelle/internal/resolve"
	"github.com/bazelbuild/bazel-gazelle/internal/rule"
	bzl "github.com/bazelbuild/buildtools/build"
)

func TestResolveTypeScript(t *testing.T) {
	type buildFile struct {
		rel, content string
	}
	type testCase struct {
		desc      string
		index     []buildFile
		old, want string
	}
	for _, tc := range []testCase{
		{
			desc: "well_known",
			index: []buildFile{{
				rel: "google/typescriptbuf",
				content: `
ts_library(
    name = "bad_typescript",
    srcs = ["any.ts"],
)
`,
			}},
			old: `
ts_library(
    name = "dep_typescript",
    _imports = ["google/typescriptbuf/any.ts"],
)
`,
			want: `
ts_library(
    name = "dep_typescript",
    deps = ["@com_google_typescriptbuf//:any_typescript"],
)
`,
		}, {
			desc: "index",
			index: []buildFile{{
				rel: "foo",
				content: `
ts_library(
    name = "foo_typescript",
    srcs = ["foo.ts"],
)
`,
			}},
			old: `
ts_library(
    name = "dep_typescript",
    _imports = ["foo/foo.ts"],
)
`,
			want: `
ts_library(
    name = "dep_typescript",
    deps = ["//foo:foo_typescript"],
)
`,
		}, {
			desc: "index_local",
			old: `
ts_library(
    name = "foo_typescript",
    srcs = ["foo.ts"],
)

ts_library(
    name = "dep_typescript",
    _imports = ["test/foo.ts"],
)
`,
			want: `
ts_library(
    name = "foo_typescript",
    srcs = ["foo.ts"],
)

ts_library(
    name = "dep_typescript",
    deps = [":foo_typescript"],
)
`,
		}, {
			desc: "index_ambiguous",
			index: []buildFile{{
				rel: "foo",
				content: `
ts_library(
    name = "a_typescript",
    srcs = ["foo.ts"],
)

ts_library(
    name = "b_typescript",
    srcs = ["foo.ts"],
)
`,
			}},
			old: `
ts_library(
    name = "dep_typescript",
    _imports = ["foo/foo.ts"],
)
`,
			want: `ts_library(name = "dep_typescript")`,
		}, {
			desc: "index_self",
			old: `
ts_library(
    name = "dep_typescript",
    srcs = ["foo.ts"],
    _imports = ["test/foo.ts"],
)
`,
			want: `
ts_library(
    name = "dep_typescript",
    srcs = ["foo.ts"],
)
`,
		}, {
			desc: "unknown",
			old: `
ts_library(
    name = "dep_typescript",
    _imports = ["foo/bar/unknown.ts"],
)
`,
			want: `
ts_library(
    name = "dep_typescript",
    deps = ["//foo/bar:bar_typescript"],
)
`,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			c := config.New()
			c.Exts[typescriptName] = &TypeScriptConfig{}
			lang := New()
			ix := resolve.NewRuleIndex(map[string]resolve.Resolver{"ts_library": lang})
			rc := (*repos.RemoteCache)(nil)
			for _, bf := range tc.index {
				f, err := rule.LoadData(filepath.Join(bf.rel, "BUILD.bazel"), []byte(bf.content))
				if err != nil {
					t.Fatal(err)
				}
				for _, r := range f.Rules {
					ix.AddRule(c, r, f)
				}
			}
			f, err := rule.LoadData("test/BUILD.bazel", []byte(tc.old))
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range f.Rules {
				convertImportsAttr(r)
				ix.AddRule(c, r, f)
			}
			ix.Finish()
			for _, r := range f.Rules {
				lang.Resolve(c, ix, rc, r, label.New("", "test", r.Name()))
			}
			f.Sync()
			got := strings.TrimSpace(string(bzl.Format(f.File)))
			want := strings.TrimSpace(tc.want)
			if got != want {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func convertImportsAttr(r *rule.Rule) {
	value := r.AttrStrings("_imports")
	if value == nil {
		value = []string(nil)
	}
	r.DelAttr("_imports")
	r.SetPrivateAttr(config.GazelleImportsKey, value)
}
