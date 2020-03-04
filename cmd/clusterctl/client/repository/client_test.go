/*
Copyright 2020 The Kubernetes Authors.

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

package repository

import (
	"os"
	"testing"

	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/internal/test"
)

func Test_newRepositoryClient_LocalFileSystemRepository(t *testing.T) {
	tmpDir := createTempDir(t)
	defer os.RemoveAll(tmpDir)

	dst1 := createLocalTestProviderFile(t, tmpDir, "bootstrap-foo/v1.0.0/bootstrap-components.yaml", "")
	dst2 := createLocalTestProviderFile(t, tmpDir, "bootstrap-bar/v2.0.0/bootstrap-components.yaml", "")

	type fields struct {
		provider config.Provider
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "successfully creates repository client with local filesystem backend and scheme == \"\"",
			fields: fields{
				provider: config.NewProvider("foo", dst1, clusterctlv1.BootstrapProviderType),
			},
		},
		{
			name: "successfully creates repository client with local filesystem backend and scheme == \"file\"",
			fields: fields{
				provider: config.NewProvider("bar", "file://"+dst2, clusterctlv1.BootstrapProviderType),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoClient, err := newRepositoryClient(tt.fields.provider, test.NewFakeVariableClient())
			if err != nil {
				t.Fatalf("got error %v when none was expected", err)
			}
			if _, ok := repoClient.repository.(*localRepository); !ok {
				t.Fatalf("got repository of type %T when *repository.localRepository was expected", repoClient.repository)
			}
		})
	}
}

func Test_newRepositoryClient_GithubRepository(t *testing.T) {
	githubProvider := "https://github.com/myorg/myrepo/releases/v1.0.0/bootstrap-components.yaml"
	type fields struct {
		provider config.Provider
	}
	tests := []struct {
		name   string
		fields fields
		err    bool
	}{
		{
			name: "successfully creates repository client with github backend and scheme == \"https\"",
			fields: fields{
				provider: config.NewProvider("foo", githubProvider, clusterctlv1.BootstrapProviderType),
			},
		},
		{
			name: "returns error if provider url is not properly formatted github url",
			fields: fields{
				provider: config.NewProvider("foo", "https://github.com/myorg/myrepo", clusterctlv1.BootstrapProviderType),
			},
			err: true,
		},
		{
			name: "returns error if provider url is not a github or localfilesystem",
			fields: fields{
				provider: config.NewProvider("foo", "https://gitlab.com/myorg/myrepo", clusterctlv1.BootstrapProviderType),
			},
			err: true,
		},
		{
			name: "returns error if provider url doesn't have correct scheme",
			fields: fields{
				provider: config.NewProvider("foo", "http://github.com/myorg/myrepo", clusterctlv1.BootstrapProviderType),
			},
			err: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoClient, err := newRepositoryClient(tt.fields.provider, test.NewFakeVariableClient())
			if err != nil && !tt.err {
				t.Fatalf("got error %v when none was expected", err)
			}
			if err == nil && tt.err {
				t.Fatal("expected error, but received none")
			}
			if repoClient != nil {
				if _, ok := repoClient.repository.(*gitHubRepository); !ok {
					t.Fatalf("got repository of type %T when *repository.gitHubRepository was expected", repoClient.repository)
				}
			}
		})
	}

}
