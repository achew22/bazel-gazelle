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

import "github.com/bazelbuild/bazel-gazelle/internal/rule"

var typescriptKinds = map[string]rule.KindInfo{
	"ts_library": {
		NonEmptyAttrs:  map[string]bool{"srcs": true},
		MergeableAttrs: map[string]bool{"srcs": true},
		ResolveAttrs:   map[string]bool{"deps": true},
	},
}

var typescriptLoads = []rule.LoadInfo{
	{
		Name: "@build_bazel_rules_typescript//:defs.bzl",
		Symbols: []string{
			"ts_library",
		},
	},
}

func (_ *typescriptLang) Kinds() map[string]rule.KindInfo { return typescriptKinds }
func (_ *typescriptLang) Loads() []rule.LoadInfo          { return typescriptLoads }
