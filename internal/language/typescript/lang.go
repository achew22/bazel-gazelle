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

// Package typescript provides support for typescriptcol buffer rules.
// It generates typescript_library rules only (not go_typescript_library or any other
// language-specific implementations).
//
// Configuration
//
// Configuration is largely controlled by Mode. In disable mode, typescript rules are
// left alone (neither generated nor deleted). In legacy mode, filegroups are
// emitted containing typescripts. In default mode, typescript_library rules are
// emitted. The typescript mode may be set with the -typescript command line flag or the
// "# gazelle:typescript" directive.
//
// The configuration is largely public, and other languages may depend on it.
// For example, go uses Mode to determine whether to generate go_typescript_library
// rules and ignore static .pb.go files.
//
// Rule generation
//
// Currently, Gazelle generates at most one typescript_library per directory. TypeScripts
// in the same package are grouped together into a typescript_library. If there are
// sources for multiple packages, the package name that matches the directory
// name will be chosen; if there is no such package, an error will be printed.
// We expect to provide support for multiple typescript_libraries in the future
// when Go has support for multiple packages and we have better rule matching.
// The generated typescript_library will be named after the directory, not the
// typescript or the package. For example, for foo/bar/baz.ts, a typescript_library
// rule will be generated named //foo/bar:bar_typescript.
//
// Dependency resolution
//
// typescript_library rules are indexed by their srcs attribute. Gazelle attempts
// to resolve typescript imports (e.g., import foo/bar/bar.ts) to the
// typescript_library that contains the named source file
// (e.g., //foo/bar:bar_typescript). If no indexed typescript_library provides the source
// file, Gazelle will guess a label, following conventions.
//
// No attempt is made to resolve typescripts to rules in external repositories,
// since there's no indication that a typescript import comes from an external
// repository. In the future, build files in external repos will be indexed,
// so we can support this (#12).
//
// Gazelle has special cases for Well Known Types (i.e., imports of the form
// google/typescriptbuf/*.ts). These are resolved to rules in
// @com_google_typescriptbuf.
package typescript

import "github.com/bazelbuild/bazel-gazelle/internal/language"

const typescriptName = "typescript"

type typescriptLang struct{}

func (_ *typescriptLang) Name() string { return typescriptName }

func New() language.Language {
	return &typescriptLang{}
}
