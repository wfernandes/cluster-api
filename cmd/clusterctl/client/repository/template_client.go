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
	"github.com/pkg/errors"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/internal/util"
	logf "sigs.k8s.io/cluster-api/cmd/clusterctl/log"
)

// TemplateOptions defines a set of well-know variables that all the cluster templates are expected to manage;
// this set of variables defines a simple, day1 experience that will be made accessible via flags in the clusterctl CLI.
// Please note that each provider/each template is allowed to add more variables, but additional variables are exposed
// only via environment variables or the clusterctl configuration file.
// type TemplateOptions struct {
// 	ClusterName       string
// 	Namespace         string
// 	KubernetesVersion string
// 	ControlplaneCount int
// 	WorkerCount       int
// }

// TemplateClient has methods to work with cluster templates hosted on a provider repository.
// Templates are yaml files to be used for creating a guest cluster.
type TemplateClient interface {
	Get(version, flavor, targetNamespace string) (Template, error)
}

type VariablesGetter interface {
	Get(key string) (string, error)
}

type YamlProcessor interface {
	ArtifactName(version, flavor string) string
	// GetVariables parses the template rawYaml and sets the variables
	GetVariables([]byte) ([]string, error)
	// ProcessYAML processes the template and returns the final yaml
	// ProcessYAML(*template) ([]byte, error)
	Process([]byte, VariablesGetter) ([]byte, error)
}

// templateClient implements TemplateClient.
type templateClient struct {
	listVariablesOnly     bool
	provider              config.Provider
	repository            Repository
	configVariablesClient config.VariablesClient
	processor             YamlProcessor
}

type TemplateClientInput struct {
	listVariablesOnly     bool
	provider              config.Provider
	repository            Repository
	configVariablesClient config.VariablesClient
}

// Ensure templateClient implements the TemplateClient interface.
var _ TemplateClient = &templateClient{}

// newTemplateClient returns the default templateClient.
func newTemplateClient(input TemplateClientInput) *templateClient {
	// this is where we can create or pass in the correct processor and
	// fetcher through options pattern.
	return &templateClient{
		listVariablesOnly:     input.listVariablesOnly,
		provider:              input.provider,
		repository:            input.repository,
		configVariablesClient: input.configVariablesClient,
		processor: newDefaultYamlProcessor(
			input.listVariablesOnly,
		),
	}
}

// Get return the template for the flavor specified.
// In case the template does not exists, an error is returned.
// TODO: (wfernandes) Fix this documentation.
// Get assumes the following naming convention for templates: cluster-template[-<flavor_name>].yaml
func (c *templateClient) Get(version, flavor, targetNamespace string) (Template, error) {
	log := logf.Log

	if targetNamespace == "" {
		return nil, errors.New("invalid arguments: please provide a targetNamespace")
	}

	name := c.processor.ArtifactName(version, flavor)

	// read the component YAML, reading the local override file if it exists, otherwise read from the provider repository
	rawArtifact, err := getLocalOverride(&newOverrideInput{
		configVariablesClient: c.configVariablesClient,
		provider:              c.provider,
		version:               version,
		filePath:              name,
	})
	if err != nil {
		return nil, err
	}

	if rawArtifact == nil {
		log.V(5).Info("Fetching", "File", name, "Provider", c.provider.ManifestLabel(), "Version", version)
		rawArtifact, err = c.repository.GetFile(version, name)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read %q from provider's repository %q", name, c.provider.ManifestLabel())
		}
	} else {
		log.V(1).Info("Using", "Override", name, "Provider", c.provider.ManifestLabel(), "Version", version)
	}

	// GetVariables should parse the template and set the variables on the
	// template object
	variables, err := c.processor.GetVariables(rawArtifact)
	if err != nil {
		return nil, err
	}
	// get variables for the template object
	if c.listVariablesOnly {
		return &template{
			variables: variables,
		}, nil
	}

	// process the template
	processedYaml, err := c.processor.Process(rawArtifact, c.configVariablesClient)
	if err != nil {
		return nil, err
	}
	// Transform the yaml in a list of objects, so following transformation can work on typed objects (instead of working on a string/slice of bytes).
	objs, err := util.ToUnstructured(processedYaml)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse yaml")
	}

	// Ensures all the template components are deployed in the target namespace (applies only to namespaced objects)
	// This is required in order to ensure a cluster and all the related objects are in a single namespace, that is a requirement for
	// the clusterctl move operation (and also for many controller reconciliation loops).
	objs = fixTargetNamespace(objs, targetNamespace)

	return &template{
		objs:            objs,
		targetNamespace: targetNamespace,
	}, nil
}
