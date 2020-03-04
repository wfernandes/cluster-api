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

package repository

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/internal/test"
)

// Client is used to interact with provider repositories.
// Provider repository are expected to contain two types of YAML files:
// - YAML files defining the provider components (CRD, Controller, RBAC etc.)
// - YAML files defining the cluster templates (Cluster, Machines)
type Client interface {
	config.Provider

	// GetVersion return the list of versions that are available in a provider repository
	GetVersions() ([]string, error)

	// Components provide access to YAML file for creating provider components.
	Components() ComponentsClient

	// Templates provide access to YAML file for generating workload cluster templates.
	// Please note that templates are expected to exist for the infrastructure providers only.
	Templates(version string) TemplateClient

	// Metadata provide access to YAML with the provider's metadata.
	Metadata(version string) MetadataClient
}

// repositoryClient implements Client.
type repositoryClient struct {
	config.Provider
	configVariablesClient config.VariablesClient
	repository            Repository
}

// ensure repositoryClient implements Client.
var _ Client = &repositoryClient{}

func (c *repositoryClient) GetVersions() ([]string, error) {
	return c.repository.GetVersions()
}

func (c *repositoryClient) Components() ComponentsClient {
	return newComponentsClient(c.Provider, c.repository, c.configVariablesClient)
}

func (c *repositoryClient) Templates(version string) TemplateClient {
	return newTemplateClient(c.Provider, version, c.repository, c.configVariablesClient)
}

func (c *repositoryClient) Metadata(version string) MetadataClient {
	return newMetadataClient(c.Provider, version, c.repository)
}

// Option is a configuration option supplied to New
type Option func(*repositoryClient)

// InjectRepository allows to override the repository implementation to use;
// by default, the repository implementation to use is created according to the
// repository URL.
func InjectRepository(repository Repository) Option {
	return func(c *repositoryClient) {
		c.repository = repository
	}
}

// New returns a Client.
func New(provider config.Provider, configVariablesClient config.VariablesClient, options ...Option) (Client, error) {
	return newRepositoryClient(provider, configVariablesClient, options...)
}

func newRepositoryClient(provider config.Provider, configVariablesClient config.VariablesClient, options ...Option) (*repositoryClient, error) {
	client := &repositoryClient{
		Provider:              provider,
		configVariablesClient: configVariablesClient,
	}
	for _, o := range options {
		o(client)
	}

	// if there is an injected repository, use it, otherwise use a default one
	if client.repository == nil {
		r, err := repositoryFactory(provider, configVariablesClient)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get repository client for the %s with name %s", provider.Type(), provider.Name())
		}
		client.repository = r
	}

	return client, nil
}

// Repository defines the behavior of a repository implementation.
// clusterctl is designed to support different repository types; each repository implementation should be aware of
// the provider version they are hosting, and possibly to host more than one version.
type Repository interface {
	// DefaultVersion returns the default provider version returned by a repository.
	// In case the repository URL points to latest, this method returns the current latest version; in other cases
	// it returns the version of the provider hosted in the repository.
	DefaultVersion() string

	// RootPath returns the path inside the repository where the YAML file for creating provider components and
	// the YAML file for generating workload cluster templates are stored.
	// This value is derived from the repository URL; all the paths returned by this interface should be relative to this path.
	RootPath() string

	// ComponentsPath return the path (a folder name or file name) of the YAML file for creating provider components.
	// This value is derived from the repository URL.
	ComponentsPath() string

	// GetFile return a file for a given provider version.
	GetFile(version string, path string) ([]byte, error)

	// GetVersion return the list of versions that are available in a provider repository
	GetVersions() ([]string, error)
}

var _ Repository = &test.FakeRepository{}

//repositoryFactory returns the repository implementation corresponding to the provider URL.
func repositoryFactory(providerConfig config.Provider, configVariablesClient config.VariablesClient) (Repository, error) {
	// parse the repository url
	rURL, err := url.Parse(providerConfig.URL())
	if err != nil {
		return nil, errors.Errorf("failed to parse repository url %q", providerConfig.URL())
	}

	// if the url is a github repository
	if rURL.Scheme == httpsScheme && rURL.Host == githubDomain {
		repo, err := newGitHubRepository(providerConfig, configVariablesClient)
		if err != nil {
			return nil, errors.Wrap(err, "error creating the GitHub repository client")
		}
		return repo, nil
	}

	// if the url is a local filesystem repository
	if rURL.Scheme == "file" || rURL.Scheme == "" {
		repo, err := newLocalRepository(providerConfig, configVariablesClient)
		if err != nil {
			return nil, errors.Wrap(err, "error creating the local filesystem repository client")
		}
		return repo, nil
	}

	return nil, errors.Errorf("invalid provider url. there are no provider implementation for %q schema", rURL.Scheme)
}

const overrideFolder = "overrides"

// getLocalOverride return local override file from the config folder, if it exists.
// This is required for development purposes, but it can be used also in production as a workaround for problems on the official repositories
func getLocalOverride(provider config.Provider, version, path string) ([]byte, error) {
	// local override files are searched at $home/.cluster-api/overrides/<provider-label>/<version>/<path>
	homeFolder := filepath.Join(homedir.HomeDir(), config.ConfigFolder)
	overridePath := filepath.Join(homeFolder, overrideFolder, provider.ManifestLabel(), version, path)

	// it the local override exists, use it
	_, err := os.Stat(overridePath)
	if err == nil {
		content, err := ioutil.ReadFile(overridePath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read local override for %s/%s/%s", provider.ManifestLabel(), version, path)
		}
		return content, nil
	}

	// it the local override does not exists, return (so files from the provider's repository could be used)
	if os.IsNotExist(err) {
		return nil, nil
	}

	// blocks for any other error
	return nil, err
}
