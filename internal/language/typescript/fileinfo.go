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
	"bytes"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
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

	for _, match := range typescriptRe.FindAllSubmatch(content, -1) {
		// We only extract imports now
		imp := unquoteTypeScriptString(match[0])
		info.Imports = append(info.Imports, imp)
	}
	sort.Strings(info.Imports)

	return info
}

// Based on https://regex101.com/r/xA9kG3/122
// TODO(achew22): be a proper typescript parser based on a LL parser. It will
// be much faster than the regexp.
func buildTypeScriptRegexp() *regexp.Regexp {
	typescriptReSrc := `import(?:["'\s]*([\w*{}\n\r\t, ]+)from\s*)?["'\s].*([@\w/_-]+)["'\s].*;$`
	return regexp.MustCompile(typescriptReSrc)
}

func unquoteTypeScriptString(q []byte) string {
	// Adjust quotes so that Unquote is happy. We need a double quoted string
	// without unescaped double quote characters inside.
	noQuotes := bytes.Split(q[1:len(q)-1], []byte{'"'})
	if len(noQuotes) != 1 {
		for i := 0; i < len(noQuotes)-1; i++ {
			if len(noQuotes[i]) == 0 || noQuotes[i][len(noQuotes[i])-1] != '\\' {
				noQuotes[i] = append(noQuotes[i], '\\')
			}
		}
		q = append([]byte{'"'}, bytes.Join(noQuotes, []byte{'"'})...)
		q = append(q, '"')
	}
	if q[0] == '\'' {
		q[0] = '"'
		q[len(q)-1] = '"'
	}

	s, err := strconv.Unquote(string(q))
	if err != nil {
		log.Panicf("unquoting string literal %s from typescript: %v", q, err)
	}
	return s
}
