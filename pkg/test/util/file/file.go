// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"istio.io/istio/pkg/test"
)

// AsBytes is a simple wrapper around os.ReadFile provided for completeness.
func AsBytes(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// AsBytesOrFail calls AsBytes and fails the test if any errors occurred.
func AsBytesOrFail(t test.Failer, filename string) []byte {
	t.Helper()
	content, err := AsBytes(filename)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

// MustAsBytes calls AsBytes and panics the test if any errors occurred.
func MustAsBytes(filename string) []byte {
	content, err := AsBytes(filename)
	if err != nil {
		panic(err)
	}
	return content
}

// AsStringArray is a convenience wrapper around os.ReadFile that converts the content to a string.
func AsStringArray(files ...string) ([]string, error) {
	out := make([]string, 0, len(files))
	for _, f := range files {
		b, err := AsBytes(f)
		if err != nil {
			return nil, err
		}
		out = append(out, string(b))
	}
	return out, nil
}

// AsStringArrayOrFail calls AsStringOrFail and then converts to string.
func AsStringArrayOrFail(t test.Failer, files ...string) []string {
	t.Helper()
	out, err := AsStringArray(files...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// AsString is a convenience wrapper around os.ReadFile that converts the content to a string.
func AsString(filename string) (string, error) {
	b, err := AsBytes(filename)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// AsStringOrFail calls AsBytesOrFail and then converts to string.
func AsStringOrFail(t test.Failer, filename string) string {
	t.Helper()
	return string(AsBytesOrFail(t, filename))
}

// MustAsString calls MustAsBytes and then converts to string.
func MustAsString(filename string) string {
	return string(MustAsBytes(filename))
}

// NormalizePath expands the homedir (~) and returns an error if the file doesn't exist.
func NormalizePath(originalPath string) (string, error) {
	if originalPath == "" {
		return "", nil
	}
	// trim leading/trailing spaces from the path and if it uses the homedir ~, expand it.
	out, err := expandHome(strings.TrimSpace(originalPath))
	if err != nil {
		return "", err
	}

	// Verify that the file exists.
	if _, err := os.Stat(out); os.IsNotExist(err) {
		return "", fmt.Errorf("failed normalizing file %s: %v", originalPath, err)
	}

	return out, nil
}

func expandHome(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}

	if path[0] != '~' {
		return path, nil
	}

	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return "", errors.New("cannot expand user-specific home dir")
	}

	dir, f := os.LookupEnv("HOME")
	if !f {
		return "", errors.New("HOME is not set")
	}

	return filepath.Join(dir, path[1:]), nil
}

// ReadTarFile reads a tar compress file from the embedded
func ReadTarFile(filePath string) (string, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	tr := tar.NewReader(bytes.NewBuffer(b))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return "", err
		}
		if hdr.Name != strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)) {
			continue
		}
		contents, err := io.ReadAll(tr)
		if err != nil {
			return "", err
		}
		return string(contents), nil
	}
	return "", fmt.Errorf("file not found %v", filePath)
}

// ReadDir returns the names of all files in the given directory. This is not recursive.
// The base path is appended; for example, ReadDir("dir") -> ["dir/file1", "dir/folder1"]
func ReadDir(filePath string, extensions ...string) ([]string, error) {
	dir, err := os.ReadDir(filePath)
	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, d := range dir {
		matched := len(extensions) == 0 // If none are set, match anything
		for _, ext := range extensions {
			if filepath.Ext(d.Name()) == ext {
				matched = true
				break
			}
		}
		if matched {
			res = append(res, filepath.Join(filePath, d.Name()))
		}
	}
	return res, nil
}

func ReadDirOrFail(t test.Failer, filePath string, extensions ...string) []string {
	t.Helper()
	res, err := ReadDir(filePath, extensions...)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func WriteOrFail(t test.Failer, filePath string, contents []byte) {
	t.Helper()
	err := os.WriteFile(filePath, contents, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
}
