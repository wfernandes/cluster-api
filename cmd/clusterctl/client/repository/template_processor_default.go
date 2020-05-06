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
	"fmt"
)

type defaultYamlProcessor struct {
	skipVariables bool
}

func newDefaultYamlProcessor(skipVariables bool) *defaultYamlProcessor {
	return &defaultYamlProcessor{
		skipVariables: skipVariables,
	}
}

func (tp *defaultYamlProcessor) ArtifactName(version, flavor string) string {
	// building template name according with the naming convention
	name := "cluster-template"
	if flavor != "" {
		name = fmt.Sprintf("%s-%s", name, flavor)
	}
	name = fmt.Sprintf("%s.yaml", name)

	return name
}

func (tp *defaultYamlProcessor) GetVariables(rawArtifact []byte) ([]string, error) {
	return inspectVariables(rawArtifact), nil
}

func (tp *defaultYamlProcessor) Process(rawArtifact []byte, variablesGetter VariablesGetter) ([]byte, error) {
	variables := inspectVariables(rawArtifact)
	processedYaml, err := replaceVariables(
		rawArtifact,
		variables,
		variablesGetter,
		// NOTE: WHAT TO DO ABOUT THIS????
		tp.skipVariables,
	)
	if err != nil {
		return nil, err
	}
	return processedYaml, nil
}
