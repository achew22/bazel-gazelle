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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestTypeScriptRegexpGroupNames(t *testing.T) {
	names := typescriptRe.SubexpNames()
	nameMap := map[string]int{
		"import":  importSubexpIndex,
		"package": packageSubexpIndex,
		"optkey":  optkeySubexpIndex,
		"optval":  optvalSubexpIndex,
		"service": serviceSubexpIndex,
	}
	for name, index := range nameMap {
		if names[index] != name {
			t.Errorf("typescript regexp subexp %d is %s ; want %s", index, names[index], name)
		}
	}
	if len(names)-1 != len(nameMap) {
		t.Errorf("typescript regexp has %d groups ; want %d", len(names), len(nameMap))
	}
}

func TestTypeScriptFileInfo(t *testing.T) {
	for _, tc := range []struct {
		desc, name, typescript string
		want              FileInfo
	}{
		{
			desc:  "empty",
			name:  "empty^file.ts",
			typescript: "",
			want:  FileInfo{},
		}, {
			desc:  "simple package",
			name:  "package.ts",
			typescript: "package foo;",
			want: FileInfo{
				PackageName: "foo",
			},
		}, {
			desc:  "full package",
			name:  "full.ts",
			typescript: "package foo.bar.baz;",
			want: FileInfo{
				PackageName: "foo.bar.baz",
			},
		}, {
			desc: "import simple",
			name: "imp.ts",
			typescript: `import 'single.ts';
import "double.ts";`,
			want: FileInfo{
				Imports: []string{"double.ts", "single.ts"},
			},
		}, {
			desc: "import quote",
			name: "quote.ts",
			typescript: `import '""\".ts"';
import "'.ts";`,
			want: FileInfo{
				Imports: []string{"\"\"\".ts\"", "'.ts"},
			},
		}, {
			desc:  "import escape",
			name:  "escape.ts",
			typescript: `import '\n\012\x0a.ts';`,
			want: FileInfo{
				Imports: []string{"\n\n\n.ts"},
			},
		}, {
			desc: "import two",
			name: "two.ts",
			typescript: `import "first.ts";
import "second.ts";`,
			want: FileInfo{
				Imports: []string{"first.ts", "second.ts"},
			},
		}, {
			desc:  "go_package",
			name:  "gopkg.ts",
			typescript: `option go_package = "github.com/example/project;projectpb";`,
			want: FileInfo{
				Options: []Option{{Key: "go_package", Value: "github.com/example/project;projectpb"}},
			},
		}, {
			desc:  "service",
			name:  "service.ts",
			typescript: `service ChatService {}`,
			want: FileInfo{
				HasServices: true,
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			dir, err := ioutil.TempDir(os.Getenv("TEST_TEMPDIR"), "TestTypeScriptFileinfo")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(dir)
			if err := ioutil.WriteFile(filepath.Join(dir, tc.name), []byte(tc.ts), 0600); err != nil {
				t.Fatal(err)
			}

			got := typescriptFileInfo(dir, tc.name)

			// Clear fields we don't care about for testing.
			got = FileInfo{
				PackageName: got.PackageName,
				Imports:     got.Imports,
				Options:     got.Options,
				HasServices: got.HasServices,
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %#v; want %#v", got, tc.want)
			}
		})
	}
}
