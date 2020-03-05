/*
Copyright 2019 The Kubernetes Authors.

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

package config

import (
	"path/filepath"
	"testing"

	"k8s.io/client-go/util/homedir"
)

func TestConfigClient(t *testing.T) {

	// reader := test.NewFakeReader()

	basepath := "/tmp/some-dir"
	t.Run("overrides path", func(t *testing.T) {

		tests := []struct {
			name                  string
			cfgPath               string
			expectedOverridesPath string
		}{
			{
				name:                  "from provided config file path",
				cfgPath:               filepath.Join(basepath, "config.yaml"),
				expectedOverridesPath: filepath.Join(basepath, overrideFolder),
			},
			{
				name:                  "from default config file path",
				cfgPath:               "",
				expectedOverridesPath: filepath.Join(homedir.HomeDir(), ConfigFolder, overrideFolder),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c, err := New(tt.cfgPath)
				if err != nil {
					t.Fatalf("expected no error, got %s", err)
				}
				if c.OverridesPath() != tt.expectedOverridesPath {
					t.Fatalf("expected %q, got %q", tt.expectedOverridesPath, c.OverridesPath())
				}
			})
		}
	})
}
