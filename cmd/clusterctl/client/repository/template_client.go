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

type TemplateProcessor interface {
	// Fetch is repobsible for fetching the template artifacts required for
	// processing the Yaml.
	Fetch(version, flavor string) (*template, error)
	// GetVariables parses the template rawYaml and sets the variables
	GetVariables(*template) error
	// ProcessYAML processes the template and returns the final yaml
	ProcessYAML(*template) ([]byte, error)
}

// templateClient implements TemplateClient.
type templateClient struct {
	listVariablesOnly     bool
	provider              config.Provider
	repository            Repository
	configVariablesClient config.VariablesClient
	processor             TemplateProcessor
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
		processor: newDefaultTemplateProcessor(
			input.listVariablesOnly,
			input.configVariablesClient,
			input.provider,
			input.repository,
		),
	}
}

// Get return the template for the flavor specified.
// In case the template does not exists, an error is returned.
// TODO: (wfernandes) Fix this documentation.
// Get assumes the following naming convention for templates: cluster-template[-<flavor_name>].yaml
func (c *templateClient) Get(version, flavor, targetNamespace string) (Template, error) {
	// log := logf.Log

	if targetNamespace == "" {
		return nil, errors.New("invalid arguments: please provide a targetNamespace")
	}

	// we are always reading templateClient for a well know version, that usually is
	// the version of the provider installed in the management cluster.
	// NOTE: (wfernandes) Why are we doing this? Does version really need to
	// be part of templateClient or can it be just passed into Get?
	// version := c.version

	t, err := c.processor.Fetch(version, flavor)
	if err != nil {
		return nil, err
	}

	// return NewTemplate(c.processor) // rawYaml, c.configVariablesClient, targetNamespace, listVariablesOnly)

	// this template object would be returned by the TemplateFetcher
	// NOTE: (wfernandes) this is now holding the reference to the rawYaml.
	// Check if there is any difference from previous implementation. Is this
	// efficient?
	// t := &template{
	// 	rawYaml:         rawYaml,
	// 	targetNamespace: targetNamespace,
	// }

	// GetVariables should parse the template and set the variables on the
	// template object
	err = c.processor.GetVariables(t)
	if err != nil {
		return nil, err
	}
	// get variables for the template object
	if c.listVariablesOnly {
		return t, nil
	}

	// process the template
	processedYaml, err := c.processor.ProcessYAML(t)
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
	t.objs = objs
	t.targetNamespace = targetNamespace

	return t, nil
}
