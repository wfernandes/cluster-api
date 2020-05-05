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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/internal/util"
)

// Template wraps a YAML file that defines the cluster objects (Cluster, Machines etc.).
// It is important to notice that clusterctl applies a set of processing steps to the “raw” cluster template YAML read
// from the provider repositories:
// 1. Checks for all the variables in the cluster template YAML file and replace with corresponding config values
// 2. Ensure all the cluster objects are deployed in the target namespace
type Template interface {
	// Variables required by the template.
	// This value is derived by the template YAML.
	Variables() []string

	// TargetNamespace where the template objects will be installed.
	TargetNamespace() string

	// Yaml returns yaml defining all the cluster template objects as a byte array.
	Yaml() ([]byte, error)

	// Objs returns the cluster template as a list of Unstructured objects.
	Objs() []unstructured.Unstructured
}

// template implements Template.
type template struct {
	// TODO (wfernandes) should this be a map[string][interface{}]. That way
	// we can store the values as well.
	variables       []string
	targetNamespace string
	objs            []unstructured.Unstructured
	// TODO (wfernandes): should this be a set of files or should we expect the fetchers to put
	// it all in a single yaml?
	rawYaml []byte
}

// Ensures template implements the Template interface.
var _ Template = &template{}

func (t *template) Variables() []string {
	return t.variables
}

func (t *template) TargetNamespace() string {
	return t.targetNamespace
}

func (t *template) Objs() []unstructured.Unstructured {
	return t.objs
}

func (t *template) Yaml() ([]byte, error) {
	return util.FromUnstructured(t.objs)
}

func (t *template) RawYaml() []byte {
	return t.rawYaml
}

// NewTemplate returns a new objects embedding a cluster template YAML file.
func NewTemplate(_ []byte, _ config.VariablesClient, targetNamespace string, _ bool) (*template, error) {
	// 	// Inspect variables and replace with values from the configuration.
	// 	variables := inspectVariables(rawYaml)
	// 	if listVariablesOnly {
	// 		return &template{
	// 			variables:       variables,
	// 			targetNamespace: targetNamespace,
	// 		}, nil
	// 	}

	// 	yaml, err := replaceVariables(rawYaml, variables, configVariablesClient, false)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "failed to perform variable substitution")
	// 	}

	// 	// Transform the yaml in a list of objects, so following transformation can work on typed objects (instead of working on a string/slice of bytes).
	// 	objs, err := util.ToUnstructured(yaml)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "failed to parse yaml")
	// 	}

	// 	// Ensures all the template components are deployed in the target namespace (applies only to namespaced objects)
	// 	// This is required in order to ensure a cluster and all the related objects are in a single namespace, that is a requirement for
	// 	// the clusterctl move operation (and also for many controller reconciliation loops).
	// 	objs = fixTargetNamespace(objs, targetNamespace)

	// 	return &template{
	// 		variables:       variables,
	// 		targetNamespace: targetNamespace,
	// 		objs:            objs,
	// 	}, nil
	return nil, nil
}
