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
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"sort"
)

// FileInfo contains metadata extracted from a .ts file.
type FileInfo struct {
	Path, Name string

	PackageName string

	Options []Option
	Imports []string
}

// Option represents a top-level option statement in a .ts file. Only
// string options are supported for now.
type Option struct {
	Key, Value string
}

var typescriptRe = buildTypeScriptRegexp()

func print(v string, a ...interface{}) {
	fmt.Printf(fmt.Sprintf("\u001b[31;1m%s\u001b[0m", v), a...)
}

func typescriptFileInfo(dir, name string) FileInfo {
	info := FileInfo{
		Path: filepath.Join(dir, name),
		Name: name,
	}
	content, err := ioutil.ReadFile(info.Path)
	if err != nil {
		log.Printf("%s: error reading typescript file: %v", info.Path, err)
		return info
	}

	for _, match := range typescriptRe.FindAllStringSubmatch(string(content), -1) {
		// We only extract imports now
		//print("Match: %v\n", match)
		imp := match[2]
		info.Imports = append(info.Imports, imp)
	}
	sort.Strings(info.Imports)

	return info
}

// TODO(achew22): be a proper typescript parser based on a LL parser. It will
// be much faster than the regexp and more correct.
func buildTypeScriptRegexp() *regexp.Regexp {
	typescriptReSrc := `import ([a-zA-Z][a-zA-Z0-9]*) from ["'](.*)["']`
	return regexp.MustCompile(typescriptReSrc)
}
