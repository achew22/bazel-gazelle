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

import "path/filepath"

// typescriptPackage contains metadata for a set of .ts files that have the
// same package name. This translates to a typescript_library rule.
type typescriptPackage struct {
	name    string
	files   map[string]FileInfo
	imports map[string]bool
	options map[string]string
}

func newTypeScriptPackage(name string) *typescriptPackage {
	return &typescriptPackage{
		name:    name,
		files:   map[string]FileInfo{},
		imports: map[string]bool{},
		options: map[string]string{},
	}
}

func (p *typescriptPackage) addFile(info FileInfo) {
	p.files[info.Name] = info
	for _, imp := range info.Imports {
		p.imports[imp] = true
	}
	for _, opt := range info.Options {
		p.options[opt.Key] = opt.Value
	}
}

func (p *typescriptPackage) addGenFile(dir, name string) {
	p.files[name] = FileInfo{
		Name: name,
		Path: filepath.Join(dir, filepath.FromSlash(name)),
	}
}
